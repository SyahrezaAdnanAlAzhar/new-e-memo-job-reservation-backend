package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"e-memo-job-reservation-api/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID             int     `json:"uid"`
	UserType           string  `json:"typ"`
	EmployeeNPK        *string `json:"npk,omitempty"`
	EmployeePositionID int     `json:"pos_id"`
	DepartmentID       *int    `json:"dept_id,omitempty"`
	AreaID             *int    `json:"area_id,omitempty"`
	TokenID            string  `json:"jti"`
	jwt.RegisteredClaims
}

type TokenStorer interface {
	StoreRefreshToken(ctx context.Context, userID int, tokenID string, expiresAt time.Time) error
}

func GenerateTokens(user *model.AppUser, employee *model.Employee, tokenStore TokenStorer) (accessToken string, refreshToken string, err error) {
	// GENERATE ACCESS TOKEN
	accessLifespanStr := os.Getenv("ACCESS_TOKEN_LIFESPAN")
	accessDuration, err := time.ParseDuration(accessLifespanStr)

	if err != nil {
		log.Printf("Invalid ACCESS_TOKEN_LIFESPAN format, defaulting to 15m. Error: %v", err)
		accessDuration = 15 * time.Minute
	}

	var npkClaim *string
	var deptIDClaim, areaIDClaim *int
	if employee != nil {
		npkClaim = &employee.NPK
		deptIDClaim = &employee.DepartmentID
		if employee.AreaID.Valid {
			areaID := int(employee.AreaID.Int64)
			areaIDClaim = &areaID
		}
	}

	if user.EmployeeNPK.Valid {
		npkClaim = &user.EmployeeNPK.String
	}

	accessClaims := &Claims{
		UserID:             user.ID,
		UserType:           user.UserType,
		EmployeeNPK:        npkClaim,
		EmployeePositionID: user.EmployeePositionID,
		DepartmentID:       deptIDClaim,
		AreaID:             areaIDClaim,
		TokenID:            uuid.New().String(),
		RegisteredClaims:   jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessDuration))},
	}

	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		return "", "", err
	}

	// GENERATE REFRESH TOKEN
	refreshLifespanStr := os.Getenv("REFRESH_TOKEN_LIFESPAN")
	refreshDuration, err := time.ParseDuration(refreshLifespanStr)
	if err != nil {
		log.Printf("Invalid REFRESH_TOKEN_LIFESPAN format, defaulting to 720h. Error: %v", err)
		refreshDuration = 720 * time.Hour
	}

	// [BARU] Hitung waktu kedaluwarsa absolut
	expiresAt := time.Now().Add(refreshDuration)

	refreshClaims := &Claims{
		UserID:             user.ID,
		EmployeePositionID: user.EmployeePositionID,
		TokenID:            uuid.New().String(),
		RegisteredClaims:   jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expiresAt)}, // Gunakan expiresAt
	}

	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(os.Getenv("JWT_REFRESH_SECRET_KEY")))
	if err != nil {
		return "", "", err
	}

	// --- STORE REFRESH TOKEN ---
	err = tokenStore.StoreRefreshToken(context.Background(), user.ID, refreshClaims.TokenID, expiresAt)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func ValidateToken(tokenString string, isRefreshToken bool) (*Claims, error) {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if isRefreshToken {
		secretKey = os.Getenv("JWT_REFRESH_SECRET_KEY")
	}

	if secretKey == "" {
		return nil, errors.New("jwt secret key is not set")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
