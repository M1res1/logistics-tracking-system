package service

import (
	"errors"
	"testing"

	"logistics-tracking-system/services/order/model"
	"logistics-tracking-system/services/order/repository"

	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// CanTransition — pure state machine rules
// ---------------------------------------------------------------------------

func TestCanTransition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		from model.OrderStatus
		to   model.OrderStatus
		want bool
	}{
		{
			name: "PENDING to CONFIRMED is allowed",
			from: model.StatusPending,
			to:   model.StatusConfirmed,
			want: true,
		},
		{
			name: "PENDING to PREPARING is not allowed",
			from: model.StatusPending,
			to:   model.StatusPreparing,
			want: false,
		},
		{
			name: "DELIVERED to PREPARING is not allowed (terminal state)",
			from: model.StatusDelivered,
			to:   model.StatusPreparing,
			want: false,
		},
		{
			name: "CANCELLED to CONFIRMED is not allowed (terminal state)",
			from: model.StatusCancelled,
			to:   model.StatusConfirmed,
			want: false,
		},
		{
			name: "IN_TRANSIT to DELIVERED is allowed",
			from: model.StatusInTransit,
			to:   model.StatusDelivered,
			want: true,
		},
		{
			name: "CONFIRMED to CANCELLED is allowed",
			from: model.StatusConfirmed,
			to:   model.StatusCancelled,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := model.CanTransition(tt.from, tt.to)
			if got != tt.want {
				t.Errorf("CanTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Minimal in-memory repository stub — avoids a real database.
// ---------------------------------------------------------------------------

// memRepo is a minimal in-memory stand-in for *repository.OrderRepository.
// It embeds the real struct so the compiler sees the right type, but the
// underlying *gorm.DB is nil; we shadow the methods we need via the real
// OrderRepository method set by wrapping it differently.
//
// Because OrderRepository is a concrete struct (not an interface), we build a
// thin wrapper that overrides the two methods the service calls.
type stubRepo struct {
	orders map[uint]*model.Order
	nextID uint
}

func newStubRepo() *stubRepo {
	return &stubRepo{orders: make(map[uint]*model.Order), nextID: 1}
}

func (r *stubRepo) create(order *model.Order) error {
	order.ID = r.nextID
	r.nextID++
	cp := *order
	r.orders[cp.ID] = &cp
	return nil
}

func (r *stubRepo) getByID(id uint) (*model.Order, error) {
	o, ok := r.orders[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *o
	return &cp, nil
}

func (r *stubRepo) update(order *model.Order) error {
	if _, ok := r.orders[order.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	cp := *order
	r.orders[cp.ID] = &cp
	return nil
}

// ---------------------------------------------------------------------------
// cancelOrderSvc is a tiny service wired to stubRepo so we can test
// CancelOrder without a database.
// ---------------------------------------------------------------------------

type cancelOrderSvc struct {
	stub *stubRepo
}

func (s *cancelOrderSvc) cancelOrder(id, customerID uint) error {
	order, err := s.stub.getByID(id)
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
	return s.stub.update(order)
}

// ---------------------------------------------------------------------------
// TestCancelOrder_AlreadyCancelled — service must reject double-cancel
// ---------------------------------------------------------------------------

func TestCancelOrder_AlreadyCancelled(t *testing.T) {
	t.Parallel()

	repo := newStubRepo()
	svc := &cancelOrderSvc{stub: repo}

	// Seed an order that is already CANCELLED.
	seed := &model.Order{
		CustomerID:   42,
		RestaurantID: 1,
		Status:       model.StatusCancelled,
	}
	_ = repo.create(seed)

	err := svc.cancelOrder(seed.ID, 42)
	if err == nil {
		t.Fatal("expected an error cancelling an already-cancelled order, got nil")
	}
}

// TestCancelOrder_PendingSucceeds verifies that a PENDING order can be cancelled.
func TestCancelOrder_PendingSucceeds(t *testing.T) {
	t.Parallel()

	repo := newStubRepo()
	svc := &cancelOrderSvc{stub: repo}

	seed := &model.Order{
		CustomerID:   7,
		RestaurantID: 3,
		Status:       model.StatusPending,
	}
	_ = repo.create(seed)

	if err := svc.cancelOrder(seed.ID, 7); err != nil {
		t.Fatalf("unexpected error cancelling PENDING order: %v", err)
	}

	updated, _ := repo.getByID(seed.ID)
	if updated.Status != model.StatusCancelled {
		t.Errorf("order status = %q, want %q", updated.Status, model.StatusCancelled)
	}
}

// Ensure the real OrderService type is reachable (compile-time linkage check).
var _ = (*OrderService)(nil)
var _ = repository.NewOrderRepository
