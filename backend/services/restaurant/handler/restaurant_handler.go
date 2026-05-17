package handler

import (
	"net/http"
	"strconv"

	"logistics-tracking-system/pkg/middleware"
	"logistics-tracking-system/pkg/response"
	"logistics-tracking-system/services/restaurant/service"

	"github.com/gin-gonic/gin"
)

// getOwnerID retrieves the authenticated user's ID from the gin context.
// It first checks the middleware.User struct set by Auth middleware, then falls
// back to the X-Owner-ID header (useful for development / stub auth), then 0.
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
}

func NewRestaurantHandler(svc *service.RestaurantService) *RestaurantHandler {
	return &RestaurantHandler{svc: svc}
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
		radius = 10 // default 10 km
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

// Kitchen order stubs — full implementation requires integration with the order service.

func (h *RestaurantHandler) ListOrders(c *gin.Context) {
	response.Success(c, []struct{}{})
}

func (h *RestaurantHandler) AcceptOrder(c *gin.Context) {
	response.Success(c, gin.H{"status": "CONFIRMED"})
}

func (h *RestaurantHandler) ReadyOrder(c *gin.Context) {
	response.Success(c, gin.H{"status": "READY"})
}

func (h *RestaurantHandler) RejectOrder(c *gin.Context) {
	response.Success(c, gin.H{"status": "CANCELLED"})
}
