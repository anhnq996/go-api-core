package cache

import (
	"context"
	"encoding/json"
	"fmt"
)

// Hash Operations

// HSet set hash field value
func (c *redisCache) HSet(ctx context.Context, key string, field string, value interface{}) error {
	// Convert value to string
	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	case []byte:
		strValue = string(v)
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		strValue = string(data)
	}

	return c.client.HSet(ctx, key, field, strValue).Err()
}

// HGet get hash field value
func (c *redisCache) HGet(ctx context.Context, key string, field string) (string, error) {
	return c.client.HGet(ctx, key, field).Result()
}

// HGetAll lấy tất cả fields trong hash
func (c *redisCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}

// HDel xóa hash fields
func (c *redisCache) HDel(ctx context.Context, key string, fields ...string) error {
	return c.client.HDel(ctx, key, fields...).Err()
}

// HExists kiểm tra field tồn tại trong hash
func (c *redisCache) HExists(ctx context.Context, key string, field string) (bool, error) {
	return c.client.HExists(ctx, key, field).Result()
}
