package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func MakeJWT(user_id pgtype.UUID, jwt_secret string, expires_in time.Duration) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256, 
		jwt.RegisteredClaims{
			Issuer: "capybara cloud",
			Subject: user_id.String(),
			Audience: jwt.ClaimStrings{"capybara cloud app"},
			IssuedAt: &jwt.NumericDate{Time: time.Now()},
			ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(expires_in)},
		},
	)

	signed_jwt, err := token.SignedString([]byte(jwt_secret))

	if err != nil {
		return "", err
	}

	return signed_jwt, nil
}

func ValidateJWT(token string, jwt_secret string) (string, error) {
	claims := jwt.RegisteredClaims{}
	
	_, err := jwt.ParseWithClaims(
		token, 
		&claims, 
		func (token *jwt.Token) (any, error) {
			return []byte(jwt_secret), nil 
		}, 
		func (parser *jwt.Parser) {},
	)

	if err != nil {
		return "", err
	}
	
	sub := claims.Subject
	
	return sub , nil
}