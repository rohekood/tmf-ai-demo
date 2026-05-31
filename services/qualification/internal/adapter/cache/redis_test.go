package cache_test

import (
	"context"
	"testing"
	"tmf/services/qualification/internal/adapter/cache"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestNewRedisClient(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()

		redisContainer, err := redis.Run(ctx,
			"redis:7",
		)
		if err != nil {
			t.Skipf("Skipping integration test (testcontainers error): %v", err)
		}
		defer func() { _ = redisContainer.Terminate(ctx) }()

		addr, err := redisContainer.Endpoint(ctx, "")
		assert.NoError(t, err)

		client, err := cache.NewRedisClient(addr, "", "")
		assert.NoError(t, err)
		assert.NotNil(t, client)

		err = client.Close()
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		_, err := cache.NewRedisClient("invalid-host:6379", "", "")
		assert.Error(t, err)
	})
}
