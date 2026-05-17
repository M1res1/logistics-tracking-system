package handler

import (
	"net/http"
	"strconv"

	"logistics-tracking-system/pkg/middleware"
	"logistics-tracking-system/pkg/response"
	"logistics-tracking-system/services/order/model"
	"logistics-tracking-system/services/order/service"

	"github.com/gin-gonic/gin"
)

// OrderHandler wires HTTP routes to the order service.
type OrderHandler struct {
	svc *service.OrderService
}

// NewOrderHandler constructs an OrderHandler.
func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// customerIDFromContext extracts the authenticated customer's ID from the Gin
// context set by the auth middleware. Falls back to the X-User-ID header and
// then to 1 when the middleware does not populate the context.
func customerIDFromContext(c *gin.Context) uint {
	if user, ok := middleware.GetUserFromCtx(c); ok {
		return user.ID
	}
	if header := c.GetHeader("X-User-ID"); header != "" {
		if id, err := strconv.ParseUint(header, 10, 64); err == nil {
			return uint(id)
		}
	}
	return 1 // fallback
}

// parseIDParam extracts and validates the :id URL parameter.
func parseIDParam(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return uint(id), true
}

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

type createOrderItemReq struct {
	MenuItemID          uint    `json:"menu_item_id"`
	Quantity            int     `json:"quantity"`
	UnitPrice           float64 `json:"unit_price"`
	SpecialInstructions string  `json:"special_instructions"`
}

type createOrderReq struct {
	RestaurantID    uint                 `json:"restaurant_id"    binding:"required"`
	Items           []createOrderItemReq `json:"items"            binding:"required,min=1"`
	DeliveryAddress string               `json:"delivery_address" binding:"required"`
	Lat             float64              `json:"lat"`
	Lng             float64              `json:"lng"`
}

type updateStatusReq struct {
	Status string `json:"status" binding:"required"`
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

// CreateOrder handles POST /orders.
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req createOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	customerID := customerIDFromContext(c)

	items := make([]service.CreateOrderItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, service.CreateOrderItem{
			MenuItemID:          it.MenuItemID,
			Quantity:            it.Quantity,
			UnitPrice:           it.UnitPrice,
			SpecialInstructions: it.SpecialInstructions,
		})
	}

	order, err := h.svc.CreateOrder(customerID, req.RestaurantID, items, req.DeliveryAddress, req.Lat, req.Lng)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, order)
}

// GetOrder handles GET /orders/:id.
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	customerID := customerIDFromContext(c)

	order, err := h.svc.GetOrder(id, customerID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, order)
}

// ListMyOrders handles GET /orders/my?page=1&limit=10.
func (h *OrderHandler) ListMyOrders(c *gin.Context) {
	customerID := customerIDFromContext(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	orders, total, err := h.svc.ListMyOrders(customerID, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{
		"orders": orders,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

// CancelOrder handles POST /orders/:id/cancel.
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	customerID := customerIDFromContext(c)

	if err := h.svc.CancelOrder(id, customerID); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "order cancelled"})
}

// UpdateStatus handles PUT /orders/:id/status (internal endpoint).
func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	var req updateStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	order, err := h.svc.UpdateStatus(id, model.OrderStatus(req.Status))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, order)
}
