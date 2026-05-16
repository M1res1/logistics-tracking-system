package service

import (
	"errors"
	"math"
	"time"

	"logistics-tracking-system/services/delivery/model"
	"logistics-tracking-system/services/delivery/repository"

	"gorm.io/gorm"
)

type DeliveryService struct {
	db   *gorm.DB
	repo *repository.DeliveryRepository
}

func NewDeliveryService(db *gorm.DB, repo *repository.DeliveryRepository) *DeliveryService {
	return &DeliveryService{db: db, repo: repo}
}

func HaversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func EstimateETA(distanceKm float64) int {
	return int(math.Ceil(distanceKm / 30.0 * 60))
}

type AssignRequest struct {
	OrderID uint
	Lat     float64
	Lng     float64
}

func (s *DeliveryService) AssignDelivery(req *AssignRequest) (*model.DeliveryAssignment, error) {
	driver, err := s.repo.FindNearestAvailableDriver(req.Lat, req.Lng)
	if err != nil {
		return nil, err
	}

	dist := HaversineDistance(req.Lat, req.Lng, driver.LastLat, driver.LastLng)

	a := &model.DeliveryAssignment{
		OrderID:    req.OrderID,
		DriverID:   driver.DriverID,
		Status:     "PENDING",
		DistanceKm: dist,
	}

	if err := s.repo.CreateAssignment(a); err != nil {
		return nil, err
	}

	driver.IsAvailable = false
	_ = s.repo.UpdateDriverStatus(driver)

	return a, nil
}

func (s *DeliveryService) AcceptDelivery(id uint) (*model.DeliveryAssignment, error) {
	a, err := s.repo.GetAssignment(id)
	if err != nil {
		return nil, err
	}
	if a.Status != "PENDING" {
		return nil, errors.New("delivery not in pending state")
	}
	a.Status = "ACCEPTED"
	if err := s.repo.UpdateAssignment(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *DeliveryService) PickupDelivery(id uint) (*model.DeliveryAssignment, error) {
	a, err := s.repo.GetAssignment(id)
	if err != nil {
		return nil, err
	}
	if a.Status != "ACCEPTED" {
		return nil, errors.New("delivery must be accepted before pickup")
	}
	now := time.Now()
	a.Status = "IN_TRANSIT"
	a.PickupTime = &now
	if err := s.repo.UpdateAssignment(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *DeliveryService) CompleteDelivery(id uint) (*model.DeliveryAssignment, error) {
	a, err := s.repo.GetAssignment(id)
	if err != nil {
		return nil, err
	}
	if a.Status != "IN_TRANSIT" {
		return nil, errors.New("delivery must be in transit to complete")
	}
	now := time.Now()
	a.Status = "DELIVERED"
	a.DeliveryTime = &now
	if err := s.repo.UpdateAssignment(a); err != nil {
		return nil, err
	}

	ds, err := s.repo.GetDriverStatus(a.DriverID)
	if err == nil {
		ds.IsAvailable = true
		_ = s.repo.UpdateDriverStatus(ds)
	}

	return a, nil
}

func (s *DeliveryService) RejectBusyDriver(id uint) error {
	a, err := s.repo.GetAssignment(id)
	if err != nil {
		return err
	}
	ds, err := s.repo.GetDriverStatus(a.DriverID)
	if err != nil {
		return err
	}
	if !ds.IsAvailable {
		return errors.New("driver is not available")
	}
	return nil
}

func (s *DeliveryService) UpdateLocation(deliveryID uint, lat, lng float64) error {
	loc := &model.DeliveryLocation{
		DeliveryID: deliveryID,
		Lat:        lat,
		Lng:        lng,
	}
	return s.repo.AddLocation(loc)
}

func (s *DeliveryService) GetLocation(deliveryID uint) (*model.DeliveryLocation, error) {
	return s.repo.GetLatestLocation(deliveryID)
}

func (s *DeliveryService) GetAvailableDrivers() ([]model.DriverStatus, error) {
	return s.repo.GetAvailableDrivers()
}
