package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
)

type E2ESuite struct {
	suite.Suite
	ctx          context.Context
	pgContainer  *postgres.PostgresContainer
	rmqContainer *rabbitmq.RabbitMQContainer
	dbURL        string
	rabbitURL    string
	binaries     map[string]string
	sourceDirs   map[string]string // NEW: Map service name to source dir
	processes    map[string]*exec.Cmd
}

func TestE2ESuite(t *testing.T) {
	suite.Run(t, new(E2ESuite))
}

func (s *E2ESuite) SetupSuite() {
	s.ctx = context.Background()
	s.binaries = make(map[string]string)
	s.sourceDirs = make(map[string]string)
	s.processes = make(map[string]*exec.Cmd)

	// 1. Start RabbitMQ
	var err error
	s.rmqContainer, err = rabbitmq.Run(s.ctx,
		"rabbitmq:3.12-management-alpine",
		rabbitmq.WithAdminPassword("guest"),
	)
	s.Require().NoError(err)
	s.rabbitURL, err = s.rmqContainer.AmqpURL(s.ctx)
	s.Require().NoError(err)

	// 2. Start Postgres
	s.pgContainer, err = postgres.Run(s.ctx,
		"postgres:16", // Switched from alpine to debian to avoid exec format errors in some envs
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	s.Require().NoError(err)
	s.dbURL, err = s.pgContainer.ConnectionString(s.ctx, "sslmode=disable")
	s.Require().NoError(err)

	// 3. Setup Databases
	s.setupDatabases()

	// 4. Build Services
	s.buildService("shopping-cart", "../../services/shopping-cart/cmd/server", "../../services/shopping-cart")
	s.buildService("pocv", "../../services/pocv/cmd/server", "../../services/pocv")
	s.buildService("qualification", "../../services/qualification/cmd/server", "../../services/qualification")
	s.buildService("mock-inventory", "../../services/qualification/cmd/mock-inventory", "")
	s.buildService("mock-gis", "../../services/qualification/cmd/mock-gis", "")
	s.buildService("mock-billing", "../../services/pocv/cmd/mock-billing", "")

	// 5. Start Services
	s.startService("mock-inventory", "", "Mock Inventory Started")
	s.startService("mock-gis", "", "Mock GIS Started")
	s.startService("mock-billing", "", "Mock Billing Started")
	s.startService("qualification", "", "Qualification Service Ready")
	s.startService("shopping-cart", "cart", "Shopping Cart Service Started")
	// Wait for "Starting" instead of "Started" to avoid buffering issues/hangs blocking the test
	s.startService("pocv", "pocv", "POCV Service Started")
}

func (s *E2ESuite) TearDownSuite() {
	for name, cmd := range s.processes {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		fmt.Printf("Stopped %s\n", name)
	}
	if s.pgContainer != nil {
		_ = s.pgContainer.Terminate(s.ctx)
	}
	if s.rmqContainer != nil {
		_ = s.rmqContainer.Terminate(s.ctx)
	}
}

func (s *E2ESuite) setupDatabases() {
	conn, err := pgx.Connect(s.ctx, s.dbURL)
	s.Require().NoError(err)
	defer func() { _ = conn.Close(s.ctx) }()

	dbs := []string{"cart", "pocv", "party"}
	for _, dbName := range dbs {
		_, _ = conn.Exec(s.ctx, fmt.Sprintf("CREATE DATABASE %s;", dbName))
	}
}

func (s *E2ESuite) buildService(name, path, srcDir string) {
	outPath := filepath.Join(os.TempDir(), name)
	cmd := exec.Command("go", "build", "-o", outPath, path)
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err, "Failed to build %s: %s", name, string(out))
	s.binaries[name] = outPath
	if srcDir != "" {
		abs, _ := filepath.Abs(srcDir)
		s.sourceDirs[name] = abs
	}
}

