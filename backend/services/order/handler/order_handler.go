package handler

import (
	"net/http"
	"strconv"

	"food-delivery/pkg/response"
	"food-delivery/services/order/service"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	svc service.OrderService
}

func NewOrderHandler(svc service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	customerID := c.GetUint("userID")

	var req service.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	order, err := h.svc.CreateOrder(customerID, req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, order)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	customerID := c.GetUint("userID")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}

	order, err := h.svc.GetOrder(id, customerID)
	if err == service.ErrOrderNotFound {
		response.NotFound(c, "order not found")
		return
	}
	if err == service.ErrForbidden {
		response.Forbidden(c, "you don't own this order")
		return
	}
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, order)
}

func (h *OrderHandler) ListMyOrders(c *gin.Context) {
	customerID := c.GetUint("userID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	orders, total, err := h.svc.ListMyOrders(customerID, page, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    orders,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	customerID := c.GetUint("userID")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}

	order, err := h.svc.CancelOrder(id, customerID)
	if err == service.ErrOrderNotFound {
		response.NotFound(c, "order not found")
		return
	}
	if err == service.ErrForbidden {
		response.Forbidden(c, "you don't own this order")
		return
	}
	if err == service.ErrCannotCancel {
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": err.Error()})
		return
	}
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, order)
}

// UpdateStatus is an internal endpoint — other services call this to advance order state
func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}

	var req service.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	order, err := h.svc.UpdateStatus(id, req.Status)
	if err == service.ErrOrderNotFound {
		response.NotFound(c, "order not found")
		return
	}
	if err == service.ErrInvalidTransition {
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": err.Error()})
		return
	}
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, order)
}

func parseID(c *gin.Context) (uint, error) {
	v, err := strconv.ParseUint(c.Param("id"), 10, 64)
	return uint(v), err
}
