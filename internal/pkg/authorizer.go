package pkg

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/docker/distribution/registry/auth"

	"github.com/syols/keeper/config"
)

type UserClaims struct {
	jwt.StandardClaims
	Username string
}

type Authorizer struct {
	sign []byte
}

func NewAuthorizer(config config.Config) Authorizer {
	return Authorizer{
		sign: []byte(config.Server.Sign),
	}
}

func (a *Authorizer) CreateToken(username string, ttl time.Duration) (string, error) {
	claims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().UTC().Add(ttl).Unix(),
			Issuer:    username,
		},
		Username: username,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.sign)
}

func (a *Authorizer) VerifyToken(token string) (string, error) {
	var claims UserClaims
	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, auth.ErrAuthenticationFailure
		}
		return a.sign, nil
	})
	return claims.Username, err
}