func (s *E2ESuite) startService(name, dbName, waitForLog string) {
	binPath := s.binaries[name]
	cmd := exec.Command(binPath)

	// Set working directory if we have one (for migrations)
	if dir, ok := s.sourceDirs[name]; ok {
		cmd.Dir = dir
	}
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("RABBITMQ_URL=%s", s.rabbitURL))

	if dbName != "" {
		host, _ := s.pgContainer.Host(s.ctx)
		port, _ := s.pgContainer.MappedPort(s.ctx, "5432")
		customDSN := fmt.Sprintf("postgres://postgres:postgres@%s:%s/%s?sslmode=disable", host, port.Port(), dbName)
		cmd.Env = append(cmd.Env, fmt.Sprintf("DB_URL=%s", customDSN))
		cmd.Env = append(cmd.Env, fmt.Sprintf("DATABASE_URL=%s", customDSN))
	}

	// Capture stdout for waiting
	stdout, err := cmd.StdoutPipe()
	s.Require().NoError(err)
	cmd.Stderr = os.Stderr // Pipe stderr directly

	err = cmd.Start()
	s.Require().NoError(err, "Failed to start %s", name)
	s.processes[name] = cmd

	// Wait for log
	ready := make(chan struct{})
	go func() {
		readyClosed := false
		// Simple scanner
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				chunk := string(buf[:n])
				fmt.Print(chunk) // Echo to test output
				if !readyClosed && contains(chunk, waitForLog) {
					close(ready)
					readyClosed = true
					// CONTINUE READING - don't return!
				}
			}
			if err != nil {
				return
			}
		}
	}()

	select {
	case <-ready:
		fmt.Printf("✅ %s is ready\n", name)
	case <-time.After(30 * time.Second):
		s.FailNow(fmt.Sprintf("Timeout waiting for %s to start", name))
	}
}

func contains(s, substr string) bool {
	// Simple check, import strings if needed, but keeping it minimal
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper to publish command
func (s *E2ESuite) publish(ch *amqp.Channel, exchange, routingKey, payload string) {
	err := ch.PublishWithContext(s.ctx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType:   "application/json",
		Body:          []byte(payload),
		CorrelationId: uuid.New().String(),
	})
	s.Require().NoError(err)
}

// Helper to snoop all messages
func (s *E2ESuite) startSnoop(ch *amqp.Channel, exchange string) {
	q, _ := ch.QueueDeclare("", false, true, true, false, nil)
	_ = ch.QueueBind(q.Name, "#", exchange, false, nil)
	msgs, _ := ch.Consume(q.Name, "", true, false, false, false, nil)
	go func() {
		for d := range msgs {
			fmt.Printf("[SNOOP %s] %s: %s\n", exchange, d.RoutingKey, string(d.Body))
		}
	}()
}

// ---------------- TESTS ----------------

func (s *E2ESuite) Test1_HappyPath_FiberOrder() {
	conn, _ := amqp.Dial(s.rabbitURL)
	defer func() { _ = conn.Close() }()
	ch, _ := conn.Channel()
	defer func() { _ = ch.Close() }()

	// PRE-DECLARE EXCHANGES to avoid 404 if services are slow
	_ = ch.ExchangeDeclare("ex.domain.market", "topic", true, false, false, false, nil)
	_ = ch.ExchangeDeclare("ex.domain.commerce", "topic", true, false, false, false, nil)
	_ = ch.ExchangeDeclare("ex.domain.order", "topic", true, false, false, false, nil)

	// Start Snoops
	s.startSnoop(ch, "ex.domain.market")
	s.startSnoop(ch, "ex.domain.commerce")
	s.startSnoop(ch, "ex.domain.order")

	// 1. Qualify
	qQual, _ := ch.QueueDeclare("", false, true, true, false, nil)
	_ = ch.QueueBind(qQual.Name, "evt.qual.checked", "ex.domain.market", false, nil)
	msgsQual, _ := ch.Consume(qQual.Name, "", true, false, false, false, nil)

	s.publish(ch, "ex.domain.market", "cmd.qual.eligibility.check",
		`{"address": {"street": "Main St", "city": "Berlin"}, "categoryFilter": ["Internet"]}`)

	select {
	case d := <-msgsQual:
		s.Contains(string(d.Body), "Qualified")
	case <-time.After(20 * time.Second):
		s.Fail("Timeout waiting for qualification")
	}

	// 2. Cart Add
	qCart, _ := ch.QueueDeclare("", false, true, true, false, nil)
	_ = ch.QueueBind(qCart.Name, "evt.cart.session.updated", "ex.domain.commerce", false, nil)
	msgsCart, _ := ch.Consume(qCart.Name, "", true, false, false, false, nil)

	cartID := uuid.New().String()
	offeringID := uuid.New().String() // Use UUID for Postgres compatibility
	s.publish(ch, "ex.domain.commerce", "cmd.cart.item.add",
		fmt.Sprintf(`{"cartId": "%s", "offeringId": "%s", "quantity": 1}`, cartID, offeringID))

	select {
	case <-msgsCart:
		// Cart updated
	case <-time.After(20 * time.Second):
		s.Fail("Timeout waiting for cart")
	}

	// 3. Checkout
	qOrder, _ := ch.QueueDeclare("", false, true, true, false, nil)
	_ = ch.QueueBind(qOrder.Name, "cmd.order.management.create", "ex.domain.order", false, nil)
	msgsOrder, _ := ch.Consume(qOrder.Name, "", true, false, false, false, nil)

	s.publish(ch, "ex.domain.order", "cmd.order.checkout.submit", fmt.Sprintf(`{"cartId": "%s"}`, cartID))

	// Wait for Final Command (Success)
	// Note: POCV emits cmd.order.management.create.
	// The test needs to listen to this to verify success.
	// We bind qOrder to it.
	select {
	case d := <-msgsOrder:
		// Success!
		s.Contains(string(d.Body), cartID)
	case <-time.After(30 * time.Second):
		s.Fail("Timeout waiting for order creation")
	}
}

