package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"tmf/pkg/rabbitmq"
	"tmf/services/customer-management/internal/domain"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Handlers manages command and query handling
// Handlers manages command and query handling
type Handlers struct {
	repo           domain.Repository
	publisher      rabbitmq.Publisher
	tm             domain.TransactionManager
	eventPublisher domain.EventPublisher
}

func NewHandlers(repo domain.Repository, publisher rabbitmq.Publisher, tm domain.TransactionManager, eventPublisher domain.EventPublisher) *Handlers {
	return &Handlers{
		repo:           repo,
		publisher:      publisher,
		tm:             tm,
		eventPublisher: eventPublisher,
	}
}

// Payloads

type OnboardCustomerPayload struct {
	ID              string               `json:"id"`
	Name            string               `json:"name"`
	PartyID         string               `json:"partyId"`
	PartyType       string               `json:"partyType"`
	Accounts        []CustomerAccountDTO `json:"accounts"`
	CreditProfiles  []CreditProfileDTO   `json:"creditProfiles"`
	ContactMediums  []ContactMediumDTO   `json:"contactMediums"`
	Characteristics []CharacteristicDTO  `json:"characteristics"`
	PrivacyConsents []PrivacyConsentDTO  `json:"privacyConsents"`
	RelatedParties  []RelatedPartyDTO    `json:"relatedParties"`
	PaymentMethods  []PaymentMethodDTO   `json:"paymentMethods"`
	MarketSegments  []MarketSegmentDTO   `json:"marketSegments"`
}

type CustomerAccountDTO struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	AccountStatus string `json:"accountStatus"`
	AccountType   string `json:"accountType"`
	BillFormat    string `json:"billFormat"`
	BillingCycle  string `json:"billingCycle"`
}

type CreditProfileDTO struct {
	ID              string `json:"id"`
	CreditRiskScore int    `json:"creditRiskScore"`
	CreditScore     int    `json:"creditScore"`
}

type ContactMediumDTO struct {
	ID              string `json:"id"`
	MediumType      string `json:"mediumType"`
	Preferred       bool   `json:"preferred"`
	Value           string `json:"value"`
	Street1         string `json:"street1"`
	Street2         string `json:"street2"`
	City            string `json:"city"`
	StateOrProvince string `json:"stateOrProvince"`
	Postcode        string `json:"postcode"`
	Country         string `json:"country"`
}

type CharacteristicDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	ValueType string `json:"valueType"`
}

type PrivacyConsentDTO struct {
	ID            string `json:"id"`
	ConsentType   string `json:"consentType"`
	Status        string `json:"status"`
	ValidForStart string `json:"validForStart"`
}

type RelatedPartyDTO struct {
	ID             string `json:"id"`
	RelatedPartyID string `json:"relatedPartyId"`
	Role           string `json:"role"`
	Name           string `json:"name"`
	ValidForStart  string `json:"validForStart"`
	ValidForEnd    string `json:"validForEnd"`
}

type PaymentMethodDTO struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	Token         string `json:"token"`
	Details       string `json:"details"` // JSON string
	IsDefault     bool   `json:"isDefault"`
	ValidForStart string `json:"validForStart"`
	ValidForEnd   string `json:"validForEnd"`
}

type MarketSegmentDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type UpdateCustomerPayload struct {
	ID              string                `json:"id"`
	Status          domain.CustomerStatus `json:"status"`
	Name            string                `json:"name"`
	PartyID         string                `json:"partyId"`
	ContactMediums  []ContactMediumDTO    `json:"contactMediums"`
	Characteristics []CharacteristicDTO   `json:"characteristics"`
	PrivacyConsents []PrivacyConsentDTO   `json:"privacyConsents"`
	Accounts        []CustomerAccountDTO  `json:"accounts"`
	CreditProfiles  []CreditProfileDTO    `json:"creditProfiles"`
	RelatedParties  []RelatedPartyDTO     `json:"relatedParties"`
	PaymentMethods  []PaymentMethodDTO    `json:"paymentMethods"`
	MarketSegments  []MarketSegmentDTO    `json:"marketSegments"`
}

type GetCustomerPayload struct {
	ID string `json:"id"`
}

