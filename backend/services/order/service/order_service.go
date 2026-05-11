package service

import (
	"context"
	"errors"
	"fmt"

	"food-delivery/pkg/kafka"
	"food-delivery/services/order/model"
	"food-delivery/services/order/repository"
)

// TODO: move these to config later
const (
	taxRate     = 0.10
	deliveryFee = 3.99
)

var (
	ErrOrderNotFound     = errors.New("order not found")
	ErrForbidden         = errors.New("access denied")
	ErrCannotCancel      = errors.New("order cannot be cancelled at this stage")
	ErrInvalidTransition = errors.New("invalid status transition")
)

type CreateOrderRequest struct {
	RestaurantID    uint             `json:"restaurant_id" binding:"required"`
	DeliveryAddress string           `json:"delivery_address" binding:"required"`
	DeliveryLat     float64          `json:"delivery_lat"`
	DeliveryLng     float64          `json:"delivery_lng"`
	Items           []OrderItemInput `json:"items" binding:"required,min=1,dive"`
}

type OrderItemInput struct {
	MenuItemID          uint    `json:"menu_item_id" binding:"required"`
	Quantity            int     `json:"quantity" binding:"required,min=1"`
	UnitPrice           float64 `json:"unit_price" binding:"required,gt=0"`
	SpecialInstructions string  `json:"special_instructions"`
}

type UpdateStatusRequest struct {
	Status model.Status `json:"status" binding:"required"`
}

type OrderService interface {
	CreateOrder(customerID uint, req CreateOrderRequest) (*model.Order, error)
	GetOrder(orderID, customerID uint) (*model.Order, error)
	ListMyOrders(customerID uint, page, limit int) ([]model.Order, int64, error)
	CancelOrder(orderID, customerID uint) (*model.Order, error)
	UpdateStatus(orderID uint, newStatus model.Status) (*model.Order, error)
}

type orderService struct {
	repo           repository.OrderRepo
	orderProducer  *kafka.Producer
	statusProducer *kafka.Producer
}

func NewOrderService(repo repository.OrderRepo, orderProducer, statusProducer *kafka.Producer) OrderService {
	return &orderService{repo: repo, orderProducer: orderProducer, statusProducer: statusProducer}
}

func (s *orderService) CreateOrder(customerID uint, req CreateOrderRequest) (*model.Order, error) {
	// TODO: validate that each menu_item_id exists by calling restaurant service
	// for now we trust what the client sends — need to fix before prod

	var subtotal float64
	items := make([]model.OrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		subtotal += item.UnitPrice * float64(item.Quantity)
		items = append(items, model.OrderItem{
			MenuItemID:          item.MenuItemID,
			Quantity:            item.Quantity,
			UnitPrice:           item.UnitPrice,
			SpecialInstructions: item.SpecialInstructions,
		})
	}

	tax := subtotal * taxRate
	order := &model.Order{
		CustomerID:      customerID,
		RestaurantID:    req.RestaurantID,
		Status:          model.StatusPending,
		Subtotal:        subtotal,
		Tax:             tax,
		DeliveryFee:     deliveryFee,
		Total:           subtotal + tax + deliveryFee,
		DeliveryAddress: req.DeliveryAddress,
		DeliveryLat:     req.DeliveryLat,
		DeliveryLng:     req.DeliveryLng,
		Items:           items,
	}

	if err := s.repo.Create(order); err != nil {
		return nil, err
	}

	// fire kafka event — if broker is down just skip, not critical for v1
	if s.orderProducer != nil {
		_ = s.orderProducer.Publish(context.Background(), fmt.Sprintf("%d", order.ID), map[string]interface{}{
			"event":         "order.created",
			"order_id":      order.ID,
			"customer_id":   order.CustomerID,
			"restaurant_id": order.RestaurantID,
			"total":         order.Total,
		})
	}

	return order, nil
}

func (s *orderService) GetOrder(orderID, customerID uint) (*model.Order, error) {
	order, err := s.repo.FindByID(orderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}
	if order.CustomerID != customerID {
		return nil, ErrForbidden
	}
	return order, nil
}

func (s *orderService) ListMyOrders(customerID uint, page, limit int) ([]model.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.FindByCustomer(customerID, page, limit)
}

func (s *orderService) CancelOrder(orderID, customerID uint) (*model.Order, error) {
	order, err := s.repo.FindByID(orderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}
	if order.CustomerID != customerID {
		return nil, ErrForbidden
	}
	if !model.CanTransition(order.Status, model.StatusCancelled) {
		return nil, ErrCannotCancel
	}

	oldStatus := order.Status
	if err := s.repo.UpdateStatus(orderID, model.StatusCancelled); err != nil {
		return nil, err
	}
	order.Status = model.StatusCancelled

	if s.statusProducer != nil {
		_ = s.statusProducer.Publish(context.Background(), fmt.Sprintf("%d", order.ID), map[string]interface{}{
			"event":      "order.status_changed",
			"order_id":   order.ID,
			"old_status": oldStatus,
			"new_status": model.StatusCancelled,
		})
	}

	return order, nil
}

func (s *orderService) UpdateStatus(orderID uint, newStatus model.Status) (*model.Order, error) {
	order, err := s.repo.FindByID(orderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}
	if !model.CanTransition(order.Status, newStatus) {
		return nil, ErrInvalidTransition
	}

	oldStatus := order.Status
	if err := s.repo.UpdateStatus(orderID, newStatus); err != nil {
		return nil, err
	}
	order.Status = newStatus

	if s.statusProducer != nil {
		_ = s.statusProducer.Publish(context.Background(), fmt.Sprintf("%d", order.ID), map[string]interface{}{
			"event":      "order.status_changed",
			"order_id":   order.ID,
			"old_status": oldStatus,
			"new_status": newStatus,
		})
	}

	return order, nil
}
