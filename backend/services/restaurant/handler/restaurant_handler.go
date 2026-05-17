package handler

import (
	"net/http"
	"strconv"
	"time"

	"logistics-tracking-system/pkg/middleware"
	"logistics-tracking-system/pkg/response"
	"logistics-tracking-system/services/restaurant/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func getOwnerID(c *gin.Context) uint {
	if user, ok := middleware.GetUserFromCtx(c); ok {
		return user.ID
	}
	if header := c.GetHeader("X-Owner-ID"); header != "" {
		if id, err := strconv.ParseUint(header, 10, 64); err == nil {
			return uint(id)
		}
	}
	return 0
}

type RestaurantHandler struct {
	svc *service.RestaurantService
	db  *gorm.DB
}

func NewRestaurantHandler(svc *service.RestaurantService, db *gorm.DB) *RestaurantHandler {
	return &RestaurantHandler{svc: svc, db: db}
}

func (h *RestaurantHandler) Create(c *gin.Context) {
	var req service.CreateRestaurantReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	ownerID := getOwnerID(c)
	r, err := h.svc.CreateRestaurant(ownerID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": r})
}

func (h *RestaurantHandler) List(c *gin.Context) {
	lat, _ := strconv.ParseFloat(c.Query("lat"), 64)
	lng, _ := strconv.ParseFloat(c.Query("lng"), 64)
	radius, _ := strconv.ParseFloat(c.Query("radius"), 64)
	if radius == 0 {
		radius = 10
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	restaurants, total, err := h.svc.ListRestaurants(lat, lng, radius, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{"restaurants": restaurants, "total": total})
}

func (h *RestaurantHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}
	r, err := h.svc.GetRestaurant(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "restaurant not found")
		return
	}
	response.Success(c, r)
}

func (h *RestaurantHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}
	var req service.UpdateRestaurantReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	ownerID := getOwnerID(c)
	r, err := h.svc.UpdateRestaurant(uint(id), ownerID, req)
	if err != nil {
		if err.Error() == "forbidden" {
			response.Error(c, http.StatusForbidden, "forbidden")
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, r)
}

func (h *RestaurantHandler) Toggle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}
	ownerID := getOwnerID(c)
	r, err := h.svc.ToggleRestaurant(uint(id), ownerID)
	if err != nil {
		if err.Error() == "forbidden" {
			response.Error(c, http.StatusForbidden, "forbidden")
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, r)
}

func (h *RestaurantHandler) ListOrders(c *gin.Context) {
	restaurantID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}
	type OrderRow struct {
		ID              uint      `json:"id"`
		CustomerID      uint      `json:"customer_id"`
		Status          string    `json:"status"`
		Total           float64   `json:"total"`
		DeliveryAddress string    `json:"delivery_address"`
		CreatedAt       time.Time `json:"created_at"`
	}
	var orders []OrderRow
	if err := h.db.Table("orders").
		Where("restaurant_id = ?", restaurantID).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, orders)
}

func (h *RestaurantHandler) AcceptOrder(c *gin.Context) {
	orderID, _ := strconv.ParseUint(c.Param("orderId"), 10, 64)
	if err := h.db.Table("orders").Where("id = ?", orderID).Update("status", "CONFIRMED").Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{"status": "CONFIRMED"})
}

func (h *RestaurantHandler) ReadyOrder(c *gin.Context) {
	orderID, _ := strconv.ParseUint(c.Param("orderId"), 10, 64)
	if err := h.db.Table("orders").Where("id = ?", orderID).Update("status", "READY").Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{"status": "READY"})
}

func (h *RestaurantHandler) RejectOrder(c *gin.Context) {
	orderID, _ := strconv.ParseUint(c.Param("orderId"), 10, 64)
	if err := h.db.Table("orders").Where("id = ?", orderID).Update("status", "CANCELLED").Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{"status": "CANCELLED"})
}
