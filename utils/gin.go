package utils

import (
	"errors"
	"github.com/gin-gonic/gin"
	"sample-go-server/utils/log"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func HandleError(c *gin.Context, status int, err error) {
	l := log.Get()
	l.Err(err)
	// todo response message should be filtered
	c.AbortWithStatusJSON(status, ErrorResponse{Message: err.Error()})
}

func GetMiddlewareValues(c *gin.Context) (userUuid string, isAdmin bool, err error) {
	role, ok := c.Get(MiddlewareUserRoleKey)
	if !ok {
		return "", false, errors.New("cannot get user role")
	}

	uUuid, ok := c.Get(MiddlewareUserUuidKey)
	if !ok {
		return "", false, errors.New("cannot get user uuid")
	}

	return uUuid.(string), role.(bool), nil
}
