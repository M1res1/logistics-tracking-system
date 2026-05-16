package handler

import (
	"net/http"
	"strconv"

	"logistics-tracking-system/pkg/response"
	"logistics-tracking-system/services/delivery/service"

	"github.com/gin-gonic/gin"
)

type TrackingHandler struct {
	svc *service.DeliveryService
}

func NewTrackingHandler(svc *service.DeliveryService) *TrackingHandler {
	return &TrackingHandler{svc: svc}
}

type updateLocationReq struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (h *TrackingHandler) UpdateLocation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}

	var req updateLocationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.svc.UpdateLocation(uint(id), req.Lat, req.Lng); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"ok": true})
}

func (h *TrackingHandler) GetLocation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}

	loc, err := h.svc.GetLocation(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "location not found")
		return
	}

	response.Success(c, loc)
}

func (h *TrackingHandler) AvailableDrivers(c *gin.Context) {
	drivers, err := h.svc.GetAvailableDrivers()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, drivers)
}
