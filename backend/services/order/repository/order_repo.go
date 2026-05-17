package repository

import (
	"logistics-tracking-system/services/order/model"

	"gorm.io/gorm"
)

// OrderRepository handles persistence for orders.
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository constructs an OrderRepository backed by the given DB.
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create persists a new order (and its items, via GORM associations).
func (r *OrderRepository) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

// GetByID retrieves an order by primary key, preloading its items.
func (r *OrderRepository) GetByID(id uint) (*model.Order, error) {
	var order model.Order
	if err := r.db.Preload("Items").First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// ListByCustomer returns a paginated list of orders for a given customer.
// It also returns the total count of matching rows (for pagination metadata).
func (r *OrderRepository) ListByCustomer(customerID uint, page, limit int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	offset := (page - 1) * limit

	if err := r.db.Model(&model.Order{}).
		Where("customer_id = ?", customerID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Where("customer_id = ?", customerID).
		Preload("Items").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// Update saves changes to an existing order record.
func (r *OrderRepository) Update(order *model.Order) error {
	return r.db.Save(order).Error
}
