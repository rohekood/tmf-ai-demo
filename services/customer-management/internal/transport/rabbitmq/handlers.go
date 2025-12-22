package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"tmf/services/customer-management/internal/domain"
	infraRabbit "tmf/services/customer-management/internal/infrastructure/rabbitmq"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Handlers manages command and query handling
type Handlers struct {
	repo      domain.Repository
	publisher *infraRabbit.Publisher
}

func NewHandlers(repo domain.Repository, publisher *infraRabbit.Publisher) *Handlers {
	return &Handlers{repo: repo, publisher: publisher}
}

// Payloads

type OnboardCustomerPayload struct {
	ID             string               `json:"id"`
	Name           string               `json:"name"`
	PartyID        string               `json:"partyId"`
	PartyType      string               `json:"partyType"`
	Accounts       []CustomerAccountDTO `json:"accounts"`
	CreditProfiles []CreditProfileDTO   `json:"creditProfiles"`
	ContactMediums []ContactMediumDTO   `json:"contactMediums"`
}

type CustomerAccountDTO struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	AccountStatus string `json:"accountStatus"`
	AccountType   string `json:"accountType"`
}

type CreditProfileDTO struct {
	ID              string `json:"id"`
	CreditRiskScore int    `json:"creditRiskScore"`
	CreditScore     int    `json:"creditScore"`
}

type ContactMediumDTO struct {
	ID         string `json:"id"`
	MediumType string `json:"mediumType"`
	Preferred  bool   `json:"preferred"`
	Value      string `json:"value"`
}

type UpdateCustomerPayload struct {
	ID     string                `json:"id"`
	Status domain.CustomerStatus `json:"status"`
	Name   string                `json:"name"`
}

type GetCustomerPayload struct {
	ID string `json:"id"`
}

// Handlers

func (h *Handlers) HandleOnboardCustomer(ctx context.Context, d amqp.Delivery) error {
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
			ID:         cm.ID,
			MediumType: cm.MediumType,
			Preferred:  cm.Preferred,
			Value:      cm.Value,
		})
	}

	if err := h.repo.CreateCustomer(ctx, customer); err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	// Publish event
	if err := h.publisher.Publish(ctx, d.Exchange, EvtCustomerCreated, customer); err != nil {
		log.Printf("failed to publish event: %v", err)
	}

	return nil
}

func (h *Handlers) HandleUpdateCustomer(ctx context.Context, d amqp.Delivery) error {
	var payload UpdateCustomerPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	customer, err := h.repo.GetCustomer(ctx, payload.ID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	oldStatus := customer.Status
	customer.Name = payload.Name
	customer.Status = payload.Status

	if err := h.repo.UpdateCustomer(ctx, customer); err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	// Publish events
	h.publisher.Publish(ctx, d.Exchange, EvtCustomerUpdated, customer)
	if oldStatus != customer.Status {
		h.publisher.Publish(ctx, d.Exchange, EvtCustomerStateChange, customer)
	}

	return nil
}

func (h *Handlers) HandleGetCustomer(ctx context.Context, d amqp.Delivery) error {
	var payload GetCustomerPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	customer, err := h.repo.GetCustomer(ctx, payload.ID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	// If ReplyTo is set, send the customer back
	if d.ReplyTo != "" {
		return h.publisher.Publish(ctx, "", d.ReplyTo, customer)
	}

	return nil
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
	var payload PartyEventPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal party event: %w", err)
	}

	switch d.RoutingKey {
	case EvtPartyUpdated:
		return h.handlePartyUpdated(ctx, payload)
	case EvtPartyDeleted:
		return h.handlePartyDeleted(ctx, payload)
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
				log.Printf("failed to sync party update to customer %s: %v", cust.ID, err)
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
			log.Printf("failed to close customer %s on party deletion: %v", cust.ID, err)
		}
	}
	return nil
}
