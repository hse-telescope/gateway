package token

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hse-telescope/gateway/internal/clients/auth"
)

type Provider struct {
	auth      auth.Wrapper
	publicKey string
}

type UserInfo struct {
	jwt.RegisteredClaims
	UserID int `json:"user_id"`
}

func New(auth auth.Wrapper) Provider {
	return Provider{
		auth: auth,
	}
}

func (p Provider) ParseToken(token string) (UserInfo, bool) {
	var userInfo UserInfo
	jwtToken, err := jwt.ParseWithClaims(token, &userInfo, func(t *jwt.Token) (interface{}, error) {
		if t == nil {
			return nil, errors.New("nil token provided")
		}
		return &p.publicKey, nil
	})
	if err != nil || !jwtToken.Valid {
		return UserInfo{}, false
	}
	return userInfo, true
}
