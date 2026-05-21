package controller

import (
	"net/http"

	"sorint-fleet/internal/dto"
	"sorint-fleet/internal/model"
	"sorint-fleet/internal/service"
	"sorint-fleet/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserController struct {
	userSvc service.UserService
}

func NewUserController(userSvc service.UserService) *UserController {
	return &UserController{userSvc: userSvc}
}

func (ctrl *UserController) List(c *gin.Context) {
	params := dto.ParseListUsersParams(c)
 
	users, total, err := ctrl.userSvc.List(params)
	if err != nil {
		response.InternalError(c, err)
		return
	}
 
	response.OK(c, dto.PaginatedResponse[model.User]{
		Items:  users,
		Total:  total,
		Limit:  params.Limit,
		Offset: params.Offset,
	})
}

func (ctrl *UserController) ListPending(c *gin.Context) {
	users, err := ctrl.userSvc.ListPending()
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, gin.H{"users": users})
}

func (ctrl *UserController) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Id not valid")
		return
	}

	user, err := ctrl.userSvc.GetByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, user)
}

func (ctrl *UserController) UpdateRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Id not valid")
		return
	}

	var input service.UpdateRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := ctrl.userSvc.UpdateRole(id, input.Role)
	if err != nil {
		if err.Error() == "user not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "user": user})
}

func (ctrl *UserController) Approve(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Id not valid")
		return
	}

	user, err := ctrl.userSvc.Approve(id)
	if err != nil {
		switch err.Error() {
		case "user not found":
			response.NotFound(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.OK(c, user)
}

func (ctrl *UserController) Reject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Id not valid")
		return
	}

	user, err := ctrl.userSvc.Reject(id)
	if err != nil {
		switch err.Error() {
		case "user not found":
			response.NotFound(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.OK(c, user)
}

func (ctrl *UserController) Enable(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Id not valid")
		return
	}

	user, err := ctrl.userSvc.Enable(id)
	if err != nil {
		switch err.Error() {
		case "user not found":
			response.NotFound(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.OK(c, user)
}

func (ctrl *UserController) Disable(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Id not valid")
		return
	}

	user, err := ctrl.userSvc.Disable(id)
	if err != nil {
		switch err.Error() {
		case "user not found":
			response.NotFound(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.OK(c, user)
}
