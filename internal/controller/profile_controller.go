package controller

import (
	"sorint-fleet/internal/dto"
	"sorint-fleet/internal/middleware"
	"sorint-fleet/internal/service"
	"sorint-fleet/pkg/response"

	"github.com/gin-gonic/gin"
)

type ProfileController struct {
	profileSvc service.ProfileService
}

func NewProfileController(profileSvc service.ProfileService) *ProfileController {
	return &ProfileController{profileSvc: profileSvc}
}

func (ctrl *ProfileController) GetProfile(c *gin.Context) {
	uid, ok := middleware.GetUserID(c)
	if !ok {
		return
	}
	user, err := ctrl.profileSvc.GetProfile(uid)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, user)
}

func (ctrl *ProfileController) UpdateProfile(c *gin.Context) {
	uid, ok := middleware.GetUserID(c)
	if !ok {
		return
	}
	var input dto.UpdateProfileDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	user, err := ctrl.profileSvc.UpdateProfile(uid, input)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, user)
}

func (ctrl *ProfileController) RequestEmailChange(c *gin.Context) {
	uid, ok := middleware.GetUserID(c)
	if !ok {
		return
	}
	var input dto.RequestEmailChangeDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := ctrl.profileSvc.RequestEmailChange(uid, input); err != nil {
		if err.Error() == "email already in use" {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "Ti abbiamo inviato un link di conferma alla nuova email"})
}

func (ctrl *ProfileController) ConfirmEmailChange(c *gin.Context) {
	var body struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := ctrl.profileSvc.ConfirmEmailChange(body.Token); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "Email updated"})
}

func (ctrl *ProfileController) ChangePassword(c *gin.Context) {
	uid, ok := middleware.GetUserID(c)
	if !ok {
		return
	}
	var input dto.ChangePasswordDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := ctrl.profileSvc.ChangePassword(uid, input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "Password updated"})
}

func (ctrl *ProfileController) DisableAccount(c *gin.Context) {
	uid, ok := middleware.GetUserID(c)
	if !ok {
		return
	}
	var body struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := ctrl.profileSvc.DisableAccount(c.Request.Context(), uid, body.Password); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "Account disabled"})
}
