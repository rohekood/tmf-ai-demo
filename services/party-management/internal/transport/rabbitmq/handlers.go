package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"tmf/services/party-management/internal/domain"
	infraRabbit "tmf/services/party-management/internal/infrastructure/rabbitmq"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	EventExchange      = "tmf.events"
	DeadLetterExchange = "tmf.dlx"
	DeadLetterQueue    = "party.dlq"
	PartyQueue         = "party.commands"
	CommandExchange    = "tmf.party"

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
	ID              string              `json:"id"`
	GivenName       string              `json:"givenName"`
	FamilyName      string              `json:"familyName"`
	Href            string              `json:"href"`
	ContactMediums  []ContactMediumDTO  `json:"contactMediums,omitempty"`
	Identifications []IdentificationDTO `json:"identifications,omitempty"`
	RelatedParties  []RelatedPartyDTO   `json:"relatedParties,omitempty"`
	Characteristics []CharacteristicDTO `json:"characteristics,omitempty"`
}

func (p *CreateIndividualPayload) Validate() error {
	if p.ID == "" {
		return domain.ErrIDRequired
	}
	return nil
}

type CreateOrganizationPayload struct {
	ID              string              `json:"id"`
	TradingName     string              `json:"tradingName"`
	IsLegalEntity   bool                `json:"isLegalEntity"`
	Href            string              `json:"href"`
	ContactMediums  []ContactMediumDTO  `json:"contactMediums,omitempty"`
	Identifications []IdentificationDTO `json:"identifications,omitempty"`
	RelatedParties  []RelatedPartyDTO   `json:"relatedParties,omitempty"`
	Characteristics []CharacteristicDTO `json:"characteristics,omitempty"`
}

func (p *CreateOrganizationPayload) Validate() error {
	if p.ID == "" {
		return domain.ErrIDRequired
	}
	return nil
}

type CreatePartyPayload struct {
	Type         string                     `json:"@type"` // "Individual" or "Organization"
	Individual   *CreateIndividualPayload   `json:"individual,omitempty"`
	Organization *CreateOrganizationPayload `json:"organization,omitempty"`
}

func (p *CreatePartyPayload) Validate() error {
	switch p.Type {
	case "Individual":
		if p.Individual == nil {
			return fmt.Errorf("%w: missing individual payload", domain.ErrInvalidPayload)
		}
		return p.Individual.Validate()
	case "Organization":
		if p.Organization == nil {
			return fmt.Errorf("%w: missing organization payload", domain.ErrInvalidPayload)
		}
		return p.Organization.Validate()
	default:
		return domain.ErrInvalidType
	}
}

type UpdatePartyPayload struct {
	ID           string                     `json:"id"`
	Type         string                     `json:"@type"`
	Status       string                     `json:"status,omitempty"`
	Individual   *CreateIndividualPayload   `json:"individual,omitempty"`
	Organization *CreateOrganizationPayload `json:"organization,omitempty"`
}

func (p *UpdatePartyPayload) Validate() error {
	if p.ID == "" {
		return domain.ErrIDRequired
	}
	switch p.Type {
	case "Individual":
		if p.Individual == nil {
			return fmt.Errorf("%w: missing individual payload", domain.ErrInvalidPayload)
		}
		return p.Individual.Validate()
	case "Organization":
		if p.Organization == nil {
			return fmt.Errorf("%w: missing organization payload", domain.ErrInvalidPayload)
		}
		return p.Organization.Validate()
	default:
		return domain.ErrInvalidType
	}
}

type PatchPartyPayload struct {
	ID         string  `json:"id"`
	GivenName  *string `json:"givenName,omitempty"`
	FamilyName *string `json:"familyName,omitempty"`
	Status     *string `json:"status,omitempty"`
}

