package repository

import (
	"math"

	"logistics-tracking-system/services/restaurant/model"

	"gorm.io/gorm"
)

type RestaurantRepository struct {
	db *gorm.DB
}

func NewRestaurantRepository(db *gorm.DB) *RestaurantRepository {
	return &RestaurantRepository{db: db}
}

// HaversineKm returns the great-circle distance in kilometres between two points.
func HaversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	return haversineKm(lat1, lng1, lat2, lng2)
}

// haversineKm returns the great-circle distance in kilometres between two points.
func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// Restaurant CRUD

func (r *RestaurantRepository) Create(restaurant *model.Restaurant) error {
	return r.db.Create(restaurant).Error
}

func (r *RestaurantRepository) GetByID(id uint) (*model.Restaurant, error) {
	var restaurant model.Restaurant
	if err := r.db.First(&restaurant, id).Error; err != nil {
		return nil, err
	}
	return &restaurant, nil
}

func (r *RestaurantRepository) Update(restaurant *model.Restaurant) error {
	return r.db.Save(restaurant).Error
}

func (r *RestaurantRepository) List(page, limit int) ([]model.Restaurant, int64, error) {
	var restaurants []model.Restaurant
	var total int64

	offset := (page - 1) * limit
	if err := r.db.Model(&model.Restaurant{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := r.db.Offset(offset).Limit(limit).Find(&restaurants).Error; err != nil {
		return nil, 0, err
	}
	return restaurants, total, nil
}

// ListNearby loads all active restaurants and filters by haversine distance in Go.
func (r *RestaurantRepository) ListNearby(lat, lng, radiusKm float64) ([]model.Restaurant, error) {
	var all []model.Restaurant
	if err := r.db.Where("is_active = ?", true).Find(&all).Error; err != nil {
		return nil, err
	}

	var nearby []model.Restaurant
	for _, res := range all {
		if haversineKm(lat, lng, res.Lat, res.Lng) <= radiusKm {
			nearby = append(nearby, res)
		}
	}
	return nearby, nil
}

// MenuItem operations

func (r *RestaurantRepository) CreateMenuItem(item *model.MenuItem) error {
	return r.db.Create(item).Error
}

func (r *RestaurantRepository) GetMenuItem(id uint) (*model.MenuItem, error) {
	var item model.MenuItem
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *RestaurantRepository) UpdateMenuItem(item *model.MenuItem) error {
	return r.db.Save(item).Error
}

// SoftDeleteMenuItem sets is_available=false instead of removing the row.
func (r *RestaurantRepository) SoftDeleteMenuItem(id uint) error {
	return r.db.Model(&model.MenuItem{}).Where("id = ?", id).Update("is_available", false).Error
}

// ListMenuByRestaurant returns only available items for a restaurant.
func (r *RestaurantRepository) ListMenuByRestaurant(restaurantID uint) ([]model.MenuItem, error) {
	var items []model.MenuItem
	if err := r.db.Where("restaurant_id = ? AND is_available = ?", restaurantID, true).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
