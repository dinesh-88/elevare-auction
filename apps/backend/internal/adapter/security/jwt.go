package security

import (
	"errors"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
	Secret []byte
	TTL    time.Duration
}

type claims struct {
	UserID   string      `json:"user_id"`
	TenantID string      `json:"tenant_id"`
	Role     domain.Role `json:"role"`
	Locale   string      `json:"locale"`
	jwt.RegisteredClaims
}

func (j *JWTService) Issue(userID, tenantID uuid.UUID, role domain.Role, locale domain.Locale) (string, error) {
	now := time.Now()
	cl := claims{
		UserID:   userID.String(),
		TenantID: tenantID.String(),
		Role:     role,
		Locale:   string(domain.NormalizeLocale(string(locale))),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.TTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, cl)
	return token.SignedString(j.Secret)
}

func (j *JWTService) Parse(raw string) (*port.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(raw, &claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return j.Secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	cl, ok := token.Claims.(*claims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	uid, err := uuid.Parse(cl.UserID)
	if err != nil {
		return nil, err
	}
	tid, err := uuid.Parse(cl.TenantID)
	if err != nil {
		return nil, err
	}
	return &port.TokenClaims{
		UserID:   uid,
		TenantID: tid,
		Role:     cl.Role,
		Locale:   domain.NormalizeLocale(cl.Locale),
		Exp:      cl.ExpiresAt.Unix(),
	}, nil
}