func (p *PatchPartyPayload) Validate() error {
	if p.ID == "" {
		return domain.ErrIDRequired
	}
	return nil
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

type IdentificationDTO struct {
	ID                 string `json:"id"`
	IdentificationType string `json:"identificationType"`
	IdentificationID   string `json:"identificationId"`
	IssuingAuthority   string `json:"issuingAuthority"`
	IssuingDate        string `json:"issuingDate"` // string for parsing
}

type RelatedPartyDTO struct {
	ID               string `json:"id"`
	RelatedPartyID   string `json:"relatedPartyId"`
	RelatedPartyName string `json:"relatedPartyName"`
	Role             string `json:"role"`
}

type CharacteristicDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	ValueType string `json:"valueType"`
}

// --- Handlers ---

func (h *Handlers) HandleCreateParty(ctx context.Context, d amqp.Delivery) error {
	ctx = h.extractUser(ctx, d)

	// Determine type from flat JSON
	var typeInfo struct {
		Type string `json:"@type"`
	}
	if err := json.Unmarshal(d.Body, &typeInfo); err != nil {
		return fmt.Errorf("failed to unmarshal type info: %w", err)
	}

	now := time.Now()

	if typeInfo.Type == "Individual" {
		var payload CreateIndividualPayload
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal Individual payload: %w", err)
		}

		if payload.ID == "" {
			payload.ID = uuid.New().String()
		}

		if err := payload.Validate(); err != nil {
			return err
		}

		ind := &domain.Individual{
			Party: domain.Party{
				ID:        payload.ID,
				Type:      domain.PartyTypeIndividual,
				Href:      payload.Href,
				Status:    "Initialized",
				CreatedAt: now,
				UpdatedAt: now,
			},
			GivenName:  payload.GivenName,
			FamilyName: payload.FamilyName,
		}

		ind.ContactMediums = h.mapContactMediums(payload.ContactMediums, ind.ID)
		ind.Identifications = h.mapIdentifications(payload.Identifications, ind.ID)
		ind.RelatedParties = h.mapRelatedParties(payload.RelatedParties, ind.ID)
		ind.Characteristics = h.mapCharacteristics(payload.Characteristics, ind.ID)

		if err := h.repo.CreateIndividual(ctx, ind); err != nil {
			return fmt.Errorf("failed to create individual: %w", err)
		}

		// Publish events
		h.publishEvent(ctx, EvtPartyCreated, ind)
		h.publishEvent(ctx, EvtPartyStateChange, map[string]interface{}{
			"id":       ind.ID,
			"newState": ind.Status,
		})
		slog.Info("created individual", "party_id", ind.ID)

		return h.replyTo(ctx, d, ind)

	} else if typeInfo.Type == "Organization" {
		var payload CreateOrganizationPayload
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal Organization payload: %w", err)
		}

		if payload.ID == "" {
			payload.ID = uuid.New().String()
		}

		if err := payload.Validate(); err != nil {
			return err
		}

		org := &domain.Organization{
			Party: domain.Party{
				ID:        payload.ID,
				Type:      domain.PartyTypeOrganization,
				Href:      payload.Href,
				Status:    "Initialized",
				CreatedAt: now,
				UpdatedAt: now,
			},
			TradingName:   payload.TradingName,
			IsLegalEntity: payload.IsLegalEntity,
		}

		org.ContactMediums = h.mapContactMediums(payload.ContactMediums, org.ID)
		org.Identifications = h.mapIdentifications(payload.Identifications, org.ID)
		org.RelatedParties = h.mapRelatedParties(payload.RelatedParties, org.ID)
		org.Characteristics = h.mapCharacteristics(payload.Characteristics, org.ID)

		if err := h.repo.CreateOrganization(ctx, org); err != nil {
			return fmt.Errorf("failed to create organization: %w", err)
		}

		h.publishEvent(ctx, EvtPartyCreated, org)
		h.publishEvent(ctx, EvtPartyStateChange, map[string]interface{}{
			"id":       org.ID,
			"newState": org.Status,
		})
		slog.Info("created organization", "party_id", org.ID)

		return h.replyTo(ctx, d, org)
	}

	return domain.ErrInvalidType
}

