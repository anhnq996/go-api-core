package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"api-core/pkg/actionEvent"
	"api-core/pkg/jwt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository interface định nghĩa các CRUD operations cơ bản
type Repository[T any] interface {
	Create(ctx context.Context, entity *T) error
	FindAll(ctx context.Context) ([]T, error)
	FindByID(ctx context.Context, id uuid.UUID) (*T, error)
	Update(ctx context.Context, id uuid.UUID, entity *T) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)

	// Query builders
	FindWhere(ctx context.Context, condition string, args ...interface{}) ([]T, error)
	FirstWhere(ctx context.Context, condition string, args ...interface{}) (*T, error)
	UpdateWhere(ctx context.Context, condition string, updates map[string]interface{}, args ...interface{}) error
	DeleteWhere(ctx context.Context, condition string, args ...interface{}) error

	// Pagination
	Paginate(ctx context.Context, page, perPage int) ([]T, int64, error)
	FindWithPagination(ctx context.Context, page, perPage int, sort, order, search string, searchFields []string) ([]T, int64, error)

	// Bulk operations
	BulkCreate(ctx context.Context, entities []T) error

	// Database access
	DB() *gorm.DB
	WithPreload(associations ...string) *gorm.DB
}

// BaseRepository implementation với generics
type BaseRepository[T any] struct {
	db          *gorm.DB
	entityName  string
	actionEvent bool // Flag to enable/disable action event logging
}

// NewBaseRepository khởi tạo BaseRepository
func NewBaseRepository[T any](db *gorm.DB, actionEvent bool) *BaseRepository[T] {
	var entity T
	entityName := fmt.Sprintf("%T", entity)
	// Remove package prefix if exists
	if idx := strings.LastIndex(entityName, "."); idx != -1 {
		entityName = entityName[idx+1:]
	}

	return &BaseRepository[T]{
		db:          db,
		entityName:  strings.ToLower(entityName),
		actionEvent: actionEvent,
	}
}

// Create tạo entity mới
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	err := r.db.WithContext(ctx).Create(entity).Error
	if err == nil && r.actionEvent {
		// Extract ID from entity for logging
		entityID := r.extractEntityID(entity)
		userID := r.extractUserIDFromContext(ctx)

		// Log create event with dynamic job
		job := r.getJobName("action_events")
		newData := r.convertEntityToMap(entity)
		actionEvent.LogCreate(ctx, job, r.entityName, entityID, userID, newData)
	}
	return err
}

// FindAll lấy tất cả entities
func (r *BaseRepository[T]) FindAll(ctx context.Context) ([]T, error) {
	var entities []T
	err := r.db.WithContext(ctx).Find(&entities).Error
	return entities, err
}

// FindByID tìm entity theo ID
func (r *BaseRepository[T]) FindByID(ctx context.Context, id uuid.UUID) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// Update cập nhật entity
func (r *BaseRepository[T]) Update(ctx context.Context, id uuid.UUID, entity *T) error {
	// Get old data before update
	var oldEntity T
	oldErr := r.db.WithContext(ctx).First(&oldEntity, "id = ?", id).Error

	err := r.db.WithContext(ctx).Model(entity).Where("id = ?", id).Updates(entity).Error
	if err == nil && r.actionEvent {
		userID := r.extractUserIDFromContext(ctx)

		// Log update event with dynamic job
		job := r.getJobName("action_events")
		var oldData map[string]interface{}
		if oldErr == nil {
			oldData = r.convertEntityToMap(&oldEntity)
		}
		newData := r.convertEntityToMap(entity)
		actionEvent.LogUpdate(ctx, job, r.entityName, id.String(), userID, oldData, newData)
	}
	return err
}

// Delete xóa entity (soft delete nếu model có DeletedAt)
func (r *BaseRepository[T]) Delete(ctx context.Context, id uuid.UUID) error {
	var entity T
	err := r.db.WithContext(ctx).Delete(&entity, "id = ?", id).Error
	if err == nil && r.actionEvent {
		userID := r.extractUserIDFromContext(ctx)

		// Log delete event with dynamic job
		job := r.getJobName("action_events")
		data := r.convertEntityToMap(&entity)
		actionEvent.LogDelete(ctx, job, r.entityName, id.String(), userID, data)
	}
	return err
}

// Count đếm tổng số entities
func (r *BaseRepository[T]) Count(ctx context.Context) (int64, error) {
	var count int64
	var entity T
	err := r.db.WithContext(ctx).Model(&entity).Count(&count).Error
	return count, err
}

