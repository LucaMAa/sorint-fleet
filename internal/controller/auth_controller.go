package controller

import (
	"net/http"

	"sorint-fleet/internal/dto"
	"sorint-fleet/internal/middleware"
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
	if err := ctrl.authSvc.Register(input); err != nil {
		if err.Error() == "email already exist" {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Registrazione ricevuta. Il tuo account sarà attivo dopo l'approvazione dell'amministratore.",
	})
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var input dto.LoginDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	res, err := ctrl.authSvc.Login(c.Request.Context(), input, c.Writer)
	if err != nil {
		switch err.Error() {
		case "account_pending":
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "account_pending"})
		case "account_rejected":
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "account_rejected"})
		case "account_disabled":
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "account_disabled"})
		default:
			response.Unauthorized(c, err.Error())
		}
		return
	}
	response.OK(c, gin.H{
		"user":                 res.User,
		"must_change_password": res.MustChangePassword,
	})
}

func (ctrl *AuthController) Logout(c *gin.Context) {
	sessionID := middleware.GetSessionID(c)
	if err := ctrl.authSvc.Logout(c.Request.Context(), sessionID, c.Writer); err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "logged out"})
}

func (ctrl *AuthController) Google(c *gin.Context) {
	var input dto.GoogleAuthDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	res, err := ctrl.authSvc.GoogleLogin(c.Request.Context(), input.Token, c.Writer)
	if err != nil {
		switch err.Error() {
		case "account_pending":
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "account_pending"})
		case "account_rejected":
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "account_rejected"})
		case "account_disabled":
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "account_disabled"})
		default:
			response.Unauthorized(c, err.Error())
		}
		return
	}
	response.OK(c, gin.H{"user": res.User})
}

func (ctrl *AuthController) ChangePassword(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return
	}
	var input dto.ChangePasswordDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := ctrl.authSvc.ChangePassword(c.Request.Context(), userID, input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "Password aggiornata con successo"})
}

func (ctrl *AuthController) RequestPasswordReset(c *gin.Context) {
	var body struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	_ = ctrl.authSvc.RequestPasswordReset(body.Email)
	response.OK(c, gin.H{"message": "Se l'email esiste riceverai le istruzioni"})
}

func (ctrl *AuthController) ResetPassword(c *gin.Context) {
	var body struct {
		Token       string `json:"token"        binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := ctrl.authSvc.ResetPassword(body.Token, body.NewPassword); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "Password aggiornata con successo"})
}