func (h *Handlers) HandleUpdateParty(ctx context.Context, d amqp.Delivery) error {
	ctx = h.extractUser(ctx, d)

	var typeInfo struct {
		Type string `json:"@type"`
	}
	if err := json.Unmarshal(d.Body, &typeInfo); err != nil {
		return fmt.Errorf("failed to unmarshal type info: %w", err)
	}

	now := time.Now()

	if typeInfo.Type == "Individual" {
		var payload CreateIndividualPayload // Use Same payload structure for update fields
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal Update payload: %w", err)
		}

		// Status is handled separately as it's not in CreateIndividualPayload
		var statusPayload struct {
			Status string `json:"status"`
		}
		json.Unmarshal(d.Body, &statusPayload) // Ignore error, optional

		if payload.ID == "" {
			return domain.ErrIDRequired
		}

		existing, err := h.repo.GetIndividual(ctx, payload.ID)
		if err != nil {
			return fmt.Errorf("failed to get existing individual: %w", err)
		}
		oldStatus := existing.Status
		newStatus := statusPayload.Status
		if newStatus == "" {
			newStatus = oldStatus // Keep existing if not provided
		}

		ind := &domain.Individual{
			Party: domain.Party{
				ID:        payload.ID,
				Type:      domain.PartyTypeIndividual,
				Href:      payload.Href,
				Status:    newStatus,
				CreatedAt: existing.CreatedAt, // Preserve creation time
				UpdatedAt: now,
			},
			GivenName:  payload.GivenName,
			FamilyName: payload.FamilyName,
		}
		if err := h.repo.UpdateIndividual(ctx, ind); err != nil {
			return fmt.Errorf("failed to update individual: %w", err)
		}

		h.publishEvent(ctx, EvtPartyUpdated, ind)
		if oldStatus != newStatus {
			h.publishEvent(ctx, EvtPartyStateChange, map[string]interface{}{
				"id":       ind.ID,
				"oldState": oldStatus,
				"newState": newStatus,
			})
		}
		slog.Info("updated individual", "party_id", ind.ID)

		return h.replyTo(ctx, d, ind)

	} else if typeInfo.Type == "Organization" {
		var payload CreateOrganizationPayload
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal Organization update payload: %w", err)
		}

		var statusPayload struct {
			Status string `json:"status"`
		}
		json.Unmarshal(d.Body, &statusPayload)

		if payload.ID == "" {
			return domain.ErrIDRequired
		}

		existing, err := h.repo.GetOrganization(ctx, payload.ID)
		if err != nil {
			return fmt.Errorf("failed to get existing organization: %w", err)
		}
		oldStatus := existing.Status
		newStatus := statusPayload.Status
		if newStatus == "" {
			newStatus = oldStatus
		}

		org := &domain.Organization{
			Party: domain.Party{
				ID:        payload.ID,
				Type:      domain.PartyTypeOrganization,
				Href:      payload.Href,
				Status:    newStatus,
				CreatedAt: existing.CreatedAt,
				UpdatedAt: now,
			},
			TradingName:   payload.TradingName,
			IsLegalEntity: payload.IsLegalEntity,
		}
		if err := h.repo.UpdateOrganization(ctx, org); err != nil {
			return fmt.Errorf("failed to update organization: %w", err)
		}

		h.publishEvent(ctx, EvtPartyUpdated, org)
		if oldStatus != newStatus {
			h.publishEvent(ctx, EvtPartyStateChange, map[string]interface{}{
				"id":       org.ID,
				"oldState": oldStatus,
				"newState": newStatus,
			})
		}
		slog.Info("updated organization", "party_id", org.ID)

		return h.replyTo(ctx, d, org)
	}

	return domain.ErrInvalidType
}

