package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success    bool        `json:"success"`
	StatusCode int         `json:"statusCode"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Errors     interface{} `json:"errors,omitempty"`
	Meta       interface{} `json:"meta,omitempty"`
}

// Mengirimkan respons JSON sukses dengan format envelope standar.
func Success(c *gin.Context, statusCode int, message string, data interface{}, meta ...interface{}) {
	resp := APIResponse{
		Success:    true,
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	}
	if len(meta) > 0 {
		resp.Meta = meta[0]
	}
	c.JSON(statusCode, resp)
}

// Mengirimkan respons JSON error dengan format envelope standar.
func Error(c *gin.Context, statusCode int, message string, errors ...interface{}) {
	resp := APIResponse{
		Success:    false,
		StatusCode: statusCode,
		Message:    message,
	}
	if len(errors) > 0 {
		resp.Errors = errors[0]
	}
	c.JSON(statusCode, resp)
}

// Mengirimkan respons error HTTP 400 Bad Request.
func BadRequest(c *gin.Context, message string, errors ...interface{}) {
	Error(c, http.StatusBadRequest, message, errors...)
}

// Mengirimkan respons error HTTP 401 Unauthorized.
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// Mengirimkan respons error HTTP 403 Forbidden.
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

// Mengirimkan respons error HTTP 404 Not Found.
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// Mengirimkan respons error HTTP 500 Internal Server Error.
func InternalError(c *gin.Context, message string, errors ...interface{}) {
	Error(c, http.StatusInternalServerError, message, errors...)
}
