package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gopi.com/internal/apperr"
)

type errorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Meta    interface{} `json:"meta,omitempty"`
}

func respondError(c *gin.Context, err error) {
	code := apperr.CodeOf(err)
	status := httpStatus(code)
	var msg string
	if e, ok := err.(*apperr.Error); ok && e.Message != "" {
		msg = e.Message
	} else {
		msg = http.StatusText(status)
	}
	body := errorBody{Code: string(code), Message: msg}
	if e, ok := err.(*apperr.Error); ok && e.Meta != nil {
		body.Meta = e.Meta
	}
	c.JSON(status, body)
}

func httpStatus(code apperr.Code) int {
	switch code {
	case apperr.InvalidInput:
		return http.StatusBadRequest
	case apperr.NotFound:
		return http.StatusNotFound
	case apperr.Conflict:
		return http.StatusConflict
	case apperr.Unauthorized:
		return http.StatusUnauthorized
	case apperr.Forbidden:
		return http.StatusForbidden
	case apperr.Unavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
