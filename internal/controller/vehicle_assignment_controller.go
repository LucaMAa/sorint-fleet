package controller

import (
	"sorint-fleet/internal/service"
	"sorint-fleet/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type VehicleAssignmentController struct {
	svc service.VehicleAssignmentService
}

func NewVehicleAssignmentController(svc service.VehicleAssignmentService) *VehicleAssignmentController {
	return &VehicleAssignmentController{svc: svc}
}

// GET /vehicles/:id/history
func (ctrl *VehicleAssignmentController) VehicleHistory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "id not valid")
		return
	}
	list, err := ctrl.svc.GetByVehicle(id)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, list)
}

// GET /users/:id/history
func (ctrl *VehicleAssignmentController) UserHistory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "id not valid")
		return
	}
	list, err := ctrl.svc.GetByUser(id)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, list)
}