func (h *Handlers) HandlePatchParty(ctx context.Context, d amqp.Delivery) error {
	ctx = h.extractUser(ctx, d)
	var payload PatchPartyPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal PatchPartyPayload: %w", err)
	}

	if err := payload.Validate(); err != nil {
		return err
	}

	// Figure out type from database
	party, err := h.repo.GetParty(ctx, payload.ID)
	if err != nil {
		return fmt.Errorf("failed to get party for patch: %w", err)
	}

	if party.Type == domain.PartyTypeIndividual {
		existing, err := h.repo.GetIndividual(ctx, payload.ID)
		if err != nil {
			return err
		}
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

		if err := h.repo.UpdateIndividual(ctx, existing); err != nil {
			return fmt.Errorf("failed to patch individual: %w", err)
		}

		h.publishEvent(ctx, EvtPartyUpdated, existing)
		if payload.Status != nil && oldStatus != *payload.Status {
			h.publishEvent(ctx, EvtPartyStateChange, map[string]interface{}{
				"id":       existing.ID,
				"oldState": oldStatus,
				"newState": *payload.Status,
			})
		}
		slog.Info("patched individual", "party_id", existing.ID)

		return h.replyTo(ctx, d, existing)

	} else if party.Type == domain.PartyTypeOrganization {
		existingOrg, err := h.repo.GetOrganization(ctx, payload.ID)
		if err != nil {
			return err
		}
		oldStatus := existingOrg.Status
		if payload.Status != nil {
			existingOrg.Status = *payload.Status
		}
		existingOrg.UpdatedAt = time.Now()

		if err := h.repo.UpdateOrganization(ctx, existingOrg); err != nil {
			return fmt.Errorf("failed to patch organization: %w", err)
		}

		h.publishEvent(ctx, EvtPartyUpdated, existingOrg)
		if payload.Status != nil && oldStatus != *payload.Status {
			h.publishEvent(ctx, EvtPartyStateChange, map[string]interface{}{
				"id":       existingOrg.ID,
				"oldState": oldStatus,
				"newState": *payload.Status,
			})
		}
		slog.Info("patched organization", "party_id", existingOrg.ID)

		return h.replyTo(ctx, d, existingOrg)
	}

	return fmt.Errorf("%w: %s", domain.ErrInvalidType, payload.ID)
}

func (h *Handlers) HandleDeleteParty(ctx context.Context, d amqp.Delivery) error {
	ctx = h.extractUser(ctx, d)
	var payload DeletePartyPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal DeletePartyPayload: %w", err)
	}

	if payload.ID == "" {
		return domain.ErrIDRequired
	}

	if err := h.repo.DeleteParty(ctx, payload.ID); err != nil {
		return fmt.Errorf("failed to delete party: %w", err)
	}

	h.publishEvent(ctx, EvtPartyDeleted, map[string]interface{}{
		"id": payload.ID,
	})
	slog.Info("deleted party", "party_id", payload.ID)

	return h.replyTo(ctx, d, map[string]interface{}{})
}

func (h *Handlers) HandleGetParty(ctx context.Context, d amqp.Delivery) error {
	ctx = h.extractUser(ctx, d)
	var payload GetPartyPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal GetPartyPayload: %w", err)
	}

	if payload.ID == "" {
		return h.replyTo(ctx, d, map[string]string{"error": domain.ErrIDRequired.Error()})
	}

	party, err := h.repo.GetParty(ctx, payload.ID)
	if err != nil {
		return h.replyTo(ctx, d, map[string]string{"error": err.Error()})
	}

	if party.Type == domain.PartyTypeIndividual {
		ind, _ := h.repo.GetIndividual(ctx, payload.ID)
		return h.replyTo(ctx, d, ind)
	} else {
		org, _ := h.repo.GetOrganization(ctx, payload.ID)
		return h.replyTo(ctx, d, org)
	}
}

