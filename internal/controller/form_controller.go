package controller

import (
	"encoding/json"

	"sorint-fleet/internal/dto"
	"sorint-fleet/internal/model"
	"sorint-fleet/internal/service"
	"sorint-fleet/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FormController struct {
	formSvc service.FormService
}

func NewFormController(formSvc service.FormService) *FormController {
	return &FormController{formSvc: formSvc}
}

func (ctrl *FormController) CreateTemplate(c *gin.Context) {
	var input dto.CreateFormTemplateDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	template, err := ctrl.formSvc.CreateTemplate(input)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, template)
}

func (ctrl *FormController) UpdateTemplate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "id not valid")
		return
	}

	var input dto.UpdateFormTemplateDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	template, err := ctrl.formSvc.UpdateTemplate(id, input)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, template)
}

func (ctrl *FormController) DeleteTemplate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "id not valid")
		return
	}
	if err := ctrl.formSvc.DeleteTemplate(id); err != nil {
		response.InternalError(c, err)
		return
	}
	response.NoContent(c)
}

func (ctrl *FormController) ListTemplates(c *gin.Context) {
	templates, err := ctrl.formSvc.ListTemplates()
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, templates)
}

func (ctrl *FormController) GetTemplateBySlug(c *gin.Context) {
	slug := c.Param("slug")
	template, err := ctrl.formSvc.GetTemplateBySlug(slug)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	var fields []dto.FormFieldDto
	if err := json.Unmarshal(template.Fields, &fields); err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, gin.H{"template": template, "fields": fields})
}

func (ctrl *FormController) SubmitPublicForm(c *gin.Context) {
	slug := c.Param("slug")
	var input dto.PublicFormSubmissionDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	submission, err := ctrl.formSvc.CreateSubmission(slug, input, nil)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, submission)
}

func (ctrl *FormController) ListSubmissions(c *gin.Context) {
	params := dto.ParseListFormSubmissionsParams(c)
	submissions, total, err := ctrl.formSvc.ListSubmissions(params)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, dto.PaginatedResponse[model.FormSubmission]{
		Items:  submissions,
		Total:  total,
		Limit:  params.Limit,
		Offset: params.Offset,
	})
}

func (ctrl *FormController) GetSubmissionByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "id not valid")
		return
	}
	submission, err := ctrl.formSvc.GetSubmissionByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, submission)
}

func (ctrl *FormController) UpdateSubmissionStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "id not valid")
		return
	}
	var input dto.UpdateFormSubmissionStatusDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	submission, err := ctrl.formSvc.UpdateSubmissionStatus(id, input)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, submission)
}
