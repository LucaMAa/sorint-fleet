package dto

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FormFieldDto struct {
	Name        string   `json:"name" binding:"required"`
	Label       string   `json:"label" binding:"required"`
	Type        string   `json:"type" binding:"required,oneof=text textarea select radio checkbox email"`
	Required    bool     `json:"required"`
	Options     []string `json:"options,omitempty"`
	Placeholder string   `json:"placeholder,omitempty"`
}

type CreateFormTemplateDto struct {
	Name        string        `json:"name" binding:"required"`
	Slug        string        `json:"slug"`
	Description string        `json:"description"`
	Active      *bool         `json:"active"`
	Fields      []FormFieldDto `json:"fields" binding:"required,min=1"`
}

type UpdateFormTemplateDto struct {
	Name        *string        `json:"name"`
	Slug        *string        `json:"slug"`
	Description *string        `json:"description"`
	Active      *bool          `json:"active"`
	Fields      []FormFieldDto `json:"fields"`
}

type PublicFormSubmissionDto struct {
	FirstName string                 `json:"first_name"`
	LastName  string                 `json:"last_name"`
	Data      map[string]any         `json:"data" binding:"required"`
}

type UpdateFormSubmissionStatusDto struct {
	Status     string  `json:"status" binding:"required,oneof=pending approved rejected"`
	AdminNotes *string `json:"admin_notes"`
}

type ListFormSubmissionsParams struct {
	PageParams
	Status         string      `json:"status"`
	FormTemplateID *uuid.UUID  `json:"form_template_id,omitempty"`
	Search         string      `json:"search"`
}

func ParseListFormSubmissionsParams(c *gin.Context) ListFormSubmissionsParams {
	params := ListFormSubmissionsParams{PageParams: ParsePageParams(c)}
	params.Status = c.Query("status")
	if id := c.Query("form_template_id"); id != "" {
		if uid, err := uuid.Parse(id); err == nil {
			params.FormTemplateID = &uid
		}
	}
	params.Search = c.Query("search")
	return params
}
