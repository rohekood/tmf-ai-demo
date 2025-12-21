package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"tmf/services/party-management/internal/domain"
	infraRabbit "tmf/services/party-management/internal/infrastructure/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	EventExchange = "tmf.events"

	// Commands
	CmdPartyCreate = "cmd.party.create"
	CmdPartyUpdate = "cmd.party.update"
	CmdPartyPatch  = "cmd.party.patch"
	CmdPartyDelete = "cmd.party.delete"

	// Queries
	QueryPartyGet    = "query.party.get"
	QueryPartySearch = "query.party.search"

	// Events
	EvtPartyCreated     = "evt.party.created"
	EvtPartyUpdated     = "evt.party.updated"
	EvtPartyDeleted     = "evt.party.deleted"
	EvtPartyStateChange = "evt.party.stateChange"
)

// Handlers manages command and query handling
type Handlers struct {
	repo      domain.Repository
	publisher *infraRabbit.Publisher
}

func NewHandlers(repo domain.Repository, publisher *infraRabbit.Publisher) *Handlers {
	return &Handlers{
		repo:      repo,
		publisher: publisher,
	}
}

// --- Command Payloads ---

type CreateIndividualPayload struct {
	ID         string `json:"id"`
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
	Href       string `json:"href"`
}

type CreateOrganizationPayload struct {
	ID            string `json:"id"`
	TradingName   string `json:"tradingName"`
	IsLegalEntity bool   `json:"isLegalEntity"`
	Href          string `json:"href"`
}

type CreatePartyPayload struct {
	Type         string                     `json:"@type"` // "Individual" or "Organization"
	Individual   *CreateIndividualPayload   `json:"individual,omitempty"`
	Organization *CreateOrganizationPayload `json:"organization,omitempty"`
}

type UpdatePartyPayload struct {
	ID           string                     `json:"id"`
	Type         string                     `json:"@type"`
	Status       string                     `json:"status,omitempty"`
	Individual   *CreateIndividualPayload   `json:"individual,omitempty"`
	Organization *CreateOrganizationPayload `json:"organization,omitempty"`
}

type PatchPartyPayload struct {
	ID         string  `json:"id"`
	GivenName  *string `json:"givenName,omitempty"`
	FamilyName *string `json:"familyName,omitempty"`
	Status     *string `json:"status,omitempty"`
}

type DeletePartyPayload struct {
	ID string `json:"id"`
}

type GetPartyPayload struct {
	ID string `json:"id"`
}

type SearchPartyPayload struct {
	GivenName   *string `json:"givenName,omitempty"`
	FamilyName  *string `json:"familyName,omitempty"`
	TradingName *string `json:"tradingName,omitempty"`
	Type        *string `json:"type,omitempty"`
}

// --- Handlers ---

func (h *Handlers) HandleCreateParty(d amqp.Delivery) error {
	var payload CreatePartyPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal CreatePartyPayload: %w", err)
	}

	now := time.Now()

	if payload.Type == "Individual" && payload.Individual != nil {
		ind := &domain.Individual{
			Party: domain.Party{
				ID:        payload.Individual.ID,
				Type:      domain.PartyTypeIndividual,
				Href:      payload.Individual.Href,
				Status:    "Initialized",
				CreatedAt: now,
				UpdatedAt: now,
			},
			GivenName:  payload.Individual.GivenName,
			FamilyName: payload.Individual.FamilyName,
		}
		if err := h.repo.CreateIndividual(ind); err != nil {
			return fmt.Errorf("failed to create individual: %w", err)
		}

		// Publish events
		h.publishEvent(EvtPartyCreated, ind)
		h.publishEvent(EvtPartyStateChange, map[string]interface{}{
			"id":       ind.ID,
			"newState": ind.Status,
		})
		log.Printf("Created Individual: %s", ind.ID)

	} else if payload.Type == "Organization" && payload.Organization != nil {
		org := &domain.Organization{
			Party: domain.Party{
				ID:        payload.Organization.ID,
				Type:      domain.PartyTypeOrganization,
				Href:      payload.Organization.Href,
				Status:    "Initialized",
				CreatedAt: now,
				UpdatedAt: now,
			},
			TradingName:   payload.Organization.TradingName,
			IsLegalEntity: payload.Organization.IsLegalEntity,
		}
		if err := h.repo.CreateOrganization(org); err != nil {
			return fmt.Errorf("failed to create organization: %w", err)
		}

		h.publishEvent(EvtPartyCreated, org)
		h.publishEvent(EvtPartyStateChange, map[string]interface{}{
			"id":       org.ID,
			"newState": org.Status,
		})
		log.Printf("Created Organization: %s", org.ID)
	} else {
		return fmt.Errorf("invalid party type or missing payload: %s", payload.Type)
	}

	return nil
}

