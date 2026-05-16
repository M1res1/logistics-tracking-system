package repository

import (
	"auth-go/internal/model"
	"gorm.io/gorm"
)

type SessionRepository interface {
	FindByToken(token string) (*model.Session, error)
	DeleteByToken(token string) error
	Delete(session *model.Session) error
	Save(session *model.Session) error
}

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) FindByToken(token string) (*model.Session, error) {
	var session model.Session
	err := r.db.Preload("User").Where("token = ?", token).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) DeleteByToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&model.Session{}).Error
}

func (r *sessionRepository) Delete(session *model.Session) error {
	return r.db.Delete(session).Error
}

func (r *sessionRepository) Save(session *model.Session) error {
	return r.db.Save(session).Error
}
