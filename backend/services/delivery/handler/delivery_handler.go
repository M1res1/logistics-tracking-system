package handler

import (
	"net/http"
	"strconv"

	"logistics-tracking-system/pkg/response"
	"logistics-tracking-system/services/delivery/service"

	"github.com/gin-gonic/gin"
)

type DeliveryHandler struct {
	svc *service.DeliveryService
}

func NewDeliveryHandler(svc *service.DeliveryService) *DeliveryHandler {
	return &DeliveryHandler{svc: svc}
}

type assignReq struct {
	OrderID uint    `json:"order_id"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}

func (h *DeliveryHandler) Assign(c *gin.Context) {
	var req assignReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	a, err := h.svc.AssignDelivery(&service.AssignRequest{
		OrderID: req.OrderID,
		Lat:     req.Lat,
		Lng:     req.Lng,
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, a)
}

func (h *DeliveryHandler) Accept(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}
	a, err := h.svc.AcceptDelivery(uint(id))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, a)
}

func (h *DeliveryHandler) Pickup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}
	a, err := h.svc.PickupDelivery(uint(id))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, a)
}

func (h *DeliveryHandler) Complete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}
	a, err := h.svc.CompleteDelivery(uint(id))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, a)
}
