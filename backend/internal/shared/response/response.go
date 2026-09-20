package response

import (
	"log"
	"net/http"
	"strings"

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

// Mengirimkan respons error HTTP 409 Conflict.
func Conflict(c *gin.Context, message string, errors ...interface{}) {
	Error(c, http.StatusConflict, message, errors...)
}

// Mengirimkan respons error HTTP 429 Too Many Requests.
func TooManyRequests(c *gin.Context, message string, errors ...interface{}) {
	Error(c, http.StatusTooManyRequests, message, errors...)
}

// Mengirimkan respons error HTTP 500 Internal Server Error tanpa membocorkan detail teknis internal.
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

// Mencatat detail kesalahan teknis di sisi server dan mengembalikan pesan aman bagi klien (F-ERR-01).
// Gunakan helper ini setiap kali err berasal dari lapisan service/database agar teks driver
// (pgx, SQLSTATE, nama tabel/kolom) tidak pernah terkirim ke respons API.
func Fail(c *gin.Context, statusCode int, clientMessage string, err error) {
	if err != nil {
		log.Printf("Kesalahan %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	}
	Error(c, statusCode, clientMessage)
}

// Mencatat kesalahan dan mengembalikan Bad Request aman bagi klien.
func FailBadRequest(c *gin.Context, clientMessage string, err error) {
	Fail(c, http.StatusBadRequest, clientMessage, err)
}

// Mencatat kesalahan dan mengembalikan Not Found aman bagi klien.
func FailNotFound(c *gin.Context, clientMessage string, err error) {
	Fail(c, http.StatusNotFound, clientMessage, err)
}

// internalMarkers adalah fragmen teks khas driver database dan runtime yang tidak
// boleh bocor ke respons klien (F-ERR-01).
var internalMarkers = []string{
	"sqlstate", "error:", "pq:", "pgx", "duplicate key", "violates",
	"relation ", "column ", "constraint ", "stack trace", "goroutine ",
	"syntax error", "connection refused", "dial tcp",
}

// containsInternalDetail memeriksa apakah pesan error mengandung detail teknis internal.
func containsInternalDetail(message string) bool {
	lowered := strings.ToLower(message)
	for _, marker := range internalMarkers {
		if strings.Contains(lowered, marker) {
			return true
		}
	}
	return false
}

// Mengembalikan pesan error service yang aman: pesan validasi statis diteruskan apa adanya,
// sedangkan pesan yang mengandung detail driver database diganti pesan generik (F-ERR-01).
// Detail asli selalu dicatat di sisi server.
func SafeBadRequest(c *gin.Context, fallbackMessage string, err error) {
	if err == nil {
		BadRequest(c, fallbackMessage)
		return
	}
	if containsInternalDetail(err.Error()) {
		FailBadRequest(c, fallbackMessage, err)
		return
	}
	BadRequest(c, err.Error())
}

// Mengembalikan pesan Not Found yang aman dengan aturan sanitasi yang sama.
func SafeNotFound(c *gin.Context, fallbackMessage string, err error) {
	if err == nil {
		NotFound(c, fallbackMessage)
		return
	}
	if containsInternalDetail(err.Error()) {
		FailNotFound(c, fallbackMessage, err)
		return
	}
	NotFound(c, err.Error())
}
