package oauth

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"strings"
)

func JWTMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {

		var err error
		if !strings.Contains(c.Request.RequestURI, "static") && (c.Request.RequestURI != "/manifest.json") {
			_, err = checkJWT(c)
		}

		if err != nil {
			c.JSON(http.StatusBadRequest, apps.NewErrorResponse(err))
			return
		}
	}
}

func checkJWT(c *gin.Context) (*apps.JWTClaims, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))

	authValue := c.GetHeader(apps.OutgoingAuthHeader)
	if !strings.HasPrefix(authValue, "Bearer ") {
		return nil, errors.Errorf("missing %s: Bearer header", apps.OutgoingAuthHeader)
	}

	jwtoken := strings.TrimPrefix(authValue, "Bearer ")
	claims := apps.JWTClaims{}
	_, err := jwt.ParseWithClaims(jwtoken, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	return &claims, nil
}