// Exists kiểm tra entity có tồn tại không
func (r *BaseRepository[T]) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var entity T
	err := r.db.WithContext(ctx).Select("id").First(&entity, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// FindWhere tìm entities theo điều kiện
func (r *BaseRepository[T]) FindWhere(ctx context.Context, condition string, args ...interface{}) ([]T, error) {
	var entities []T
	err := r.db.WithContext(ctx).Where(condition, args...).Find(&entities).Error
	return entities, err
}

// FirstWhere tìm entity đầu tiên theo điều kiện
func (r *BaseRepository[T]) FirstWhere(ctx context.Context, condition string, args ...interface{}) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).Where(condition, args...).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// UpdateWhere cập nhật theo điều kiện
func (r *BaseRepository[T]) UpdateWhere(ctx context.Context, condition string, updates map[string]interface{}, args ...interface{}) error {
	var entity T
	return r.db.WithContext(ctx).Model(&entity).Where(condition, args...).Updates(updates).Error
}

// DeleteWhere xóa theo điều kiện
func (r *BaseRepository[T]) DeleteWhere(ctx context.Context, condition string, args ...interface{}) error {
	var entity T
	return r.db.WithContext(ctx).Where(condition, args...).Delete(&entity).Error
}

// Paginate phân trang
func (r *BaseRepository[T]) Paginate(ctx context.Context, page, perPage int) ([]T, int64, error) {
	var entities []T
	var total int64

	var entity T
	if err := r.db.WithContext(ctx).Model(&entity).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	err := r.db.WithContext(ctx).Offset(offset).Limit(perPage).Find(&entities).Error

	return entities, total, err
}

// FindWithPagination phân trang với sort, order và search
func (r *BaseRepository[T]) FindWithPagination(ctx context.Context, page, perPage int, sort, order, search string, searchFields []string) ([]T, int64, error) {
	var entities []T
	var total int64

	// Set defaults
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}
	if order == "" {
		order = "asc"
	}
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	// Build query
	query := r.db.WithContext(ctx).Model(new(T))

	// Add search condition
	if search != "" && len(searchFields) > 0 {
		var conditions []string
		var args []interface{}

		for _, field := range searchFields {
			conditions = append(conditions, field+" ILIKE ?")
			args = append(args, "%"+search+"%")
		}

		if len(conditions) > 0 {
			query = query.Where("("+strings.Join(conditions, " OR ")+")", args...)
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Add sorting (chỉ sort nếu có truyền sort field)
	if sort != "" {
		sortField := sort
		if order == "desc" {
			sortField = sort + " DESC"
		}
		query = query.Order(sortField)
	}

	// Add pagination and execute
	offset := (page - 1) * perPage
	err := query.Offset(offset).Limit(perPage).Find(&entities).Error

	return entities, total, err
}

// BulkCreate tạo nhiều entities
func (r *BaseRepository[T]) BulkCreate(ctx context.Context, entities []T) error {
	return r.db.WithContext(ctx).Create(&entities).Error
}

// DB trả về database instance
func (r *BaseRepository[T]) DB() *gorm.DB {
	return r.db
}

// WithPreload preload associations
func (r *BaseRepository[T]) WithPreload(associations ...string) *gorm.DB {
	db := r.db
	for _, assoc := range associations {
		db = db.Preload(assoc)
	}
	return db
}

// Transaction helper
func (r *BaseRepository[T]) Transaction(fn func(*gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// NotFoundError tạo error message chuẩn
func (r *BaseRepository[T]) NotFoundError(id uuid.UUID) error {
	var entity T
	return fmt.Errorf("%T with id %s not found", entity, id)
}

// Helper methods for event logging

// extractEntityID extracts ID from entity using reflection
func (r *BaseRepository[T]) extractEntityID(entity *T) string {
	// Use reflection to get ID field
	// This is a simplified version - you might want to use a more robust approach
	// For now, we'll assume entities have an ID field
	// In a real implementation, you might want to use reflection or interface{} approach

	// For now, return empty string and let the caller handle it
	// This is a placeholder - you can implement proper ID extraction based on your entity structure
	return ""
}

// extractUserIDFromContext extracts user ID from context
func (r *BaseRepository[T]) extractUserIDFromContext(ctx context.Context) string {
	// Use JWT package's function to get user ID from context
	return jwt.GetUserIDFromContext(ctx)
}

// convertEntityToMap converts entity to map[string]interface{}
func (r *BaseRepository[T]) convertEntityToMap(entity *T) map[string]interface{} {
	// Use JSON marshaling/unmarshaling to convert struct to map
	// This is a simple approach that works for most cases
	data := make(map[string]interface{})

	// Convert to JSON first
	jsonData, err := json.Marshal(entity)
	if err != nil {
		return map[string]interface{}{
			"error": "failed to convert entity to map",
		}
	}

	// Unmarshal to map
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return map[string]interface{}{
			"error": "failed to convert entity to map",
		}
	}

	return data
}

// getJobName creates dynamic job name based on entity and action
func (r *BaseRepository[T]) getJobName(action string) string {
	// You can customize this logic based on your needs
	// Examples:
	// - "user_create", "user_update", "user_delete"
	// - "action_events" (default)
	// - "crud_events"

	// For now, return a dynamic job name
	return action
}
