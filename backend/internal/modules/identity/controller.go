package identity

import (
	"errors"
	"net/http"

	"dcisp/backend/internal/middleware"
	"dcisp/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func (ctrl *Controller) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid login request payload", err.Error())
		return
	}

	tokens, profile, err := ctrl.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Unauthorized(c, "Invalid email or password")
			return
		}
		response.InternalError(c, "Failed to authenticate user", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Login successful", gin.H{
		"tokens":  tokens,
		"profile": profile,
	})
}

func (ctrl *Controller) GetMe(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		response.Unauthorized(c, "Unauthenticated user")
		return
	}

	userID := userIDVal.(uuid.UUID)
	profile, err := ctrl.service.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.NotFound(c, "User profile not found")
			return
		}
		response.InternalError(c, "Failed to retrieve user profile", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "User profile retrieved successfully", profile)
}

func (ctrl *Controller) GetRoles(c *gin.Context) {
	roles, err := ctrl.service.GetRoles(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to retrieve system roles", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Roles retrieved successfully", roles)
}
