package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"tmf/pkg/rabbitmq"
	"tmf/services/party-management/internal/domain"

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
	EvtPartyCreated           = "evt.party.created"
	EvtPartyUpdated           = "evt.party.updated"
	EvtPartyDeleted           = "evt.party.deleted"
	EvtPartyStateChange       = "evt.party.stateChange"
	EvtPartyDeletionInitiated = "evt.party.deletion_initiated"

	// External Events
	EvtCustomerCreated = "evt.customer.created"

	// Saga Commands
	CmdPartyFinalizeDeletion = "cmd.party.finalize_deletion"
	CmdPartyCancelDeletion   = "cmd.party.cancel_deletion"
	CmdPartyPurge            = "cmd.party.purge"
)

// CustomerChecker is a port for checking whether a party has linked customers.
// Implemented by an adapter in main.go using the shared RPC client.
type CustomerChecker interface {
	HasCustomers(ctx context.Context, partyID string) (bool, error)
}

// Handlers manages command and query handling
type Handlers struct {
	repo            domain.Repository
	eventPublisher  domain.EventPublisher
	rpcPublisher    rabbitmq.Publisher
	tm              domain.TransactionManager
	customerChecker CustomerChecker
}

func NewHandlers(repo domain.Repository, eventPublisher domain.EventPublisher, rpcPublisher rabbitmq.Publisher, tm domain.TransactionManager) *Handlers {
	return &Handlers{
		repo:           repo,
		eventPublisher: eventPublisher,
		rpcPublisher:   rpcPublisher,
		tm:             tm,
	}
}

// WithCustomerChecker injects the optional customer checker dependency.
func (h *Handlers) WithCustomerChecker(c CustomerChecker) *Handlers {
	h.customerChecker = c
	return h
}

// --- Command Payloads ---

type CreateIndividualPayload struct {
	ID                 string                 `json:"id"`
	GivenName          string                 `json:"givenName"`
	FamilyName         string                 `json:"familyName"`
	MiddleName         string                 `json:"middleName,omitempty"`
	BirthDate          string                 `json:"birthDate,omitempty"`
	Gender             string                 `json:"gender,omitempty"`
	Href               string                 `json:"href"`
	ContactMediums     []ContactMediumDTO     `json:"contactMediums,omitempty"`
	Identifications    []IdentificationDTO    `json:"identifications,omitempty"`
	RelatedParties     []RelatedPartyDTO      `json:"relatedParties,omitempty"`
	Characteristics    []CharacteristicDTO    `json:"characteristics,omitempty"`
	ExternalReferences []ExternalReferenceDTO `json:"externalReferences,omitempty"`
	TaxExemptions      []TaxExemptionDTO      `json:"taxExemptions,omitempty"`
	Attachments        []AttachmentDTO        `json:"attachments,omitempty"`
}

func (p *CreateIndividualPayload) Validate() error {
	if p.ID == "" {
		return domain.ErrIDRequired
	}
	return nil
}

type CreateOrganizationPayload struct {
	ID                 string                 `json:"id"`
	TradingName        string                 `json:"tradingName"`
	IsLegalEntity      bool                   `json:"isLegalEntity"`
	OrganizationType   string                 `json:"organizationType,omitempty"`
	Href               string                 `json:"href"`
	ContactMediums     []ContactMediumDTO     `json:"contactMediums,omitempty"`
	Identifications    []IdentificationDTO    `json:"identifications,omitempty"`
	RelatedParties     []RelatedPartyDTO      `json:"relatedParties,omitempty"`
	Characteristics    []CharacteristicDTO    `json:"characteristics,omitempty"`
	ExternalReferences []ExternalReferenceDTO `json:"externalReferences,omitempty"`
	TaxExemptions      []TaxExemptionDTO      `json:"taxExemptions,omitempty"`
	Attachments        []AttachmentDTO        `json:"attachments,omitempty"`
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
	Search            *string `json:"search,omitempty"`
	Name              *string `json:"name,omitempty"`
	GivenName         *string `json:"givenName,omitempty"`
	FamilyName        *string `json:"familyName,omitempty"`
	TradingName       *string `json:"tradingName,omitempty"`
	Type              *string `json:"type,omitempty"`
	ExternalReference *string `json:"externalReference,omitempty"`
	Email             *string `json:"email,omitempty"`
	Status            *string `json:"status,omitempty"`
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
	ID               string   `json:"id"`
	RelatedPartyID   string   `json:"relatedPartyId"`
	RelatedPartyName string   `json:"relatedPartyName"`
	Role             string   `json:"role"`
	Permissions      []string `json:"permissions"`
}

type CharacteristicDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	ValueType string `json:"valueType"`
}

type ExternalReferenceDTO struct {
	ID                  string `json:"id"`
	ExternalSystemID    string `json:"externalSystemId"`
	ExternalReferenceID string `json:"externalReferenceId"`
}

type TaxExemptionDTO struct {
	ID                  string `json:"id"`
	CertificateNumber   string `json:"certificateNumber"`
	IssuingJurisdiction string `json:"issuingJurisdiction"`
	ValidForStart       string `json:"validForStart"`
	ValidForEnd         string `json:"validForEnd"`
}

type AttachmentDTO struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	MimeType       string `json:"mimeType"`
	AttachmentType string `json:"attachmentType"`
	RefType        string `json:"refType"`
	RefID          string `json:"refId"`
	URL            string `json:"url,omitempty"`
	Content        []byte `json:"content,omitempty"` // For upload
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

	now := time.Now().UTC()

	switch typeInfo.Type {
	case "Individual":
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
			MiddleName: payload.MiddleName,
			BirthDate:  payload.BirthDate,
			Gender:     payload.Gender,
		}

		ind.ContactMediums = h.mapContactMediums(payload.ContactMediums, ind.ID)
		ind.Identifications = h.mapIdentifications(payload.Identifications, ind.ID)
		ind.RelatedParties = h.mapRelatedParties(payload.RelatedParties, ind.ID)
		ind.Characteristics = h.mapCharacteristics(payload.Characteristics, ind.ID)
		ind.ExternalReferences = h.mapExternalReferences(payload.ExternalReferences, ind.ID)
		ind.TaxExemptions = h.mapTaxExemptions(payload.TaxExemptions, ind.ID)
		ind.Attachments = h.mapAttachments(payload.Attachments, ind.ID)

		if err := h.tm.Run(ctx, func(txCtx context.Context) error {
			if err := h.repo.CreateIndividual(txCtx, ind); err != nil {
				return fmt.Errorf("failed to create individual: %w", err)
			}

			if err := h.publishEvent(txCtx, EvtPartyCreated, ind); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		slog.Info("created individual", "party_id", ind.ID)
		return h.replyTo(ctx, d, ind)

	case "Organization":
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
			TradingName:      payload.TradingName,
			IsLegalEntity:    payload.IsLegalEntity,
			OrganizationType: payload.OrganizationType,
		}

		org.ContactMediums = h.mapContactMediums(payload.ContactMediums, org.ID)
		org.Identifications = h.mapIdentifications(payload.Identifications, org.ID)
		org.RelatedParties = h.mapRelatedParties(payload.RelatedParties, org.ID)
		org.Characteristics = h.mapCharacteristics(payload.Characteristics, org.ID)
		org.ExternalReferences = h.mapExternalReferences(payload.ExternalReferences, org.ID)
		org.TaxExemptions = h.mapTaxExemptions(payload.TaxExemptions, org.ID)
		org.Attachments = h.mapAttachments(payload.Attachments, org.ID)

		if err := h.tm.Run(ctx, func(txCtx context.Context) error {
			if err := h.repo.CreateOrganization(txCtx, org); err != nil {
				return fmt.Errorf("failed to create organization: %w", err)
			}

			if err := h.publishEvent(txCtx, EvtPartyCreated, org); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		slog.Info("created organization", "party_id", org.ID)
		return h.replyTo(ctx, d, org)

	default:
		return domain.ErrInvalidType
	}
}

func (h *Handlers) HandleUpdateParty(ctx context.Context, d amqp.Delivery) error {
	ctx = h.extractUser(ctx, d)

	var typeInfo struct {
		Type string `json:"@type"`
	}
	if err := json.Unmarshal(d.Body, &typeInfo); err != nil {
		return fmt.Errorf("failed to unmarshal type info: %w", err)
	}

	now := time.Now().UTC()

	switch typeInfo.Type {
	case "Individual":
		var payload CreateIndividualPayload // Use Same payload structure for update fields
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal Update payload: %w", err)
		}

		// Status is handled separately as it's not in CreateIndividualPayload
		var statusPayload struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(d.Body, &statusPayload); err != nil {
			// Optional, but if malformed JSON we might warn
			slog.Warn("failed to unmarshal status from update body", "error", err)
		}

		if payload.ID == "" {
			return domain.ErrIDRequired
		}

		existingParty, err := h.repo.GetParty(ctx, payload.ID)
		if err != nil {
			return fmt.Errorf("failed to get existing party: %w", err)
		}
		oldStatus := existingParty.Status
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
				CreatedAt: existingParty.CreatedAt, // Preserve creation time
				UpdatedAt: now,
			},
			GivenName:  payload.GivenName,
			FamilyName: payload.FamilyName,
			MiddleName: payload.MiddleName,
			BirthDate:  payload.BirthDate,
			Gender:     payload.Gender,
		}

		ind.ContactMediums = h.mapContactMediums(payload.ContactMediums, ind.ID)
		ind.Identifications = h.mapIdentifications(payload.Identifications, ind.ID)
		ind.RelatedParties = h.mapRelatedParties(payload.RelatedParties, ind.ID)
		ind.Characteristics = h.mapCharacteristics(payload.Characteristics, ind.ID)
		ind.ExternalReferences = h.mapExternalReferences(payload.ExternalReferences, ind.ID)
		ind.TaxExemptions = h.mapTaxExemptions(payload.TaxExemptions, ind.ID)
		ind.Attachments = h.mapAttachments(payload.Attachments, ind.ID)

		if err := h.tm.Run(ctx, func(txCtx context.Context) error {
			if err := h.repo.UpdateIndividual(txCtx, ind); err != nil {
				return fmt.Errorf("failed to update individual: %w", err)
			}

			if err := h.publishEvent(txCtx, EvtPartyUpdated, ind); err != nil {
				return err
			}
			if oldStatus != newStatus {
				if err := h.publishEvent(txCtx, EvtPartyStateChange, map[string]any{
					"id":       ind.ID,
					"oldState": oldStatus,
					"newState": newStatus,
				}); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}

		slog.Info("updated individual", "party_id", ind.ID)

		return h.replyTo(ctx, d, ind)

	case "Organization":
		var payload CreateOrganizationPayload
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal Organization update payload: %w", err)
		}

		var statusPayload struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(d.Body, &statusPayload); err != nil {
			slog.Warn("failed to unmarshal status from update body", "error", err)
		}

		if payload.ID == "" {
			return domain.ErrIDRequired
		}

		existingParty, err := h.repo.GetParty(ctx, payload.ID)
		if err != nil {
			return fmt.Errorf("failed to get existing party: %w", err)
		}
		oldStatus := existingParty.Status
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
				CreatedAt: existingParty.CreatedAt,
				UpdatedAt: now,
			},
			TradingName:      payload.TradingName,
			IsLegalEntity:    payload.IsLegalEntity,
			OrganizationType: payload.OrganizationType,
		}

		org.ContactMediums = h.mapContactMediums(payload.ContactMediums, org.ID)
		org.Identifications = h.mapIdentifications(payload.Identifications, org.ID)
		org.RelatedParties = h.mapRelatedParties(payload.RelatedParties, org.ID)
		org.Characteristics = h.mapCharacteristics(payload.Characteristics, org.ID)
		org.ExternalReferences = h.mapExternalReferences(payload.ExternalReferences, org.ID)
		org.TaxExemptions = h.mapTaxExemptions(payload.TaxExemptions, org.ID)
		org.Attachments = h.mapAttachments(payload.Attachments, org.ID)

		if err := h.tm.Run(ctx, func(txCtx context.Context) error {
			if err := h.repo.UpdateOrganization(txCtx, org); err != nil {
				return fmt.Errorf("failed to update organization: %w", err)
			}

			if err := h.publishEvent(txCtx, EvtPartyUpdated, org); err != nil {
				return err
			}
			if oldStatus != newStatus {
				if err := h.publishEvent(txCtx, EvtPartyStateChange, map[string]any{
					"id":       org.ID,
					"oldState": oldStatus,
					"newState": newStatus,
				}); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}

		slog.Info("updated organization", "party_id", org.ID)

		return h.replyTo(ctx, d, org)

	default:
		return domain.ErrInvalidType
	}
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

	switch party.Type {
	case domain.PartyTypeIndividual:
		var updatedInd *domain.Individual
		if err := h.tm.Run(ctx, func(txCtx context.Context) error {
			existing, err := h.repo.GetIndividual(txCtx, payload.ID)
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
			existing.UpdatedAt = time.Now().UTC()

			if err := h.repo.UpdateIndividual(txCtx, existing); err != nil {
				return fmt.Errorf("failed to patch individual: %w", err)
			}

			if err := h.publishEvent(txCtx, EvtPartyUpdated, existing); err != nil {
				return err
			}
			if payload.Status != nil && oldStatus != *payload.Status {
				if err := h.publishEvent(txCtx, EvtPartyStateChange, map[string]any{
					"id":       existing.ID,
					"oldState": oldStatus,
					"newState": *payload.Status,
				}); err != nil {
					return err
				}
			}
			updatedInd = existing
			return nil
		}); err != nil {
			return err
		}

		slog.Info("patched individual", "party_id", updatedInd.ID)
		return h.replyTo(ctx, d, updatedInd)

	case domain.PartyTypeOrganization:
		var updatedOrg *domain.Organization
		if err := h.tm.Run(ctx, func(txCtx context.Context) error {
			existingOrg, err := h.repo.GetOrganization(txCtx, payload.ID)
			if err != nil {
				return err
			}
			oldStatus := existingOrg.Status
			if payload.Status != nil {
				existingOrg.Status = *payload.Status
			}
			existingOrg.UpdatedAt = time.Now().UTC()

			if err := h.repo.UpdateOrganization(txCtx, existingOrg); err != nil {
				return fmt.Errorf("failed to patch organization: %w", err)
			}

			if err := h.publishEvent(txCtx, EvtPartyUpdated, existingOrg); err != nil {
				return err
			}
			if payload.Status != nil && oldStatus != *payload.Status {
				if err := h.publishEvent(txCtx, EvtPartyStateChange, map[string]any{
					"id":       existingOrg.ID,
					"oldState": oldStatus,
					"newState": *payload.Status,
				}); err != nil {
					return err
				}
			}
			updatedOrg = existingOrg
			return nil
		}); err != nil {
			return err
		}

		slog.Info("patched organization", "party_id", updatedOrg.ID)
		return h.replyTo(ctx, d, updatedOrg)
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

	// Pre-check: reject immediately if the party has linked customers.
	if h.customerChecker != nil {
		hasCustomers, err := h.customerChecker.HasCustomers(ctx, payload.ID)
		if err != nil {
			slog.Warn("customer pre-check failed, proceeding with saga", "party_id", payload.ID, "error", err)
		} else if hasCustomers {
			return h.replyTo(ctx, d, map[string]string{"error": "party has linked customers and cannot be deleted"})
		}
	}

	// 1. Get current party
	if err := h.tm.Run(ctx, func(txCtx context.Context) error {
		party, err := h.repo.GetParty(txCtx, payload.ID)
		if err != nil {
			return fmt.Errorf("failed to get party: %w", err)
		}

		// 2. Initiate Deletion (Saga Start)
		oldStatus := party.Status
		newStatus := string(domain.PartyStatusDeletionPending)

		if oldStatus == newStatus {
			slog.Info("party deletion already pending", "party_id", payload.ID)
			return nil // Already pending, just return success
		}

		if party.Type == domain.PartyTypeIndividual {
			ind, err := h.repo.GetIndividual(txCtx, payload.ID)
			if err != nil {
				return err
			}
			ind.Status = newStatus
			if err := h.repo.UpdateIndividual(txCtx, ind); err != nil {
				return fmt.Errorf("failed to update status to pending: %w", err)
			}
		} else {
			org, err := h.repo.GetOrganization(txCtx, payload.ID)
			if err != nil {
				return err
			}
			org.Status = newStatus
			if err := h.repo.UpdateOrganization(txCtx, org); err != nil {
				return fmt.Errorf("failed to update status to pending: %w", err)
			}
		}

		// 3. Publish Deletion Initiated Event
		if err := h.publishEvent(txCtx, EvtPartyDeletionInitiated, map[string]any{
			"id":   payload.ID,
			"type": party.Type,
		}); err != nil {
			return err
		}

		// Also publish state change
		if err := h.publishEvent(txCtx, EvtPartyStateChange, map[string]any{
			"id":       payload.ID,
			"oldState": oldStatus,
			"newState": newStatus,
		}); err != nil {
			return err
		}

		slog.Info("party deletion initiated", "party_id", payload.ID)
		return nil
	}); err != nil {
		return err
	}

	return h.replyTo(ctx, d, map[string]string{"status": "deletion_initiated"})
}

func (h *Handlers) HandleFinalizeDeletion(ctx context.Context, d amqp.Delivery) error {
	ctx = h.extractUser(ctx, d)
	var payload DeletePartyPayload // Reusing payload structure as it has ID
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return h.tm.Run(ctx, func(txCtx context.Context) error {
		party, err := h.repo.GetParty(txCtx, payload.ID)
		if err != nil {
			return err
		}

		if party.Status != string(domain.PartyStatusDeletionPending) {
			slog.Warn("skipping finalize deletion: party not in pending state", "id", payload.ID, "status", party.Status)
			return nil
		}

		// Soft Delete: Update to "Deleted"
		newStatus := string(domain.PartyStatusDeleted)

		if party.Type == domain.PartyTypeIndividual {
			ind, _ := h.repo.GetIndividual(txCtx, payload.ID)
			ind.Status = newStatus
			if err := h.repo.UpdateIndividual(txCtx, ind); err != nil {
				return fmt.Errorf("failed to finalize deletion (ind): %w", err)
			}
		} else {
			org, _ := h.repo.GetOrganization(txCtx, payload.ID)
			org.Status = newStatus
			if err := h.repo.UpdateOrganization(txCtx, org); err != nil {
				return fmt.Errorf("failed to finalize deletion (org): %w", err)
			}
		}

		if err := h.publishEvent(txCtx, EvtPartyDeleted, map[string]any{"id": payload.ID}); err != nil {
			return err
		}
		if err := h.publishEvent(txCtx, EvtPartyStateChange, map[string]any{
			"id":       payload.ID,
			"oldState": domain.PartyStatusDeletionPending,
			"newState": newStatus,
		}); err != nil {
			return err
		}

		slog.Info("party deletion finalized (soft delete)", "id", payload.ID)
		return nil
	})
}

func (h *Handlers) HandleCancelDeletion(ctx context.Context, d amqp.Delivery) error {
	ctx = h.extractUser(ctx, d)
	var payload DeletePartyPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return h.tm.Run(ctx, func(txCtx context.Context) error {
		party, err := h.repo.GetParty(txCtx, payload.ID)
		if err != nil {
			return err
		}

		if party.Status != string(domain.PartyStatusDeletionPending) {
			return nil
		}

		// Revert to Active
		newStatus := string(domain.PartyStatusActive)

		if party.Type == domain.PartyTypeIndividual {
			ind, _ := h.repo.GetIndividual(txCtx, payload.ID)
			ind.Status = newStatus
			if err := h.repo.UpdateIndividual(txCtx, ind); err != nil {
				return fmt.Errorf("failed to cancel deletion (ind): %w", err)
			}
		} else {
			org, _ := h.repo.GetOrganization(txCtx, payload.ID)
			org.Status = newStatus
			if err := h.repo.UpdateOrganization(txCtx, org); err != nil {
				return fmt.Errorf("failed to cancel deletion (org): %w", err)
			}
		}

		if err := h.publishEvent(txCtx, EvtPartyStateChange, map[string]any{
			"id":       payload.ID,
			"oldState": domain.PartyStatusDeletionPending,
			"newState": newStatus,
		}); err != nil {
			return err
		}

		slog.Info("party deletion cancelled", "id", payload.ID)
		return nil
	})
}

func (h *Handlers) HandleCustomerCreated(ctx context.Context, d amqp.Delivery) error {
	var payload struct {
		ID      string `json:"id"`
		PartyID string `json:"partyId"`
	}
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return err
	}

	if payload.PartyID == "" {
		return nil
	}

	return h.tm.Run(ctx, func(txCtx context.Context) error {
		// Check if party is in deletion pending
		party, err := h.repo.GetParty(txCtx, payload.PartyID)
		if err != nil {
			// If not found, ignore
			return nil
		}

		if party.Status == string(domain.PartyStatusDeletionPending) {
			slog.Info("detecting race condition: customer created for pending deletion party. aborting deletion.", "party_id", payload.PartyID)

			// Revert to Active
			newStatus := string(domain.PartyStatusActive)
			if party.Type == domain.PartyTypeIndividual {
				ind, _ := h.repo.GetIndividual(txCtx, payload.PartyID)
				ind.Status = newStatus
				if err := h.repo.UpdateIndividual(txCtx, ind); err != nil {
					slog.Error("failed to revert party status (ind)", "error", err)
					return err
				}
			} else {
				org, _ := h.repo.GetOrganization(txCtx, payload.PartyID)
				org.Status = newStatus
				if err := h.repo.UpdateOrganization(txCtx, org); err != nil {
					slog.Error("failed to revert party status (org)", "error", err)
					return err
				}
			}

			if err := h.publishEvent(txCtx, EvtPartyStateChange, map[string]any{
				"id":       payload.PartyID,
				"oldState": domain.PartyStatusDeletionPending,
				"newState": newStatus,
			}); err != nil {
				return err
			}
		}
		return nil
	})
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

func (h *Handlers) HandlePurgeParty(ctx context.Context, d amqp.Delivery) error {
	ctx = h.extractUser(ctx, d)
	var payload DeletePartyPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal PurgePartyPayload: %w", err)
	}
	if payload.ID == "" {
		return domain.ErrIDRequired
	}

	party, err := h.repo.GetParty(ctx, payload.ID)
	if err != nil {
		return h.replyTo(ctx, d, map[string]string{"error": err.Error()})
	}
	if party.Status != string(domain.PartyStatusDeleted) {
		return h.replyTo(ctx, d, map[string]string{"error": "party must be in Deleted status to be permanently removed"})
	}

	if err := h.tm.Run(ctx, func(txCtx context.Context) error {
		return h.repo.DeleteParty(txCtx, payload.ID)
	}); err != nil {
		return h.replyTo(ctx, d, map[string]string{"error": err.Error()})
	}

	slog.Info("party permanently deleted", "party_id", payload.ID)
	return h.replyTo(ctx, d, map[string]string{"status": "purged", "id": payload.ID})
}

