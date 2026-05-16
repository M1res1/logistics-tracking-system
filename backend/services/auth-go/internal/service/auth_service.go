package service

import (
	"auth-go/internal/dto"
	"auth-go/internal/model"
	"auth-go/internal/repository"
	"auth-go/internal/util"
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(req *dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(req *dto.LoginRequest) (*dto.AuthResponse, error)
	Logout(ctx context.Context, authHeader string) error
	RefreshToken(refreshToken string) (*dto.AuthResponse, error)
	GetCurrentUser(email string) (*dto.UserProfile, error)
}

type authService struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	jwtUtil     *util.JwtUtil
	redis       *redis.Client
}

func NewAuthService(userRepo repository.UserRepository, sessionRepo repository.SessionRepository, jwtUtil *util.JwtUtil, redis *redis.Client) AuthService {
	return &authService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtUtil:     jwtUtil,
		redis:       redis,
	}
}

func (s *authService) Register(req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	exists, err := s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Phone:        req.Phone,
		UserType:     req.UserType,
		IsActive:     true,
	}

	if err := s.userRepo.Save(user); err != nil {
		return nil, err
	}

	accessToken, err := s.jwtUtil.GenerateToken(user.Email, user.UserType, user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtUtil.GenerateRefreshToken(user.Email)
	if err != nil {
		return nil, err
	}

	if err := s.saveSession(user, refreshToken, ""); err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) Login(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	accessToken, err := s.jwtUtil.GenerateToken(user.Email, user.UserType, user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtUtil.GenerateRefreshToken(user.Email)
	if err != nil {
		return nil, err
	}

	if err := s.saveSession(user, refreshToken, req.DeviceInfo); err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) Logout(ctx context.Context, authHeader string) error {
	if authHeader == "" || len(authHeader) < 7 {
		return nil
	}
	token := authHeader[7:]

	expiry, err := s.jwtUtil.ExtractExpiration(token)
	if err != nil {
		return nil
	}

	duration := time.Until(expiry)
	if duration > 0 {
		return s.redis.Set(ctx, "blacklist:"+token, "true", duration).Err()
	}
	return nil
}

func (s *authService) RefreshToken(refreshToken string) (*dto.AuthResponse, error) {
	session, err := s.sessionRepo.FindByToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if session.ExpiresAt.Before(time.Now()) {
		s.sessionRepo.Delete(session)
		return nil, errors.New("refresh token expired")
	}

	user := &session.User
	newAccessToken, err := s.jwtUtil.GenerateToken(user.Email, user.UserType, user.ID)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) GetCurrentUser(email string) (*dto.UserProfile, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return &dto.UserProfile{
		ID:       user.ID,
		Email:    user.Email,
		Phone:    user.Phone,
		UserType: user.UserType,
	}, nil
}

func (s *authService) saveSession(user *model.User, token string, deviceInfo string) error {
	expiresAt, err := s.jwtUtil.ExtractExpiration(token)
	if err != nil {
		return err
	}

	session := &model.Session{
		UserID:     user.ID,
		Token:      token,
		ExpiresAt:  expiresAt,
		DeviceInfo: deviceInfo,
	}

	return s.sessionRepo.Save(session)
}
