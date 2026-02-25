package cache_test

import (
	"testing"
	"tmf/services/qualification/internal/adapter/cache"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisClient(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		client, err := cache.NewRedisClient("localhost:6379", "", 0)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		
		err = client.Close()
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		_, err := cache.NewRedisClient("invalid-host:6379", "", 0)
		assert.Error(t, err)
	})
}
