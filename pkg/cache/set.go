package cache

import (
	"context"
	"encoding/json"
	"fmt"
)

// Set Operations

// SAdd thêm members vào set
func (c *redisCache) SAdd(ctx context.Context, key string, members ...interface{}) error {
	// Convert members to interface{}
	values := make([]interface{}, len(members))
	for i, member := range members {
		switch v := member.(type) {
		case string, []byte, int, int64, float64, bool:
			values[i] = v
		default:
			data, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf("failed to marshal member: %w", err)
			}
			values[i] = string(data)
		}
	}

	return c.client.SAdd(ctx, key, values...).Err()
}

// SRem xóa members khỏi set
func (c *redisCache) SRem(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SRem(ctx, key, members...).Err()
}

// SMembers lấy tất cả members trong set
func (c *redisCache) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.client.SMembers(ctx, key).Result()
}

// SIsMember kiểm tra member có trong set không
func (c *redisCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.client.SIsMember(ctx, key, member).Result()
}

// SCard lấy số lượng members trong set
func (c *redisCache) SCard(ctx context.Context, key string) (int64, error) {
	return c.client.SCard(ctx, key).Result()
}
