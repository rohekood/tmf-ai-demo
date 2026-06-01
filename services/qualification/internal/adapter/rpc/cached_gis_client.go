package rpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"tmf/services/qualification/internal/core/domain"
	"tmf/services/qualification/internal/core/ports"

	"github.com/redis/go-redis/v9"
)

// CachedGISClient decorates a GISClient with Redis caching
type CachedGISClient struct {
	next      ports.GISClient
	redis     *redis.Client
	logger    *slog.Logger
	ttl       time.Duration
	keyPrefix string
}

func NewCachedGISClient(next ports.GISClient, rdb *redis.Client, logger *slog.Logger, keyPrefix string) *CachedGISClient {
	return &CachedGISClient{
		next:      next,
		redis:     rdb,
		logger:    logger,
		ttl:       24 * time.Hour,
		keyPrefix: keyPrefix,
	}
}

func (c *CachedGISClient) CheckPolygon(ctx context.Context, addr domain.Address) (bool, error) {
	// Normalize Key (Composite of Zip+City+Street+Number)
	key := fmt.Sprintf("%sgis:polygon:%s:%s:%s:%s", c.keyPrefix, addr.Zip, addr.City, addr.Street, addr.Number)

	// Try Cache
	val, err := c.redis.Get(ctx, key).Result()
	if err == nil {
		// Hit
		c.logger.Info("GIS Cache Hit", "key", key, "val", val)
		return val == "1", nil
	}
	if err != redis.Nil {
		c.logger.Warn("Redis read failed", "error", err)
	}

	// Miss -> Call Source
	exists, err := c.next.CheckPolygon(ctx, addr)
	if err != nil {
		return false, err
	}

	// Set Cache
	strVal := "0"
	if exists {
		strVal = "1"
	}
	if err := c.redis.Set(ctx, key, strVal, c.ttl).Err(); err != nil {
		c.logger.Warn("Redis write failed", "error", err)
	}

	return exists, nil
}
