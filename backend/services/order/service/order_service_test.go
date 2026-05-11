package service_test

import (
	"errors"
	"testing"

	"food-delivery/services/order/model"
	"food-delivery/services/order/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---- state machine tests ----

func TestCanTransition_HappyPath(t *testing.T) {
	assert.True(t, model.CanTransition(model.StatusPending, model.StatusConfirmed))
	assert.True(t, model.CanTransition(model.StatusConfirmed, model.StatusPreparing))
	assert.True(t, model.CanTransition(model.StatusPreparing, model.StatusReady))
	assert.True(t, model.CanTransition(model.StatusReady, model.StatusAssigned))
	assert.True(t, model.CanTransition(model.StatusAssigned, model.StatusInTransit))
	assert.True(t, model.CanTransition(model.StatusInTransit, model.StatusDelivered))
}

func TestCanTransition_CancelAllowedOnlyEarly(t *testing.T) {
	assert.True(t, model.CanTransition(model.StatusPending, model.StatusCancelled))
	assert.True(t, model.CanTransition(model.StatusConfirmed, model.StatusCancelled))

	// can't cancel once kitchen started
	assert.False(t, model.CanTransition(model.StatusPreparing, model.StatusCancelled))
	assert.False(t, model.CanTransition(model.StatusReady, model.StatusCancelled))
	assert.False(t, model.CanTransition(model.StatusAssigned, model.StatusCancelled))
	assert.False(t, model.CanTransition(model.StatusInTransit, model.StatusCancelled))
}

func TestCanTransition_IllegalMoves(t *testing.T) {
	assert.False(t, model.CanTransition(model.StatusDelivered, model.StatusPreparing))
	assert.False(t, model.CanTransition(model.StatusCancelled, model.StatusPending))
	assert.False(t, model.CanTransition(model.StatusDelivered, model.StatusCancelled))
	assert.False(t, model.CanTransition(model.StatusPending, model.StatusDelivered)) // skip ahead
	assert.False(t, model.CanTransition(model.StatusInTransit, model.StatusConfirmed))
}

// ---- mock repo ----

type mockRepo struct{ mock.Mock }

func (m *mockRepo) Create(o *model.Order) error {
	return m.Called(o).Error(0)
}
func (m *mockRepo) FindByID(id uint) (*model.Order, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}
func (m *mockRepo) FindByCustomer(cid uint, page, limit int) ([]model.Order, int64, error) {
	args := m.Called(cid, page, limit)
	return args.Get(0).([]model.Order), args.Get(1).(int64), args.Error(2)
}
func (m *mockRepo) UpdateStatus(id uint, s model.Status) error {
	return m.Called(id, s).Error(0)
}

// ---- cancel tests ----

func TestCancelOrder_AlreadyCancelled(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewOrderService(repo, nil, nil)

	repo.On("FindByID", uint(1)).Return(&model.Order{
		ID: 1, CustomerID: 10, Status: model.StatusCancelled,
	}, nil)

	_, err := svc.CancelOrder(1, 10)
	assert.ErrorIs(t, err, service.ErrCannotCancel)
	repo.AssertNotCalled(t, "UpdateStatus")
}

func TestCancelOrder_AlreadyDelivered(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewOrderService(repo, nil, nil)

	repo.On("FindByID", uint(2)).Return(&model.Order{
		ID: 2, CustomerID: 10, Status: model.StatusDelivered,
	}, nil)

	_, err := svc.CancelOrder(2, 10)
	assert.ErrorIs(t, err, service.ErrCannotCancel)
	repo.AssertNotCalled(t, "UpdateStatus")
}

func TestCancelOrder_WhilePreparingNotAllowed(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewOrderService(repo, nil, nil)

	repo.On("FindByID", uint(3)).Return(&model.Order{
		ID: 3, CustomerID: 10, Status: model.StatusPreparing,
	}, nil)

	_, err := svc.CancelOrder(3, 10)
	assert.ErrorIs(t, err, service.ErrCannotCancel)
}

func TestCancelOrder_PendingWorks(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewOrderService(repo, nil, nil)

	repo.On("FindByID", uint(4)).Return(&model.Order{
		ID: 4, CustomerID: 10, Status: model.StatusPending,
	}, nil)
	repo.On("UpdateStatus", uint(4), model.StatusCancelled).Return(nil)

	order, err := svc.CancelOrder(4, 10)
	assert.NoError(t, err)
	assert.Equal(t, model.StatusCancelled, order.Status)
}

func TestCancelOrder_WrongCustomer(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewOrderService(repo, nil, nil)

	repo.On("FindByID", uint(5)).Return(&model.Order{
		ID: 5, CustomerID: 99, Status: model.StatusPending,
	}, nil)

	_, err := svc.CancelOrder(5, 10)
	assert.ErrorIs(t, err, service.ErrForbidden)
	repo.AssertNotCalled(t, "UpdateStatus")
}

func TestCancelOrder_NotFound(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewOrderService(repo, nil, nil)

	repo.On("FindByID", uint(999)).Return(nil, errors.New("record not found"))

	_, err := svc.CancelOrder(999, 10)
	assert.ErrorIs(t, err, service.ErrOrderNotFound)
}

// ---- update status tests ----

func TestUpdateStatus_ValidMove(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewOrderService(repo, nil, nil)

	repo.On("FindByID", uint(10)).Return(&model.Order{
		ID: 10, Status: model.StatusConfirmed,
	}, nil)
	repo.On("UpdateStatus", uint(10), model.StatusPreparing).Return(nil)

	order, err := svc.UpdateStatus(10, model.StatusPreparing)
	assert.NoError(t, err)
	assert.Equal(t, model.StatusPreparing, order.Status)
}

func TestUpdateStatus_InvalidMove(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewOrderService(repo, nil, nil)

	repo.On("FindByID", uint(11)).Return(&model.Order{
		ID: 11, Status: model.StatusDelivered,
	}, nil)

	_, err := svc.UpdateStatus(11, model.StatusPreparing)
	assert.ErrorIs(t, err, service.ErrInvalidTransition)
	repo.AssertNotCalled(t, "UpdateStatus")
}

func TestUpdateStatus_CancelledOrderIsTerminal(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewOrderService(repo, nil, nil)

	repo.On("FindByID", uint(12)).Return(&model.Order{
		ID: 12, Status: model.StatusCancelled,
	}, nil)

	_, err := svc.UpdateStatus(12, model.StatusConfirmed)
	assert.ErrorIs(t, err, service.ErrInvalidTransition)
	repo.AssertNotCalled(t, "UpdateStatus")
}
