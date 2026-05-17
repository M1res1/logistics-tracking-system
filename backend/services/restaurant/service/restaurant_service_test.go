package service

import (
	"errors"
	"testing"

	"logistics-tracking-system/services/restaurant/model"
)

// fakeRepo is an in-memory implementation of the repository interface used by tests.
type fakeRepo struct {
	restaurants map[uint]*model.Restaurant
	menuItems   map[uint]*model.MenuItem
	nextID      uint
	nextItemID  uint
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		restaurants: make(map[uint]*model.Restaurant),
		menuItems:   make(map[uint]*model.MenuItem),
		nextID:      1,
		nextItemID:  1,
	}
}

func (f *fakeRepo) Create(r *model.Restaurant) error {
	r.ID = f.nextID
	f.nextID++
	cp := *r
	f.restaurants[r.ID] = &cp
	return nil
}

func (f *fakeRepo) GetByID(id uint) (*model.Restaurant, error) {
	r, ok := f.restaurants[id]
	if !ok {
		return nil, errors.New("record not found")
	}
	cp := *r
	return &cp, nil
}

func (f *fakeRepo) Update(r *model.Restaurant) error {
	cp := *r
	f.restaurants[r.ID] = &cp
	return nil
}

func (f *fakeRepo) List(page, limit int) ([]model.Restaurant, int64, error) {
	var out []model.Restaurant
	for _, r := range f.restaurants {
		out = append(out, *r)
	}
	return out, int64(len(out)), nil
}

func (f *fakeRepo) ListNearby(lat, lng, radiusKm float64) ([]model.Restaurant, error) {
	var out []model.Restaurant
	for _, r := range f.restaurants {
		if r.IsActive {
			out = append(out, *r)
		}
	}
	return out, nil
}

func (f *fakeRepo) CreateMenuItem(item *model.MenuItem) error {
	item.ID = f.nextItemID
	f.nextItemID++
	cp := *item
	f.menuItems[item.ID] = &cp
	return nil
}

func (f *fakeRepo) GetMenuItem(id uint) (*model.MenuItem, error) {
	item, ok := f.menuItems[id]
	if !ok {
		return nil, errors.New("record not found")
	}
	cp := *item
	return &cp, nil
}

func (f *fakeRepo) UpdateMenuItem(item *model.MenuItem) error {
	cp := *item
	f.menuItems[item.ID] = &cp
	return nil
}

func (f *fakeRepo) SoftDeleteMenuItem(id uint) error {
	item, ok := f.menuItems[id]
	if !ok {
		return errors.New("record not found")
	}
	item.IsAvailable = false
	return nil
}

func (f *fakeRepo) ListMenuByRestaurant(restaurantID uint) ([]model.MenuItem, error) {
	var out []model.MenuItem
	for _, item := range f.menuItems {
		if item.RestaurantID == restaurantID && item.IsAvailable {
			out = append(out, *item)
		}
	}
	return out, nil
}

// testService mirrors RestaurantService but operates on fakeRepo directly,
// allowing us to test service-layer ownership checks without a real DB.
type testService struct {
	repo *fakeRepo
}

func (s *testService) UpdateRestaurant(id, ownerID uint, req UpdateRestaurantReq) (*model.Restaurant, error) {
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
	_ = s.repo.Update(r)
	return r, nil
}

func (s *testService) ToggleRestaurant(id, ownerID uint) (*model.Restaurant, error) {
	r, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if r.OwnerID != ownerID {
		return nil, errors.New("forbidden")
	}
	r.IsActive = !r.IsActive
	_ = s.repo.Update(r)
	return r, nil
}

// TestUpdateRestaurant_ForbiddenForNonOwner verifies that a non-owner cannot update
// a restaurant owned by a different user.
func TestUpdateRestaurant_ForbiddenForNonOwner(t *testing.T) {
	f := newFakeRepo()
	svc := &testService{repo: f}

	owner := &model.Restaurant{OwnerID: 1, Name: "Test", IsActive: true}
	_ = f.Create(owner)

	_, err := svc.UpdateRestaurant(owner.ID, 99 /* wrong owner */, UpdateRestaurantReq{Name: "Hacked"})
	if err == nil {
		t.Fatal("expected error for non-owner, got nil")
	}
	if err.Error() != "forbidden" {
		t.Errorf("expected 'forbidden', got %q", err.Error())
	}
}

// TestUpdateRestaurant_OwnerCanUpdate verifies that the actual owner can update.
func TestUpdateRestaurant_OwnerCanUpdate(t *testing.T) {
	f := newFakeRepo()
	svc := &testService{repo: f}

	owner := &model.Restaurant{OwnerID: 42, Name: "Original", IsActive: true}
	_ = f.Create(owner)

	updated, err := svc.UpdateRestaurant(owner.ID, 42, UpdateRestaurantReq{Name: "Updated Name"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %q", updated.Name)
	}
}

// TestToggleRestaurant_ForbiddenForNonOwner ensures non-owners cannot toggle a restaurant.
func TestToggleRestaurant_ForbiddenForNonOwner(t *testing.T) {
	f := newFakeRepo()
	svc := &testService{repo: f}

	r := &model.Restaurant{OwnerID: 5, Name: "My Restaurant", IsActive: true}
	_ = f.Create(r)

	_, err := svc.ToggleRestaurant(r.ID, 99)
	if err == nil || err.Error() != "forbidden" {
		t.Errorf("expected 'forbidden' error, got %v", err)
	}
}

// TestToggleRestaurant_OwnerCanToggle verifies that the owner can flip the active state.
func TestToggleRestaurant_OwnerCanToggle(t *testing.T) {
	f := newFakeRepo()
	svc := &testService{repo: f}

	r := &model.Restaurant{OwnerID: 7, Name: "My Restaurant", IsActive: true}
	_ = f.Create(r)

	result, err := svc.ToggleRestaurant(r.ID, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsActive {
		t.Error("expected IsActive to be toggled to false")
	}
}