func (h *Handlers) HandleUpdateParty(d amqp.Delivery) error {
	var payload UpdatePartyPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal UpdatePartyPayload: %w", err)
	}

	now := time.Now()

	if payload.Type == "Individual" && payload.Individual != nil {
		// Fetch existing to check state change
		existing, err := h.repo.GetIndividual(payload.ID)
		if err != nil {
			return fmt.Errorf("failed to get existing individual: %w", err)
		}
		oldStatus := existing.Status

		ind := &domain.Individual{
			Party: domain.Party{
				ID:        payload.ID,
				Type:      domain.PartyTypeIndividual,
				Href:      payload.Individual.Href,
				Status:    payload.Status,
				UpdatedAt: now,
			},
			GivenName:  payload.Individual.GivenName,
			FamilyName: payload.Individual.FamilyName,
		}
		if err := h.repo.UpdateIndividual(ind); err != nil {
			return fmt.Errorf("failed to update individual: %w", err)
		}

		h.publishEvent(EvtPartyUpdated, ind)
		if oldStatus != payload.Status && payload.Status != "" {
			h.publishEvent(EvtPartyStateChange, map[string]interface{}{
				"id":       ind.ID,
				"oldState": oldStatus,
				"newState": payload.Status,
			})
		}
		log.Printf("Updated Individual: %s", ind.ID)

	} else if payload.Type == "Organization" && payload.Organization != nil {
		existing, err := h.repo.GetOrganization(payload.ID)
		if err != nil {
			return fmt.Errorf("failed to get existing organization: %w", err)
		}
		oldStatus := existing.Status

		org := &domain.Organization{
			Party: domain.Party{
				ID:        payload.ID,
				Type:      domain.PartyTypeOrganization,
				Href:      payload.Organization.Href,
				Status:    payload.Status,
				UpdatedAt: now,
			},
			TradingName:   payload.Organization.TradingName,
			IsLegalEntity: payload.Organization.IsLegalEntity,
		}
		if err := h.repo.UpdateOrganization(org); err != nil {
			return fmt.Errorf("failed to update organization: %w", err)
		}

		h.publishEvent(EvtPartyUpdated, org)
		if oldStatus != payload.Status && payload.Status != "" {
			h.publishEvent(EvtPartyStateChange, map[string]interface{}{
				"id":       org.ID,
				"oldState": oldStatus,
				"newState": payload.Status,
			})
		}
		log.Printf("Updated Organization: %s", org.ID)
	} else {
		return fmt.Errorf("invalid party type for update: %s", payload.Type)
	}

	return nil
}

