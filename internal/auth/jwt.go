package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID      string   `json:"user_id"`
	EmployeeID  string   `json:"employee_id"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secretKey []byte
	issuer    string
	expiry    time.Duration
}

func NewJWTService(secretKey string, expiry time.Duration) *JWTService {
	return &JWTService{
		secretKey: []byte(secretKey),
		issuer:    "hr-management-system",
		expiry:    expiry,
	}
}

func (j *JWTService) GenerateToken(userID, employeeID, email, role string, permissions []string) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID:      userID,
		EmployeeID:  employeeID,
		Email:       email,
		Role:        role,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.expiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.Issuer != j.issuer {
			return nil, fmt.Errorf("invalid issuer")
		}

		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return nil, fmt.Errorf("token has expired")
		}

		if claims.NotBefore != nil && claims.NotBefore.Time.After(time.Now()) {
			return nil, fmt.Errorf("token used before valid time")
		}

		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (j *JWTService) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("invalid token for refresh: %w", err)
	}

	return j.GenerateToken(
		claims.UserID,
		claims.EmployeeID,
		claims.Email,
		claims.Role,
		claims.Permissions,
	)
}

// GetClaims extracts claims from a token without validation (for middleware use)
func (j *JWTService) GetClaims(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid claims")
}

// IsTokenExpired checks if a token is expired
func (j *JWTService) IsTokenExpired(tokenString string) bool {
	claims, err := j.GetClaims(tokenString)
	if err != nil {
		return true
	}

	if claims.ExpiresAt == nil {
		return false
	}

	return claims.ExpiresAt.Time.Before(time.Now())
}
