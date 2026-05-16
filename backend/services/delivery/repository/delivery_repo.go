package repository

import (
	"errors"
	"math"

	"logistics-tracking-system/services/delivery/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DeliveryRepository struct {
	db *gorm.DB
}

func NewDeliveryRepository(db *gorm.DB) *DeliveryRepository {
	return &DeliveryRepository{db: db}
}

func (r *DeliveryRepository) CreateAssignment(a *model.DeliveryAssignment) error {
	return r.db.Create(a).Error
}

func (r *DeliveryRepository) GetAssignment(id uint) (*model.DeliveryAssignment, error) {
	var a model.DeliveryAssignment
	if err := r.db.First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *DeliveryRepository) UpdateAssignment(a *model.DeliveryAssignment) error {
	return r.db.Save(a).Error
}

func (r *DeliveryRepository) AddLocation(loc *model.DeliveryLocation) error {
	return r.db.Create(loc).Error
}

func (r *DeliveryRepository) GetLatestLocation(deliveryID uint) (*model.DeliveryLocation, error) {
	var loc model.DeliveryLocation
	if err := r.db.Where("delivery_id = ?", deliveryID).Order("created_at DESC").First(&loc).Error; err != nil {
		return nil, err
	}
	return &loc, nil
}

func (r *DeliveryRepository) GetDriverStatus(driverID uint) (*model.DriverStatus, error) {
	var ds model.DriverStatus
	if err := r.db.Where("driver_id = ?", driverID).First(&ds).Error; err != nil {
		return nil, err
	}
	return &ds, nil
}

func (r *DeliveryRepository) GetOrCreateDriverStatus(driverID uint) (*model.DriverStatus, error) {
	var ds model.DriverStatus
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("driver_id = ?", driverID).First(&ds).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ds = model.DriverStatus{DriverID: driverID, IsOnline: true, IsAvailable: true}
		if err := r.db.Create(&ds).Error; err != nil {
			return nil, err
		}
		return &ds, nil
	}
	if err != nil {
		return nil, err
	}
	return &ds, nil
}

func (r *DeliveryRepository) UpdateDriverStatus(ds *model.DriverStatus) error {
	return r.db.Save(ds).Error
}

func (r *DeliveryRepository) FindNearestAvailableDriver(lat, lng float64) (*model.DriverStatus, error) {
	var drivers []model.DriverStatus
	if err := r.db.Where("is_online = true AND is_available = true").Find(&drivers).Error; err != nil {
		return nil, err
	}
	if len(drivers) == 0 {
		return nil, errors.New("no available drivers")
	}
	nearest := &drivers[0]
	minDist := haversine(lat, lng, drivers[0].LastLat, drivers[0].LastLng)
	for i := 1; i < len(drivers); i++ {
		d := haversine(lat, lng, drivers[i].LastLat, drivers[i].LastLng)
		if d < minDist {
			minDist = d
			nearest = &drivers[i]
		}
	}
	return nearest, nil
}

func (r *DeliveryRepository) GetAvailableDrivers() ([]model.DriverStatus, error) {
	var drivers []model.DriverStatus
	if err := r.db.Where("is_online = true AND is_available = true").Find(&drivers).Error; err != nil {
		return nil, err
	}
	return drivers, nil
}

func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
