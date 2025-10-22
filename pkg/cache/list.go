package cache

import (
	"context"
	"encoding/json"
	"fmt"
)

// List Operations

// LPush thêm vào đầu list
func (c *redisCache) LPush(ctx context.Context, key string, values ...interface{}) error {
	// Convert values
	convertedValues := make([]interface{}, len(values))
	for i, v := range values {
		switch val := v.(type) {
		case string, []byte, int, int64, float64:
			convertedValues[i] = val
		default:
			data, err := json.Marshal(val)
			if err != nil {
				return fmt.Errorf("failed to marshal value: %w", err)
			}
			convertedValues[i] = string(data)
		}
	}

	return c.client.LPush(ctx, key, convertedValues...).Err()
}

// RPush thêm vào cuối list
func (c *redisCache) RPush(ctx context.Context, key string, values ...interface{}) error {
	convertedValues := make([]interface{}, len(values))
	for i, v := range values {
		switch val := v.(type) {
		case string, []byte, int, int64, float64:
			convertedValues[i] = val
		default:
			data, err := json.Marshal(val)
			if err != nil {
				return fmt.Errorf("failed to marshal value: %w", err)
			}
			convertedValues[i] = string(data)
		}
	}

	return c.client.RPush(ctx, key, convertedValues...).Err()
}

// LPop lấy và xóa phần tử đầu list
func (c *redisCache) LPop(ctx context.Context, key string) (string, error) {
	return c.client.LPop(ctx, key).Result()
}

// RPop lấy và xóa phần tử cuối list
func (c *redisCache) RPop(ctx context.Context, key string) (string, error) {
	return c.client.RPop(ctx, key).Result()
}

// LRange lấy range của list
func (c *redisCache) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.LRange(ctx, key, start, stop).Result()
}