func (s *E2ESuite) Test2_UnhappyPath_QualificationFailed() {
	conn, _ := amqp.Dial(s.rabbitURL)
	defer func() { _ = conn.Close() }()
	ch, _ := conn.Channel()
	defer func() { _ = ch.Close() }()

	qQual, _ := ch.QueueDeclare("", false, true, true, false, nil)
	ch.QueueBind(qQual.Name, "evt.qual.checked", "ex.domain.market", false, nil)
	msgsQual, _ := ch.Consume(qQual.Name, "", true, false, false, false, nil)

	// City = "Nowhere" triggers mock failure
	s.publish(ch, "ex.domain.market", "cmd.qual.eligibility.check",
		`{"address": {"street": "Main St", "city": "Nowhere"}, "categoryFilter": ["Internet"]}`)

	select {
	case d := <-msgsQual:
		s.Contains(string(d.Body), "Unqualified")
	case <-time.After(20 * time.Second):
		s.Fail("Timeout waiting for qualification failure")
	}
}

func (s *E2ESuite) Test3_UnhappyPath_PaymentDeclined() {
	conn, _ := amqp.Dial(s.rabbitURL)
	defer func() { _ = conn.Close() }()
	ch, _ := conn.Channel()
	defer func() { _ = ch.Close() }()

	// 1. Cart with Expensive Item
	cartID := uuid.New().String()

	// Seed Catalog
	_ = ch.ExchangeDeclare("ex.domain.catalog", "topic", true, false, false, false, nil)
	offeringID := uuid.New().String()
	s.publish(ch, "ex.domain.catalog", "evt.catalog.offering.created",
		fmt.Sprintf(`{"id": "%s", "price": {"amount": 2000, "currency": "EUR"}}`, offeringID))
	time.Sleep(100 * time.Millisecond)

	s.publish(ch, "ex.domain.commerce", "cmd.cart.item.add",
		fmt.Sprintf(`{"cartId": "%s", "offeringId": "%s", "quantity": 10000}`, cartID, offeringID))

	// Wait for cart update to ensure it's processed (optional but safer)
	time.Sleep(1 * time.Second)

	// 2. Monitor Compensation Event
	qComp, _ := ch.QueueDeclare("", false, true, true, false, nil)
	ch.QueueBind(qComp.Name, "cmd.inventory.resource.release", "ex.domain.order", false, nil)
	msgsComp, _ := ch.Consume(qComp.Name, "", true, false, false, false, nil)

	// 3. Checkout
	s.publish(ch, "ex.domain.order", "cmd.order.checkout.submit", fmt.Sprintf(`{"cartId": "%s"}`, cartID))

	select {
	case d := <-msgsComp:
		// Success! Compensation triggered
		s.Contains(string(d.Body), "PaymentDeclined")
	case <-time.After(30 * time.Second):
		s.Fail("Timeout waiting for compensation")
	}
}