func (h *Handlers) HandleSearchParty(ctx context.Context, d amqp.Delivery) error {
	ctx = h.extractUser(ctx, d)
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

	parties, err := h.repo.SearchParties(ctx, criteria)
	if err != nil {
		return h.replyTo(ctx, d, map[string]string{"error": err.Error()})
	}

	return h.replyTo(ctx, d, parties)
}

// --- Helpers ---

func (h *Handlers) publishEvent(ctx context.Context, routingKey string, event interface{}) {
	if h.publisher == nil {
		slog.Warn("publisher is nil, skipping event publishing", "routingKey", routingKey)
		return
	}
	if err := h.publisher.Publish(ctx, EventExchange, routingKey, event); err != nil {
		slog.Error("failed to publish event", "routingKey", routingKey, "error", err)
	}
}

func (h *Handlers) replyTo(ctx context.Context, d amqp.Delivery, response interface{}) error {
	if d.ReplyTo == "" {
		slog.Debug("no ReplyTo queue specified, skipping reply")
		return nil
	}

	body, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

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
			Headers:       amqp.Table{"user": ctx.Value(domain.UserContextKey)},
			CorrelationId: d.CorrelationId,
			Body:          body,
		})
}

func (h *Handlers) extractUser(ctx context.Context, d amqp.Delivery) context.Context {
	if user, ok := d.Headers["user"].(string); ok && user != "" {
		return context.WithValue(ctx, domain.UserContextKey, user)
	}
	return ctx
}
func (h *Handlers) mapContactMediums(dtos []ContactMediumDTO, partyID string) []domain.ContactMedium {
	res := make([]domain.ContactMedium, 0, len(dtos))
	for _, dto := range dtos {
		res = append(res, domain.ContactMedium{
			ID:              dto.ID,
			PartyID:         partyID,
			MediumType:      dto.MediumType,
			Preferred:       dto.Preferred,
			Value:           dto.Value,
			Street1:         dto.Street1,
			Street2:         dto.Street2,
			City:            dto.City,
			StateOrProvince: dto.StateOrProvince,
			Postcode:        dto.Postcode,
			Country:         dto.Country,
		})
	}
	return res
}

func (h *Handlers) mapIdentifications(dtos []IdentificationDTO, partyID string) []domain.Identification {
	res := make([]domain.Identification, 0, len(dtos))
	for _, dto := range dtos {
		var issuingDate time.Time
		if dto.IssuingDate != "" {
			issuingDate, _ = time.Parse(time.RFC3339, dto.IssuingDate)
		}
		res = append(res, domain.Identification{
			ID:                 dto.ID,
			PartyID:            partyID,
			IdentificationType: dto.IdentificationType,
			IdentificationID:   dto.IdentificationID,
			IssuingAuthority:   dto.IssuingAuthority,
			IssuingDate:        issuingDate,
		})
	}
	return res
}

func (h *Handlers) mapRelatedParties(dtos []RelatedPartyDTO, partyID string) []domain.RelatedParty {
	res := make([]domain.RelatedParty, 0, len(dtos))
	for _, dto := range dtos {
		res = append(res, domain.RelatedParty{
			ID:               dto.ID,
			PartyID:          partyID,
			RelatedPartyID:   dto.RelatedPartyID,
			RelatedPartyName: dto.RelatedPartyName,
			Role:             dto.Role,
		})
	}
	return res
}

func (h *Handlers) mapCharacteristics(dtos []CharacteristicDTO, partyID string) []domain.PartyCharacteristic {
	res := make([]domain.PartyCharacteristic, 0, len(dtos))
	for _, dto := range dtos {
		res = append(res, domain.PartyCharacteristic{
			ID:        dto.ID,
			PartyID:   partyID,
			Name:      dto.Name,
			Value:     dto.Value,
			ValueType: dto.ValueType,
		})
	}
	return res
}
