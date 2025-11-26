package service

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"e-memo-job-reservation-api/internal/auth"
	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
	"e-memo-job-reservation-api/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	authRepo     *repository.AuthRepository
	userRepo     *repository.AppUserRepository
	posPermRepo  *repository.PositionPermissionRepository
	employeeRepo *repository.EmployeeRepository
}

func NewAuthService(authRepo *repository.AuthRepository, userRepo *repository.AppUserRepository, posPermRepo *repository.PositionPermissionRepository, employeeRepo *repository.EmployeeRepository) *AuthService {
	return &AuthService{
		authRepo:     authRepo,
		userRepo:     userRepo,
		posPermRepo:  posPermRepo,
		employeeRepo: employeeRepo,
	}
}

// LOGIN
func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByUsernameOrNPK(req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	var employee *model.Employee
	if user.EmployeeNPK.Valid {
		employee, err = s.employeeRepo.FindByNPK(user.EmployeeNPK.String)
		if err != nil {
			return nil, errors.New("employee data associated with user not found")
		}
	}

	accessToken, refreshToken, err := auth.GenerateTokens(user, employee, s.authRepo)
	if err != nil {
		return nil, err
	}

	userDetail, err := s.userRepo.GetUserDetailByID(user.ID)
	if err != nil {
		return nil, err
	}

	permissions, err := s.posPermRepo.FindPermissionsByPositionID(user.EmployeePositionID)
	if err != nil {
		return nil, err
	}

	var permissionNames []string
	for _, p := range permissions {
		permissionNames = append(permissionNames, p.Name)
	}
	userDetail.Permissions = permissionNames

	loginResponse := &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *userDetail,
	}

	return loginResponse, nil
}

// REFRESH TOKEN
func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenString string) (*dto.LoginResponse, error) {
	claims, err := auth.ValidateToken(refreshTokenString, true)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	err = s.authRepo.ValidateAndDelRefreshToken(ctx, claims.UserID, claims.TokenID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, errors.New("user associated with token not found")
	}

	var employee *model.Employee
	if user.EmployeeNPK.Valid {
		employee, err = s.employeeRepo.FindByNPK(user.EmployeeNPK.String)
		if err != nil {
			return nil, errors.New("employee data associated with user not found")
		}
	}

	accessToken, newRefreshToken, err := auth.GenerateTokens(user, employee, s.authRepo)
	if err != nil {
		return nil, err
	}

	userDetail, err := s.userRepo.GetUserDetailByID(user.ID)
	if err != nil {
		return nil, err
	}
	permissions, err := s.posPermRepo.FindPermissionsByPositionID(user.EmployeePositionID)
	if err != nil {
		return nil, err
	}
	var permissionNames []string
	for _, p := range permissions {
		permissionNames = append(permissionNames, p.Name)
	}
	userDetail.Permissions = permissionNames

	loginResponse := &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         *userDetail,
	}

	return loginResponse, nil
}

// LOG OUT
func (s *AuthService) Logout(ctx context.Context, tokenString string) error {
	claims, err := auth.ValidateToken(tokenString, false)
	if err != nil {
		return nil
	}

	if time.Now().After(claims.ExpiresAt.Time) {
		return nil
	}

	err = s.authRepo.BlacklistToken(ctx, claims.TokenID, claims.ExpiresAt.Time)
	if err != nil {
		return err
	}

	return s.authRepo.DeleteAllUserRefreshTokens(ctx, claims.UserID)
}

func (s *AuthService) GenerateWebSocketTicket(ctx context.Context, userID int) (string, error) {
	ticket := uuid.New().String()
	expiresIn := 15 * time.Second
	expiresAt := time.Now().Add(expiresIn)

	err := s.authRepo.StoreWebSocketTicket(ctx, ticket, userID, expiresAt)
	if err != nil {
		return "", err
	}

	return ticket, nil
}

func (s *AuthService) GeneratePublicWebSocketTicket(ctx context.Context) (string, error) {
	ticket := uuid.New().String()
	expiresIn := 15 * time.Second
	expiresAt := time.Now().Add(expiresIn)
	const publicUserID = 0

	log.Printf("[DEBUG] Storing public WebSocket ticket: %s, userID: %d, expiresAt: %v", ticket, publicUserID, expiresAt)
	err := s.authRepo.StoreWebSocketTicket(ctx, ticket, publicUserID, expiresAt)
	if err != nil {
		log.Printf("[ERROR] Failed to store WebSocket ticket in database: %v", err)
		return "", err
	}

	log.Printf("[DEBUG] Successfully stored WebSocket ticket: %s", ticket)
	return ticket, nil
}