type SearchCustomerPayload struct {
	ID      string `json:"id"`
	Search  string `json:"search"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	PartyID string `json:"partyId"`
}

type DeleteCustomerPayload struct {
	ID string `json:"id"`
}

type LogInteractionPayload struct {
	ID              string `json:"id"`
	CustomerID      string `json:"customerId"`
	InteractionDate string `json:"interactionDate"`
	Channel         string `json:"channel"`
	Type            string `json:"type"`
	Description     string `json:"description"`
	AgentID         string `json:"agentId"`
}

// Handlers

func (h *Handlers) HandleOnboardCustomer(ctx context.Context, d amqp.Delivery) error {
	ctx = h.extractUser(ctx, d)
	var payload OnboardCustomerPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if payload.ID == "" {
		payload.ID = uuid.New().String()
	}

	customer := &domain.Customer{
		ID:        payload.ID,
		Name:      payload.Name,
		Status:    domain.CustomerStatusActive,
		PartyID:   payload.PartyID,
		PartyType: payload.PartyType,
	}

	for _, acc := range payload.Accounts {
		customer.Accounts = append(customer.Accounts, domain.CustomerAccount{
			ID:            acc.ID,
			Name:          acc.Name,
			AccountStatus: acc.AccountStatus,
			AccountType:   acc.AccountType,
			BillFormat:    acc.BillFormat,
			BillingCycle:  acc.BillingCycle,
		})
	}

	for _, cp := range payload.CreditProfiles {
		customer.CreditProfiles = append(customer.CreditProfiles, domain.CreditProfile{
			ID:              cp.ID,
			CreditRiskScore: cp.CreditRiskScore,
			CreditScore:     cp.CreditScore,
		})
	}

	for _, cm := range payload.ContactMediums {
		customer.ContactMediums = append(customer.ContactMediums, domain.ContactMedium{
			ID:              cm.ID,
			MediumType:      cm.MediumType,
			Preferred:       cm.Preferred,
			Value:           cm.Value,
			Street1:         cm.Street1,
			Street2:         cm.Street2,
			City:            cm.City,
			StateOrProvince: cm.StateOrProvince,
			Postcode:        cm.Postcode,
			Country:         cm.Country,
		})
	}

	for _, ch := range payload.Characteristics {
		customer.Characteristics = append(customer.Characteristics, domain.CustomerCharacteristic{
			ID:        ch.ID,
			Name:      ch.Name,
			Value:     ch.Value,
			ValueType: ch.ValueType,
		})
	}

	for _, p := range payload.PrivacyConsents {
		customer.PrivacyConsents = append(customer.PrivacyConsents, h.mapPrivacyConsent(p, customer.ID))
	}

	for _, rp := range payload.RelatedParties {
		customer.RelatedParties = append(customer.RelatedParties, h.mapRelatedParty(rp, customer.ID))
	}

	for _, pm := range payload.PaymentMethods {
		customer.PaymentMethods = append(customer.PaymentMethods, h.mapPaymentMethod(pm, customer.ID))
	}

	for _, ms := range payload.MarketSegments {
		customer.MarketSegments = append(customer.MarketSegments, h.mapMarketSegment(ms, customer.ID))
	}

	// Transaction
	err := h.tm.RunInTransaction(ctx, func(ctx context.Context) error {
		if err := h.repo.CreateCustomer(ctx, customer); err != nil {
			return fmt.Errorf("failed to create customer: %w", err)
		}

		// Publish event
		if err := h.eventPublisher.Publish(ctx, EvtCustomerCreated, customer); err != nil {
			return fmt.Errorf("failed to publish event: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return h.replyTo(ctx, d, customer)
}

func (h *Handlers) HandleUpdateCustomer(ctx context.Context, d amqp.Delivery) error {
	ctx = h.extractUser(ctx, d)
	var payload UpdateCustomerPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	updates := make(map[string]interface{})
	if payload.Status != "" {
		updates["status"] = payload.Status
	}
	if payload.Name != "" {
		updates["name"] = payload.Name
	}
	if payload.PartyID != "" {
		updates["party_id"] = payload.PartyID
	}

	if len(payload.PrivacyConsents) > 0 {
		var privacy []domain.PrivacyConsent
		for _, p := range payload.PrivacyConsents {
			privacy = append(privacy, h.mapPrivacyConsent(p, payload.ID))
		}
		updates["privacy_consents"] = privacy
	}
	if len(payload.Accounts) > 0 {
		var accounts []domain.CustomerAccount
		for _, a := range payload.Accounts {
			accounts = append(accounts, domain.CustomerAccount{
				ID:            a.ID,
				CustomerID:    payload.ID,
				Name:          a.Name,
				AccountStatus: a.AccountStatus,
				AccountType:   a.AccountType,
			})
		}
		updates["accounts"] = accounts
	}
	if len(payload.CreditProfiles) > 0 {
		var profiles []domain.CreditProfile
		for _, p := range payload.CreditProfiles {
			profiles = append(profiles, domain.CreditProfile{
				ID:              p.ID,
				CustomerID:      payload.ID,
				CreditRiskScore: p.CreditRiskScore,
				CreditScore:     p.CreditScore,
			})
		}
		updates["credit_profiles"] = profiles
	}
	if len(payload.RelatedParties) > 0 {
		var related []domain.RelatedParty
		for _, rp := range payload.RelatedParties {
			related = append(related, h.mapRelatedParty(rp, payload.ID))
		}
		updates["related_parties"] = related
	}
	if len(payload.PaymentMethods) > 0 {
		var payment []domain.PaymentMethod
		for _, pm := range payload.PaymentMethods {
			payment = append(payment, h.mapPaymentMethod(pm, payload.ID))
		}
		updates["payment_methods"] = payment
	}
	if len(payload.MarketSegments) > 0 {
		var segments []domain.MarketSegment
		for _, ms := range payload.MarketSegments {
			segments = append(segments, h.mapMarketSegment(ms, payload.ID))
		}
		updates["market_segments"] = segments
	}

	if len(updates) == 0 {
		return fmt.Errorf("no valid fields to update provided")
	}

	err := h.tm.RunInTransaction(ctx, func(ctx context.Context) error {
		if err := h.repo.PatchCustomer(ctx, payload.ID, updates); err != nil {
			return fmt.Errorf("failed to update customer: %w", err)
		}

		// Fetch for event payload
		updatedCustomer, err := h.repo.GetCustomer(ctx, payload.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch updated customer: %w", err)
		}

		if err := h.eventPublisher.Publish(ctx, EvtCustomerUpdated, updatedCustomer); err != nil {
			return fmt.Errorf("failed to publish event: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return h.replyTo(ctx, d, map[string]string{"status": "updated"})
}

func (h *Handlers) HandleGetCustomer(ctx context.Context, d amqp.Delivery) error {
	ctx = h.extractUser(ctx, d)
	var payload GetCustomerPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	customer, err := h.repo.GetCustomer(ctx, payload.ID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	return h.replyTo(ctx, d, customer)
}

func (h *Handlers) HandleSearchCustomer(ctx context.Context, d amqp.Delivery) error {
	ctx = h.extractUser(ctx, d)
	var payload SearchCustomerPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	criteria := make(map[string]interface{})
	if payload.ID != "" {
		criteria["id"] = payload.ID
	}
	if payload.Search != "" {
		criteria["search"] = payload.Search
	}
	if payload.Name != "" {
		criteria["name"] = payload.Name
	}
	if payload.Status != "" {
		criteria["status"] = payload.Status
	}
	if payload.PartyID != "" {
		criteria["party_id"] = payload.PartyID
	}

	customers, err := h.repo.SearchCustomers(ctx, criteria)
	if err != nil {
		return fmt.Errorf("failed to search customers: %w", err)
	}

	if d.ReplyTo != "" {
		return h.replyTo(ctx, d, customers)
	}

	return nil
}

func (h *Handlers) HandleDeleteCustomer(ctx context.Context, d amqp.Delivery) error {
	ctx = h.extractUser(ctx, d)
	var payload DeleteCustomerPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	err := h.tm.RunInTransaction(ctx, func(ctx context.Context) error {
		if err := h.repo.DeleteCustomer(ctx, payload.ID); err != nil {
			return fmt.Errorf("failed to delete customer: %w", err)
		}

		// Publish event
		if err := h.eventPublisher.Publish(ctx, EvtCustomerDeleted, payload); err != nil {
			return fmt.Errorf("failed to publish delete event: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return h.replyTo(ctx, d, map[string]string{"status": "deleted"})
}

func (h *Handlers) HandleLogInteraction(ctx context.Context, d amqp.Delivery) error {
	ctx = h.extractUser(ctx, d)
	var payload LogInteractionPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if payload.ID == "" {
		payload.ID = uuid.New().String()
	}

	var interactionDate time.Time
	if payload.InteractionDate != "" {
		interactionDate, _ = time.Parse(time.RFC3339, payload.InteractionDate)
	} else {
		interactionDate = time.Now().UTC()
	}

	interaction := &domain.CustomerInteraction{
		ID:              payload.ID,
		CustomerID:      payload.CustomerID,
		InteractionDate: interactionDate,
		Channel:         payload.Channel,
		Type:            payload.Type,
		Description:     payload.Description,
		AgentID:         payload.AgentID,
	}

	if err := h.repo.AddInteraction(ctx, interaction); err != nil {
		return fmt.Errorf("failed to log interaction: %w", err)
	}

	return h.replyTo(ctx, d, map[string]string{"status": "logged", "id": interaction.ID})
}

// Party Event Handlers

type PartyEventPayload struct {
	ID          string `json:"id"`
	Type        string `json:"@type"`
	GivenName   string `json:"givenName,omitempty"`
	FamilyName  string `json:"familyName,omitempty"`
	TradingName string `json:"tradingName,omitempty"`
}

func (h *Handlers) HandlePartyEvent(ctx context.Context, d amqp.Delivery) error {
	// For events, we might have a user or use system user
	ctx = h.extractUser(ctx, d)
	if ctx.Value(domain.UserContextKey) == nil {
		ctx = context.WithValue(ctx, domain.UserContextKey, "system.customer-management")
	}

	var payload PartyEventPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal party event: %w", err)
	}

	switch d.RoutingKey {
	case EvtPartyUpdated:
		return h.handlePartyUpdated(ctx, payload)
	case EvtPartyDeleted:
		return h.handlePartyDeleted(ctx, payload)
	case EvtPartyDeletionInitiated:
		return h.handlePartyDeletionInitiated(ctx, payload)
	}

	return nil
}

func (h *Handlers) handlePartyUpdated(ctx context.Context, p PartyEventPayload) error {
	// Find all customers linked to this party
	customers, err := h.repo.SearchCustomers(ctx, map[string]interface{}{"party_id": p.ID})
	if err != nil {
		return err
	}

	for _, cust := range customers {
		newName := ""
		if p.Type == "Individual" {
			newName = fmt.Sprintf("%s %s", p.GivenName, p.FamilyName)
		} else {
			newName = p.TradingName
		}

		if cust.Name != newName {
			updates := map[string]interface{}{"name": newName}
			if err := h.repo.PatchCustomer(ctx, cust.ID, updates); err != nil {
				slog.Error("failed to sync party update to customer", "customer_id", cust.ID, "error", err)
			}
		}
	}
	return nil
}

func (h *Handlers) handlePartyDeleted(ctx context.Context, p PartyEventPayload) error {
	customers, err := h.repo.SearchCustomers(ctx, map[string]interface{}{"party_id": p.ID})
	if err != nil {
		return err
	}

	for _, cust := range customers {
		updates := map[string]interface{}{
			"status":        domain.CustomerStatusClosed,
			"status_reason": "Linked party was deleted",
		}
		if err := h.repo.PatchCustomer(ctx, cust.ID, updates); err != nil {
			slog.Error("failed to close customer on party deletion", "customer_id", cust.ID, "error", err)
		}
	}
	return nil
}

func (h *Handlers) handlePartyDeletionInitiated(ctx context.Context, p PartyEventPayload) error {
	if p.ID == "" {
		slog.Warn("received deletion initiated with empty ID")
		return nil
	}

	// Check for active customers
	customers, err := h.repo.SearchCustomers(ctx, map[string]interface{}{"party_id": p.ID})
	if err != nil {
		return fmt.Errorf("failed to search customers: %w", err)
	}

	hasActive := false
	for _, c := range customers {
		if c.Status != domain.CustomerStatusClosed && c.Status != domain.CustomerStatusSuspended {
			hasActive = true
			break
		}
	}

	partyCommandExchange := "tmf.party"
	cmdPayload := map[string]string{"id": p.ID}

	if hasActive {
		slog.Info("blocking party deletion: active customers found", "party_id", p.ID)
		if err := h.publisher.Publish(ctx, partyCommandExchange, CmdPartyCancelDeletion, cmdPayload); err != nil {
			return fmt.Errorf("failed to publish cancel command: %w", err)
		}
	} else {
		slog.Info("approving party deletion: no active customers", "party_id", p.ID)
		if err := h.publisher.Publish(ctx, partyCommandExchange, CmdPartyFinalizeDeletion, cmdPayload); err != nil {
			return fmt.Errorf("failed to publish finalize command: %w", err)
		}
	}

	return nil
}

// Helpers

func (h *Handlers) extractUser(ctx context.Context, d amqp.Delivery) context.Context {
	if user, ok := d.Headers["user"].(string); ok && user != "" {
		ctx = context.WithValue(ctx, domain.UserContextKey, user)
	}
	if auth, ok := d.Headers["Authorization"].(string); ok && auth != "" {
		type contextKey string
		const authKey contextKey = "authorization"
		ctx = context.WithValue(ctx, authKey, auth)
	}
	return ctx
}

func (h *Handlers) replyTo(ctx context.Context, d amqp.Delivery, payload interface{}) error {
	if d.ReplyTo == "" {
		return nil
	}

	if d.CorrelationId != "" {
		ctx = context.WithValue(ctx, rabbitmq.ContextKeyCorrelationID, d.CorrelationId)
	}

	return h.publisher.Publish(ctx, "", d.ReplyTo, payload)
}

func (h *Handlers) mapPrivacyConsent(dto PrivacyConsentDTO, customerID string) domain.PrivacyConsent {
	var start time.Time
	if dto.ValidForStart != "" {
		start, _ = time.Parse(time.RFC3339, dto.ValidForStart)
	}

	return domain.PrivacyConsent{
		ID:            dto.ID,
		CustomerID:    customerID,
		ConsentType:   dto.ConsentType,
		Status:        dto.Status,
		ValidForStart: start,
	}
}

func (h *Handlers) mapRelatedParty(dto RelatedPartyDTO, customerID string) domain.RelatedParty {
	var start, end time.Time
	if dto.ValidForStart != "" {
		start, _ = time.Parse(time.RFC3339, dto.ValidForStart)
	}
	if dto.ValidForEnd != "" {
		parsedEnd, _ := time.Parse(time.RFC3339, dto.ValidForEnd)
		end = parsedEnd
	}

	res := domain.RelatedParty{
		ID:             dto.ID,
		CustomerID:     customerID,
		RelatedPartyID: dto.RelatedPartyID,
		Role:           dto.Role,
		Name:           dto.Name,
		ValidForStart:  start,
	}
	if !end.IsZero() {
		res.ValidForEnd = &end
	}
	return res
}

func (h *Handlers) mapPaymentMethod(dto PaymentMethodDTO, customerID string) domain.PaymentMethod {
	var start, end time.Time
	if dto.ValidForStart != "" {
		start, _ = time.Parse(time.RFC3339, dto.ValidForStart)
	}
	if dto.ValidForEnd != "" {
		parsedEnd, _ := time.Parse(time.RFC3339, dto.ValidForEnd)
		end = parsedEnd
	}

	res := domain.PaymentMethod{
		ID:            dto.ID,
		CustomerID:    customerID,
		Type:          dto.Type,
		Token:         dto.Token,
		Details:       json.RawMessage(dto.Details),
		IsDefault:     dto.IsDefault,
		ValidForStart: start,
	}
	if !end.IsZero() {
		res.ValidForEnd = &end
	}
	return res
}

func (h *Handlers) mapMarketSegment(dto MarketSegmentDTO, customerID string) domain.MarketSegment {
	return domain.MarketSegment{
		ID:         dto.ID,
		CustomerID: customerID,
		Name:       dto.Name,
		Category:   dto.Category,
	}
}
