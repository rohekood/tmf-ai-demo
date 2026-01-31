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
	logger        *slog.Logger
}

func NewCheckEligibility(
	gis ports.GISClient,
	inv ports.InventoryClient,
	cat ports.CatalogClient,
	pub ports.EventPublisher,
	logger *slog.Logger,
) *CheckEligibility {
	return &CheckEligibility{
		gisClient:     gis,
		invClient:     inv,
		catalogClient: cat,
		publisher:     pub,
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

	g, ctx := errgroup.WithContext(ctx)

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
		inPolygon, err = uc.gisClient.CheckPolygon(ctx, cmd.Address)
		if err != nil {
			logger.Error("GIS Check failed", "error", err)
			return err
		}
		return nil
	})

	// Query Inventory
	g.Go(func() error {
		var err error
		freePorts, err = uc.invClient.GetPortCapacity(ctx, cmd.Address)
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

		offers, err := uc.catalogClient.GetOffersByCategory(ctx, category)
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
		// If any critical dependency failed, we publish an Error event
		logger.Error("Scatter-Gather failed", "error", err)
		return uc.publishResult(ctx, cmd, domain.EligibilityResult{
			Status:               domain.StatusError,
			UnavailabilityReason: fmt.Sprintf("System Dependency Failed: %v", err),
		})
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

	result := domain.EligibilityResult{
		Status:               status,
		EligibleCategories:   eligible,
		UnavailabilityReason: reason,
	}

	logger.Info("Eligibility determined", "status", status, "reason", reason)

	// 3. Publish Result Event
	return uc.publishResult(ctx, cmd, result)
}

func (uc *CheckEligibility) publishResult(ctx context.Context, cmd domain.CheckEligibilityCommand, result domain.EligibilityResult) error {
	// Enrich result
	result.QualificationID = uuid.New().String()
	result.CorrelationID = cmd.CorrelationID

	return uc.publisher.PublishEligibilityChecked(ctx, result)
}
