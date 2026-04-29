package controller

import (
	"net/http"

	"sorint-fleet/internal/service"
	"sorint-fleet/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authSvc service.AuthService
}

func NewAuthController(authSvc service.AuthService) *AuthController {
	return &AuthController{authSvc: authSvc}
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var input service.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	res, err := ctrl.authSvc.Register(input)
	if err != nil {
		if err.Error() == "email already exist" {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"token":   res.Token,
		"user":    res.User,
	})
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var input service.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	res, err := ctrl.authSvc.Login(input)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, gin.H{
		"token": res.Token,
		"user":  res.User,
	})
}
