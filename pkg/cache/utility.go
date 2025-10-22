package cache

import "context"

// Utility Operations

// Ping kiểm tra Redis connection
func (c *redisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// FlushDB xóa tất cả keys trong DB hiện tại
func (c *redisCache) FlushDB(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

// Close đóng Redis connection
func (c *redisCache) Close() error {
	return c.client.Close()
}
