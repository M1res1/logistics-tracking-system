package service

import (
	"errors"

	"logistics-tracking-system/services/order/model"
	"logistics-tracking-system/services/order/repository"
)

// CreateOrderItem carries the data for a single line item in a new order.
type CreateOrderItem struct {
	MenuItemID          uint
	Quantity            int
	UnitPrice           float64
	SpecialInstructions string
}

// OrderService implements order business logic.
type OrderService struct {
	repo *repository.OrderRepository
}

// NewOrderService constructs an OrderService.
func NewOrderService(repo *repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

// CreateOrder builds and persists a new order in PENDING status.
// Total is derived from items; tax is 10 % of total; delivery fee is a flat 2.00.
func (s *OrderService) CreateOrder(
	customerID, restaurantID uint,
	items []CreateOrderItem,
	deliveryAddress string,
	lat, lng float64,
) (*model.Order, error) {
	if len(items) == 0 {
		return nil, errors.New("order must contain at least one item")
	}

	var total float64
	orderItems := make([]model.OrderItem, 0, len(items))
	for _, it := range items {
		lineTotal := it.UnitPrice * float64(it.Quantity)
		total += lineTotal
		orderItems = append(orderItems, model.OrderItem{
			MenuItemID:          it.MenuItemID,
			Quantity:            it.Quantity,
			UnitPrice:           it.UnitPrice,
			SpecialInstructions: it.SpecialInstructions,
		})
	}

	tax := total * 0.1
	deliveryFee := 2.0

	order := &model.Order{
		CustomerID:      customerID,
		RestaurantID:    restaurantID,
		Status:          model.StatusPending,
		Total:           total,
		Tax:             tax,
		DeliveryFee:     deliveryFee,
		DeliveryAddress: deliveryAddress,
		DeliveryLat:     lat,
		DeliveryLng:     lng,
		Items:           orderItems,
	}

	if err := s.repo.Create(order); err != nil {
		return nil, err
	}
	return order, nil
}

// GetOrder retrieves an order by ID and enforces customer ownership.
func (s *OrderService) GetOrder(id, customerID uint) (*model.Order, error) {
	order, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if order.CustomerID != customerID {
		return nil, errors.New("order not found")
	}
	return order, nil
}

// ListMyOrders returns a paginated list of orders belonging to customerID.
func (s *OrderService) ListMyOrders(customerID uint, page, limit int) ([]model.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.repo.ListByCustomer(customerID, page, limit)
}

// CancelOrder cancels an order owned by customerID.
// Only PENDING and CONFIRMED orders may be cancelled.
func (s *OrderService) CancelOrder(id, customerID uint) error {
	order, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if order.CustomerID != customerID {
		return errors.New("order not found")
	}
	if !model.CanTransition(order.Status, model.StatusCancelled) {
		return errors.New("order cannot be cancelled in its current state")
	}
	order.Status = model.StatusCancelled
	return s.repo.Update(order)
}

// UpdateStatus transitions an order to newStatus via the state machine.
// This is an internal operation; ownership is not checked here.
func (s *OrderService) UpdateStatus(id uint, newStatus model.OrderStatus) (*model.Order, error) {
	order, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTransition(order.Status, newStatus) {
		return nil, errors.New("invalid status transition")
	}
	order.Status = newStatus
	if err := s.repo.Update(order); err != nil {
		return nil, err
	}
	return order, nil
}