func (h *Handlers) HandleSearchParty(ctx context.Context, d amqp.Delivery) error {
	ctx = h.extractUser(ctx, d)
	var payload SearchPartyPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal SearchPartyPayload: %w", err)
	}

	criteria := make(map[string]any)
	if payload.Search != nil {
		criteria["search"] = *payload.Search
	}
	if payload.Name != nil {
		criteria["name"] = *payload.Name
	}
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
	if payload.ExternalReference != nil {
		criteria["externalReference"] = *payload.ExternalReference
	}
	if payload.Email != nil {
		criteria["email"] = *payload.Email
	}
	if payload.Status != nil {
		criteria["status"] = *payload.Status
	}

	parties, err := h.repo.SearchParties(ctx, criteria)
	if err != nil {
		return h.replyTo(ctx, d, map[string]string{"error": err.Error()})
	}

	// Fetch full details for each party (Individual or Organization)
	result := make([]any, 0, len(parties))
	for _, party := range parties {
		if party.Type == domain.PartyTypeIndividual {
			ind, err := h.repo.GetIndividual(ctx, party.ID)
			if err != nil {
				slog.Warn("failed to get individual details", "id", party.ID, "error", err)
				result = append(result, party) // fallback to base party
				continue
			}
			result = append(result, ind)
		} else if party.Type == domain.PartyTypeOrganization {
			org, err := h.repo.GetOrganization(ctx, party.ID)
			if err != nil {
				slog.Warn("failed to get organization details", "id", party.ID, "error", err)
				result = append(result, party) // fallback to base party
				continue
			}
			result = append(result, org)
		} else {
			result = append(result, party)
		}
	}

	return h.replyTo(ctx, d, result)
}

