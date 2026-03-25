package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

const (
	DefaultTTL = 24 * time.Hour
)

type Claims struct {
	UserID         string `json:"user_id,omitempty"`
	OrganizationID string `json:"organization_id,omitempty"`
	IsSuperAdmin   bool   `json:"is_super_admin"`
	IsAdmin        bool   `json:"is_admin"`
	jwt.StandardClaims
}

func secret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return "dev-secret-change-me"
}

func GenerateUserToken(userID uuid.UUID, organizationID uuid.UUID, isAdmin bool) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:         userID.String(),
		OrganizationID: organizationID.String(),
		IsAdmin:        isAdmin,
		IsSuperAdmin:   false,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(DefaultTTL).Unix(),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret()))
}

func GenerateSuperAdminToken(adminID uuid.UUID) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:       adminID.String(),
		IsSuperAdmin: true,
		IsAdmin:      true,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(DefaultTTL).Unix(),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret()))
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret()), nil
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
