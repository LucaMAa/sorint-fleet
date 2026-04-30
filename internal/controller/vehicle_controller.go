package controller

import (
	"sorint-fleet/internal/dto"
	"sorint-fleet/internal/model"
	"sorint-fleet/internal/service"
	"sorint-fleet/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type VehicleController struct {
	vehicleSvc service.VehicleService
}

func NewVehicleController(vehicleSvc service.VehicleService) *VehicleController {
	return &VehicleController{vehicleSvc: vehicleSvc}
}

func (ctrl *VehicleController) Create(c *gin.Context) {
	var input dto.CreateVehicleDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	vehicle, err := ctrl.vehicleSvc.Create(input)
	if err != nil {
		if err.Error() == "license plate already exist" {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Created(c, vehicle)
}

func (ctrl *VehicleController) List(c *gin.Context) {
	filters := dto.ListVehiclesDto{}

	if s := c.Query("status"); s != "" {
		filters.Status = model.VehicleStatus(s)
	}
	if at := c.Query("assigned_to"); at != "" {
		uid, err := uuid.Parse(at)
		if err != nil {
			response.BadRequest(c, "assigned_to is not valid")
			return
		}
		filters.AssignedToID = &uid
	}

	vehicles, err := ctrl.vehicleSvc.List(filters)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, vehicles)
}

func (ctrl *VehicleController) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "id is not valid")
		return
	}

	vehicle, err := ctrl.vehicleSvc.GetByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, vehicle)
}

func (ctrl *VehicleController) Assign(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "id is not valid")
		return
	}

	var input dto.AssignVehicleDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	vehicle, err := ctrl.vehicleSvc.Assign(id, input)
	if err != nil {
		switch err.Error() {
		case "vehicle not found", "user not found":
			response.NotFound(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.OK(c, vehicle)
}

func (ctrl *VehicleController) Unassign(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "id is not valid")
		return
	}

	vehicle, err := ctrl.vehicleSvc.Unassign(id)
	if err != nil {
		switch err.Error() {
		case "vehicle not found":
			response.NotFound(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.OK(c, vehicle)
}

func (ctrl *VehicleController) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "id is not valid")
		return
	}

	if err := ctrl.vehicleSvc.Delete(id); err != nil {
		if err.Error() == "vehicle not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	response.NoContent(c)
}

func (vc *VehicleController) Brands(c *gin.Context) {
	brands, _ := vc.vehicleSvc.GetBrands()
	c.JSON(200, brands)
}

func (vc *VehicleController) ModelsByBrand(c *gin.Context) {
	brand := c.Query("brand")

	models, _ := vc.vehicleSvc.GetModelsByBrand(brand)
	c.JSON(200, models)
}

func (ctrl *VehicleController) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "id is not valid")
		return
	}

	var input dto.UpdateVehicleDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	vehicle, err := ctrl.vehicleSvc.Update(id, input)
	if err != nil {
		switch err.Error() {
		case "vehicle not found":
			response.NotFound(c, err.Error())
		default:
			response.InternalError(c, err)
		}
		return
	}

	response.OK(c, vehicle)
}

func (ctrl *VehicleController) ImportExcel(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}

	f, err := file.Open()
	if err != nil {
		response.InternalError(c, err)
		return
	}
	defer f.Close()

	result, err := ctrl.vehicleSvc.ImportFromExcel(f)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, result)
}