// --- Helpers ---

func (h *Handlers) publishEvent(ctx context.Context, routingKey string, event any) error {
	if h.eventPublisher == nil {
		slog.Warn("eventPublisher is nil, skipping event publishing", "routingKey", routingKey)
		return nil
	}
	if err := h.eventPublisher.Publish(ctx, EventExchange, routingKey, event); err != nil {
		slog.Error("failed to publish event", "routingKey", routingKey, "error", err)
		return err
	}
	return nil
}

func (h *Handlers) replyTo(ctx context.Context, d amqp.Delivery, response any) error {
	if d.ReplyTo == "" {
		slog.Debug("no ReplyTo queue specified, skipping reply")
		return nil
	}

	if d.CorrelationId != "" {
		ctx = context.WithValue(ctx, rabbitmq.ContextKeyCorrelationID, d.CorrelationId)
	}

	if h.rpcPublisher == nil {
		slog.Warn("rpcPublisher is nil, cannot reply")
		return nil
	}

	return h.rpcPublisher.Publish(ctx, "", d.ReplyTo, response)
}

func (h *Handlers) extractUser(ctx context.Context, d amqp.Delivery) context.Context {
	if user, ok := d.Headers["user"].(string); ok && user != "" {
		ctx = context.WithValue(ctx, domain.UserContextKey, user)
	}
	if auth, ok := d.Headers["Authorization"].(string); ok && auth != "" {
		ctx = context.WithValue(ctx, domain.AuthContextKey, auth)
	}
	return ctx
}
func (h *Handlers) mapContactMediums(dtos []ContactMediumDTO, partyID string) []domain.ContactMedium {
	res := make([]domain.ContactMedium, 0, len(dtos))
	for _, dto := range dtos {
		id := dto.ID
		if id == "" {
			id = uuid.New().String()
		}
		res = append(res, domain.ContactMedium{
			ID:              id,
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
		id := dto.ID
		if id == "" {
			id = uuid.New().String()
		}
		res = append(res, domain.Identification{
			ID:                 id,
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
			Permissions:      dto.Permissions,
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

func (h *Handlers) mapExternalReferences(dtos []ExternalReferenceDTO, partyID string) []domain.ExternalReference {
	var results []domain.ExternalReference
	for _, dto := range dtos {
		if dto.ID == "" {
			dto.ID = uuid.New().String()
		}
		results = append(results, domain.ExternalReference{
			ID:                  dto.ID,
			PartyID:             partyID,
			ExternalSystemID:    dto.ExternalSystemID,
			ExternalReferenceID: dto.ExternalReferenceID,
			CreatedAt:           time.Now().UTC(),
			UpdatedAt:           time.Now().UTC(),
		})
	}
	return results
}

func (h *Handlers) mapTaxExemptions(dtos []TaxExemptionDTO, partyID string) []domain.TaxExemption {
	var results []domain.TaxExemption
	for _, dto := range dtos {
		if dto.ID == "" {
			dto.ID = uuid.New().String()
		}
		var start, end time.Time
		if dto.ValidForStart != "" {
			start, _ = time.Parse(time.RFC3339, dto.ValidForStart)
		}
		var endPtr *time.Time
		if dto.ValidForEnd != "" {
			end, _ = time.Parse(time.RFC3339, dto.ValidForEnd)
			endPtr = &end
		}

		results = append(results, domain.TaxExemption{
			ID:                  dto.ID,
			PartyID:             partyID,
			CertificateNumber:   dto.CertificateNumber,
			IssuingJurisdiction: dto.IssuingJurisdiction,
			ValidForStart:       start,
			ValidForEnd:         endPtr,
			CreatedAt:           time.Now().UTC(),
			UpdatedAt:           time.Now().UTC(),
		})
	}
	return results
}

func (h *Handlers) mapAttachments(dtos []AttachmentDTO, partyID string) []domain.Attachment {
	var results []domain.Attachment
	for _, dto := range dtos {
		if dto.ID == "" {
			dto.ID = uuid.New().String()
		}

		// If URL is provided and RefType/ID are missing, assume S3
		if dto.RefType == "" {
			if dto.URL != "" {
				dto.RefType = "S3"
				dto.RefID = dto.URL
			} else {
				dto.RefType = "Internal"
			}
		}

		results = append(results, domain.Attachment{
			ID:             dto.ID,
			OwnerID:        partyID,
			Name:           dto.Name,
			MimeType:       dto.MimeType,
			AttachmentType: dto.AttachmentType,
			RefType:        dto.RefType,
			RefID:          dto.RefID,
			ContentData:    dto.Content, // Transferred for repository processing
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		})
	}
	return results
}
