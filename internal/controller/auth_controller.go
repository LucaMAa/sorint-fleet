package controller

import (
	"net/http"

	"sorint-fleet/internal/dto"
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
	var input dto.RegisterDto
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
	var input dto.LoginDto
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

func (ctrl *AuthController) Refresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	res, err := ctrl.authSvc.Refresh(body.RefreshToken)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, res)
}

func (ctrl *AuthController) Logout(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.authSvc.Logout(body.RefreshToken); err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, gin.H{
		"message": "logged out",
	})
}

func (ctrl *AuthController) Google(c *gin.Context) {
	var input dto.GoogleAuthDto

	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	res, err := ctrl.authSvc.GoogleLogin(input.Token)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, gin.H{
		"token": res.Token,
		"user":  res.User,
	})
}