func (h *Handlers) HandlePatchParty(d amqp.Delivery) error {
	var payload PatchPartyPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal PatchPartyPayload: %w", err)
	}

	// For patch, we need to figure out the type. Try Individual first.
	existing, err := h.repo.GetIndividual(payload.ID)
	if err == nil {
		// It's an Individual
		oldStatus := existing.Status
		if payload.GivenName != nil {
			existing.GivenName = *payload.GivenName
		}
		if payload.FamilyName != nil {
			existing.FamilyName = *payload.FamilyName
		}
		if payload.Status != nil {
			existing.Status = *payload.Status
		}
		existing.UpdatedAt = time.Now()

		if err := h.repo.UpdateIndividual(existing); err != nil {
			return fmt.Errorf("failed to patch individual: %w", err)
		}

		h.publishEvent(EvtPartyUpdated, existing)
		if payload.Status != nil && oldStatus != *payload.Status {
			h.publishEvent(EvtPartyStateChange, map[string]interface{}{
				"id":       existing.ID,
				"oldState": oldStatus,
				"newState": *payload.Status,
			})
		}
		log.Printf("Patched Individual: %s", existing.ID)
		return nil
	}

	// Try Organization
	existingOrg, err := h.repo.GetOrganization(payload.ID)
	if err == nil {
		oldStatus := existingOrg.Status
		if payload.Status != nil {
			existingOrg.Status = *payload.Status
		}
		existingOrg.UpdatedAt = time.Now()

		if err := h.repo.UpdateOrganization(existingOrg); err != nil {
			return fmt.Errorf("failed to patch organization: %w", err)
		}

		h.publishEvent(EvtPartyUpdated, existingOrg)
		if payload.Status != nil && oldStatus != *payload.Status {
			h.publishEvent(EvtPartyStateChange, map[string]interface{}{
				"id":       existingOrg.ID,
				"oldState": oldStatus,
				"newState": *payload.Status,
			})
		}
		log.Printf("Patched Organization: %s", existingOrg.ID)
		return nil
	}

	return fmt.Errorf("party not found for patch: %s", payload.ID)
}

func (h *Handlers) HandleDeleteParty(d amqp.Delivery) error {
	var payload DeletePartyPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal DeletePartyPayload: %w", err)
	}

	if err := h.repo.DeleteParty(payload.ID); err != nil {
		return fmt.Errorf("failed to delete party: %w", err)
	}

	h.publishEvent(EvtPartyDeleted, map[string]interface{}{
		"id": payload.ID,
	})
	log.Printf("Deleted Party: %s", payload.ID)
	return nil
}

func (h *Handlers) HandleGetParty(d amqp.Delivery) error {
	var payload GetPartyPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal GetPartyPayload: %w", err)
	}

	// Try Individual
	if ind, err := h.repo.GetIndividual(payload.ID); err == nil {
		return h.replyTo(d, ind)
	}

	// Try Organization
	if org, err := h.repo.GetOrganization(payload.ID); err == nil {
		return h.replyTo(d, org)
	}

	return h.replyTo(d, map[string]string{"error": "party not found"})
}

func (h *Handlers) HandleSearchParty(d amqp.Delivery) error {
	var payload SearchPartyPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal SearchPartyPayload: %w", err)
	}

	criteria := make(map[string]interface{})
	if payload.GivenName != nil {
		criteria["given_name"] = *payload.GivenName
	}
	if payload.FamilyName != nil {
		criteria["family_name"] = *payload.FamilyName
	}
	if payload.TradingName != nil {
		criteria["trading_name"] = *payload.TradingName
	}
	if payload.Type != nil {
		criteria["type"] = *payload.Type
	}

	parties, err := h.repo.SearchParties(criteria)
	if err != nil {
		return h.replyTo(d, map[string]string{"error": err.Error()})
	}

	return h.replyTo(d, parties)
}

// --- Helpers ---

func (h *Handlers) publishEvent(routingKey string, event interface{}) {
	if err := h.publisher.Publish(EventExchange, routingKey, event); err != nil {
		log.Printf("Failed to publish event %s: %v", routingKey, err)
	}
}

func (h *Handlers) replyTo(d amqp.Delivery, response interface{}) error {
	if d.ReplyTo == "" {
		log.Printf("No ReplyTo queue specified, skipping reply")
		return nil
	}

	body, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Use the publisher's channel to reply
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// For RPC replies, we publish directly to the ReplyTo queue (default exchange)
	ch, err := h.publisher.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel for reply: %w", err)
	}

	return ch.PublishWithContext(ctx,
		"",        // default exchange
		d.ReplyTo, // routing key = reply queue
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: d.CorrelationId,
			Body:          body,
		})
}
