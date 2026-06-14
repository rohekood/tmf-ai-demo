package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"tmf/services/qualification/internal/core/domain"
	"tmf/services/qualification/internal/core/ports"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type CheckEligibility struct {
	gisClient     ports.GISClient
	invClient     ports.InventoryClient
	catalogClient ports.CatalogClient
	publisher     ports.EventPublisher
	sessionRepo   ports.SessionRepository
	pricingCalc   *PricingCalculator
	logger        *slog.Logger
}

func NewCheckEligibility(
	gis ports.GISClient,
	inv ports.InventoryClient,
	cat ports.CatalogClient,
	pub ports.EventPublisher,
	sessionRepo ports.SessionRepository,
	customerClient ports.CustomerPricingClient,
	catalogPricing ports.CatalogPricingClient,
	logger *slog.Logger,
) *CheckEligibility {
	pricingCalc := NewPricingCalculator(catalogPricing, customerClient)

	return &CheckEligibility{
		gisClient:     gis,
		invClient:     inv,
		catalogClient: cat,
		publisher:     pub,
		sessionRepo:   sessionRepo,
		pricingCalc:   pricingCalc,
		logger:        logger,
	}
}

func (uc *CheckEligibility) Execute(ctx context.Context, cmd domain.CheckEligibilityCommand) error {
	logger := uc.logger.With("correlation_id", cmd.CorrelationID, "address", fmt.Sprintf("%s %s", cmd.Address.Street, cmd.Address.Number))
	logger.Info("Starting eligibility check")

	// Create a child context with timeout if not already present (failsafe)
	// Ideally the caller sets the timeout, but we enforce a max here.
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	g, groupCtx := errgroup.WithContext(ctx)

	// Results containers
	var (
		inPolygon     bool
		freePorts     int
		catalogOffers []domain.EligibleCategory
		mu            sync.Mutex // Protects catalogOffers if appended concurrently (though here we just assign)
	)

	// 1. Scatter: Async Queries

	// Query GIS
	g.Go(func() error {
		var err error
		inPolygon, err = uc.gisClient.CheckPolygon(groupCtx, cmd.Address)
		if err != nil {
			logger.Error("GIS Check failed", "error", err)
			return err
		}
		return nil
	})

	// Query Inventory
	g.Go(func() error {
		var err error
		freePorts, err = uc.invClient.GetPortCapacity(groupCtx, cmd.Address)
		if err != nil {
			logger.Error("Inventory Check failed", "error", err)
			return err
		}
		return nil
	})

	// Query Catalog
	g.Go(func() error {
		var err error
		// Logic: If filter is provided, use it. Else default to "Fiber".
		category := "Fiber"
		if len(cmd.CategoryFilter) > 0 {
			category = cmd.CategoryFilter[0] // Simplify for now
		}

		offers, err := uc.catalogClient.GetOffersByCategory(groupCtx, category)
		if err != nil {
			logger.Error("Catalog Check failed", "error", err)
			return err
		}

		mu.Lock()
		catalogOffers = offers
		mu.Unlock()
		return nil
	})

	// Wait for gathered results
	if err := g.Wait(); err != nil {
		logger.Error("Scatter-Gather failed", "error", err)
		result := domain.EligibilityResult{
			Status:               domain.StatusError,
			UnavailabilityReason: fmt.Sprintf("System Dependency Failed: %v", err),
		}
		if cmd.SessionID != "" {
			_, _ = uc.sessionRepo.Create(ctx, &domain.QualificationSession{
				ID:        cmd.SessionID,
				CustomerID: cmd.CustomerID,
				Address:   cmd.Address,
				Status:    string(domain.StatusError),
				CreatedAt: time.Now().UTC(),
				ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
			})
		}
		return uc.publishResult(ctx, cmd, result)
	}

	// 2. Logic: Rule Engine
	isQualified := false
	var eligible []domain.EligibleCategory
	var reason string

	if inPolygon {
		if freePorts > 0 {
			isQualified = true
			eligible = catalogOffers
		} else {
			reason = "No network capacity available"
		}
	} else {
		reason = "Address outside service area"
	}

	status := domain.StatusQualified
	if !isQualified {
		status = domain.StatusUnqualified
	}

	logger.Info("Eligibility determined", "status", status, "reason", reason)

	var sessionID string
	var qualifiedOffers []domain.QualifiedOffer

	if isQualified {
		// Authenticated callers (CustomerID set) get segment/tier pricing;
		// anonymous callers get the generic catalog base price.
		anonymous := cmd.CustomerID == ""
		for _, category := range eligible {
			var price *domain.Price
			var err error
			priceType := domain.PriceTypeCustomer
			if anonymous {
				priceType = domain.PriceTypeGeneric
				price, err = uc.pricingCalc.CalculateGenericPrice(ctx, category.ID)
			} else {
				price, err = uc.pricingCalc.CalculatePrice(ctx, category.ID, cmd.CustomerID)
			}
			if err != nil {
				logger.Warn("Failed to calculate price", "offeringId", category.ID, "priceType", priceType, "error", err)
				continue
			}
			qualifiedOffers = append(qualifiedOffers, domain.QualifiedOffer{
				OfferingID:   category.ID,
				OfferingName: category.Name,
				Price:        *price,
				PriceType:    priceType,
				Eligibility:  "QUALIFIED",
			})
		}
	}

	// Always persist the session so async callers can poll for the result.
	// Use cmd.SessionID if the caller pre-generated one (async flow).
	session := &domain.QualificationSession{
		ID:              cmd.SessionID,
		CustomerID:      cmd.CustomerID,
		Address:         cmd.Address,
		QualifiedOffers: qualifiedOffers,
		Status:          string(status),
		CreatedAt:       time.Now().UTC(),
		ExpiresAt:       time.Now().UTC().Add(24 * time.Hour),
	}
	var err error
	sessionID, err = uc.sessionRepo.Create(ctx, session)
	if err != nil {
		logger.Error("Failed to create session", "error", err)
	} else {
		logger.Info("Created qualification session", "sessionId", sessionID)
	}

	result := domain.EligibilityResult{
		Status:               status,
		SessionID:            sessionID,
		EligibleCategories:   eligible,
		QualifiedOffers:      qualifiedOffers,
		UnavailabilityReason: reason,
	}

	// 3. Publish Result Event
	return uc.publishResult(ctx, cmd, result)
}

func (uc *CheckEligibility) publishResult(ctx context.Context, cmd domain.CheckEligibilityCommand, result domain.EligibilityResult) error {
	// Enrich result
	result.QualificationID = uuid.New().String()
	result.CorrelationID = cmd.CorrelationID

	return uc.publisher.PublishEligibilityChecked(ctx, result)
}
