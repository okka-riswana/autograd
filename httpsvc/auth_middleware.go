package httpsvc

import (
	"errors"
	"net/http"
	"strings"

	"github.com/fahmifan/autograd/model"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

// ErrUnauthorized error
var ErrUnauthorized = errors.New("unauthorized")

// ErrMissingAuthorization error
var ErrMissingAuthorization = errors.New("missing Authorization header")

func (s *Server) authorizedOne(perms ...model.Permission) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, err := parseTokenFromHeader(&c.Request().Header)
			if err != nil {
				log.Error(err)
				return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid token"})
			}

			user, ok := auth(token)
			if !ok {
				log.Error(ErrUnauthorized)
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": ErrUnauthorized.Error()})
			}

			setUserToCtx(c, user)
			if user == nil {
				log.Warn("user nil")
				return responseError(c, ErrUnauthorized)
			}

			for _, p := range perms {
				if user.Role.Granted(p) {
					return next(c)
				}
			}

			return responseError(c, ErrUnauthorized)
		}
	}
}

// AuthMiddleware ..
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token, err := parseTokenFromHeader(&c.Request().Header)
		if err != nil {
			log.Error(err)
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid token"})
		}

		user, ok := auth(token)
		if !ok {
			log.Error(ErrUnauthorized)
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": ErrUnauthorized.Error()})
		}

		setUserToCtx(c, user)
		return next(c)
	}
}

func parseTokenFromHeader(header *http.Header) (string, error) {
	var token string

	authHeaders := strings.Split(header.Get("Authorization"), " ")
	if authHeaders[0] != "Bearer" {
		err := ErrMissingAuthorization
		log.WithField("Authorization", header.Get("Authorization")).Error(err)
		return token, err
	}

	token = strings.Trim(authHeaders[1], " ")
	if token == "" {
		err := ErrMissingAuthorization
		log.WithField("Authorization", header.Get("Authorization")).Error(err)
		return token, err
	}

	return token, nil
}
