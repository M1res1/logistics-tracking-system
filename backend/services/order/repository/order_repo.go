package repository

import (
	"food-delivery/services/order/model"

	"gorm.io/gorm"
)

type OrderRepo interface {
	Create(order *model.Order) error
	FindByID(id uint) (*model.Order, error)
	FindByCustomer(customerID uint, page, limit int) ([]model.Order, int64, error)
	UpdateStatus(id uint, status model.Status) error
}

type orderRepo struct {
	db *gorm.DB
}

func NewOrderRepo(db *gorm.DB) OrderRepo {
	return &orderRepo{db: db}
}

func (r *orderRepo) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepo) FindByID(id uint) (*model.Order, error) {
	var o model.Order
	err := r.db.Preload("Items").First(&o, id).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepo) FindByCustomer(customerID uint, page, limit int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	if err := r.db.Model(&model.Order{}).Where("customer_id = ?", customerID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Where("customer_id = ?", customerID).
		Preload("Items").
		Order("created_at desc").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&orders).Error

	return orders, total, err
}

func (r *orderRepo) UpdateStatus(id uint, status model.Status) error {
	return r.db.Model(&model.Order{}).Where("id = ?", id).Update("status", status).Error
}
