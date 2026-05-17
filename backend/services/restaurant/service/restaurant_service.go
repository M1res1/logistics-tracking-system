package service

import (
	"errors"

	"logistics-tracking-system/services/restaurant/model"
	"logistics-tracking-system/services/restaurant/repository"
)

// Request structs

type CreateRestaurantReq struct {
	Name         string `json:"name"`
	Address      string `json:"address"`
	Lat          float64 `json:"lat"`
	Lng          float64 `json:"lng"`
	CuisineTypes string `json:"cuisine_types"`
	OpeningTime  string `json:"opening_time"`
	ClosingTime  string `json:"closing_time"`
}

type UpdateRestaurantReq struct {
	Name         string `json:"name"`
	Address      string `json:"address"`
	CuisineTypes string `json:"cuisine_types"`
	OpeningTime  string `json:"opening_time"`
	ClosingTime  string `json:"closing_time"`
}

type AddMenuItemReq struct {
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	Category        string  `json:"category"`
	Price           float64 `json:"price"`
	PrepTimeMinutes int     `json:"prep_time_minutes"`
}

type UpdateMenuItemReq struct {
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	Category        string  `json:"category"`
	Price           float64 `json:"price"`
	IsAvailable     bool    `json:"is_available"`
	PrepTimeMinutes int     `json:"prep_time_minutes"`
}

// RestaurantService encapsulates business logic for restaurants and menu items.
type RestaurantService struct {
	repo *repository.RestaurantRepository
}

func NewRestaurantService(repo *repository.RestaurantRepository) *RestaurantService {
	return &RestaurantService{repo: repo}
}

func (s *RestaurantService) CreateRestaurant(ownerID uint, req CreateRestaurantReq) (*model.Restaurant, error) {
	r := &model.Restaurant{
		OwnerID:      ownerID,
		Name:         req.Name,
		Address:      req.Address,
		Lat:          req.Lat,
		Lng:          req.Lng,
		CuisineTypes: req.CuisineTypes,
		OpeningTime:  req.OpeningTime,
		ClosingTime:  req.ClosingTime,
		IsActive:     true,
	}
	if err := s.repo.Create(r); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *RestaurantService) GetRestaurant(id uint) (*model.Restaurant, error) {
	return s.repo.GetByID(id)
}

func (s *RestaurantService) UpdateRestaurant(id, ownerID uint, req UpdateRestaurantReq) (*model.Restaurant, error) {
	r, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if r.OwnerID != ownerID {
		return nil, errors.New("forbidden")
	}
	r.Name = req.Name
	r.Address = req.Address
	r.CuisineTypes = req.CuisineTypes
	r.OpeningTime = req.OpeningTime
	r.ClosingTime = req.ClosingTime
	if err := s.repo.Update(r); err != nil {
		return nil, err
	}
	return r, nil
}

// ListRestaurants returns restaurants filtered by geo-proximity when lat/lng are nonzero,
// or all restaurants (paginated) otherwise.
func (s *RestaurantService) ListRestaurants(lat, lng, radiusKm float64, page, limit int) ([]model.Restaurant, int64, error) {
	if lat != 0 || lng != 0 {
		nearby, err := s.repo.ListNearby(lat, lng, radiusKm)
		if err != nil {
			return nil, 0, err
		}
		return nearby, int64(len(nearby)), nil
	}
	return s.repo.List(page, limit)
}

// ToggleRestaurant flips the IsActive flag, enforcing ownership.
func (s *RestaurantService) ToggleRestaurant(id, ownerID uint) (*model.Restaurant, error) {
	r, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if r.OwnerID != ownerID {
		return nil, errors.New("forbidden")
	}
	r.IsActive = !r.IsActive
	if err := s.repo.Update(r); err != nil {
		return nil, err
	}
	return r, nil
}

// Menu operations

func (s *RestaurantService) GetMenu(restaurantID uint) ([]model.MenuItem, error) {
	return s.repo.ListMenuByRestaurant(restaurantID)
}

func (s *RestaurantService) AddMenuItem(restaurantID, ownerID uint, req AddMenuItemReq) (*model.MenuItem, error) {
	if err := s.verifyOwnership(restaurantID, ownerID); err != nil {
		return nil, err
	}
	item := &model.MenuItem{
		RestaurantID:    restaurantID,
		Name:            req.Name,
		Description:     req.Description,
		Category:        req.Category,
		Price:           req.Price,
		PrepTimeMinutes: req.PrepTimeMinutes,
		IsAvailable:     true,
	}
	if err := s.repo.CreateMenuItem(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *RestaurantService) UpdateMenuItem(restaurantID, itemID, ownerID uint, req UpdateMenuItemReq) (*model.MenuItem, error) {
	if err := s.verifyOwnership(restaurantID, ownerID); err != nil {
		return nil, err
	}
	item, err := s.repo.GetMenuItem(itemID)
	if err != nil {
		return nil, err
	}
	item.Name = req.Name
	item.Description = req.Description
	item.Category = req.Category
	item.Price = req.Price
	item.IsAvailable = req.IsAvailable
	item.PrepTimeMinutes = req.PrepTimeMinutes
	if err := s.repo.UpdateMenuItem(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *RestaurantService) DeleteMenuItem(restaurantID, itemID, ownerID uint) error {
	if err := s.verifyOwnership(restaurantID, ownerID); err != nil {
		return err
	}
	return s.repo.SoftDeleteMenuItem(itemID)
}

// verifyOwnership checks that the given ownerID owns the restaurant.
func (s *RestaurantService) verifyOwnership(restaurantID, ownerID uint) error {
	r, err := s.repo.GetByID(restaurantID)
	if err != nil {
		return err
	}
	if r.OwnerID != ownerID {
		return errors.New("forbidden")
	}
	return nil
}
