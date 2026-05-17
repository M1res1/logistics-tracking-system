package handler

import (
	"net/http"
	"strconv"

	"logistics-tracking-system/pkg/response"
	"logistics-tracking-system/services/restaurant/service"

	"github.com/gin-gonic/gin"
)

type MenuHandler struct {
	svc *service.RestaurantService
}

func NewMenuHandler(svc *service.RestaurantService) *MenuHandler {
	return &MenuHandler{svc: svc}
}

func (h *MenuHandler) GetMenu(c *gin.Context) {
	restaurantID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid restaurant id")
		return
	}
	items, err := h.svc.GetMenu(uint(restaurantID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, items)
}

func (h *MenuHandler) AddItem(c *gin.Context) {
	restaurantID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid restaurant id")
		return
	}
	var req service.AddMenuItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	ownerID := getOwnerID(c)
	item, err := h.svc.AddMenuItem(uint(restaurantID), ownerID, req)
	if err != nil {
		if err.Error() == "forbidden" {
			response.Error(c, http.StatusForbidden, "forbidden")
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *MenuHandler) UpdateItem(c *gin.Context) {
	restaurantID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid restaurant id")
		return
	}
	itemID, err := strconv.ParseUint(c.Param("itemId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid item id")
		return
	}
	var req service.UpdateMenuItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	ownerID := getOwnerID(c)
	item, err := h.svc.UpdateMenuItem(uint(restaurantID), uint(itemID), ownerID, req)
	if err != nil {
		if err.Error() == "forbidden" {
			response.Error(c, http.StatusForbidden, "forbidden")
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *MenuHandler) DeleteItem(c *gin.Context) {
	restaurantID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid restaurant id")
		return
	}
	itemID, err := strconv.ParseUint(c.Param("itemId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid item id")
		return
	}
	ownerID := getOwnerID(c)
	if err := h.svc.DeleteMenuItem(uint(restaurantID), uint(itemID), ownerID); err != nil {
		if err.Error() == "forbidden" {
			response.Error(c, http.StatusForbidden, "forbidden")
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{"deleted": true})
}
