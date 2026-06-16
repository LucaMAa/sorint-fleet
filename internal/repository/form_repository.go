package repository

import (
	"errors"

	"sorint-fleet/internal/config"
	"sorint-fleet/internal/dto"
	"sorint-fleet/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FormRepository interface {
	CreateTemplate(template *model.FormTemplate) error
	UpdateTemplate(template *model.FormTemplate) error
	FindTemplateByID(id uuid.UUID) (*model.FormTemplate, error)
	FindTemplateBySlug(slug string) (*model.FormTemplate, error)
	FindAllTemplates() ([]model.FormTemplate, error)
	DeleteTemplate(id uuid.UUID) error

	CreateSubmission(submission *model.FormSubmission) error
	UpdateSubmission(submission *model.FormSubmission) error
	FindSubmissionByID(id uuid.UUID) (*model.FormSubmission, error)
	ListSubmissions(params dto.ListFormSubmissionsParams) ([]model.FormSubmission, int64, error)
}

type formRepository struct {
	db *gorm.DB
}

func NewFormRepository() FormRepository {
	return &formRepository{db: config.DB}
}

func (r *formRepository) CreateTemplate(template *model.FormTemplate) error {
	return r.db.Create(template).Error
}

func (r *formRepository) UpdateTemplate(template *model.FormTemplate) error {
	return r.db.Save(template).Error
}

func (r *formRepository) FindTemplateByID(id uuid.UUID) (*model.FormTemplate, error) {
	var template model.FormTemplate
	err := r.db.First(&template, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &template, err
}

func (r *formRepository) FindTemplateBySlug(slug string) (*model.FormTemplate, error) {
	var template model.FormTemplate
	err := r.db.Where("slug = ?", slug).First(&template).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &template, err
}

func (r *formRepository) FindAllTemplates() ([]model.FormTemplate, error) {
	var templates []model.FormTemplate
	err := r.db.Order("created_at DESC").Find(&templates).Error
	return templates, err
}

func (r *formRepository) DeleteTemplate(id uuid.UUID) error {
	return r.db.Delete(&model.FormTemplate{}, "id = ?", id).Error
}

func (r *formRepository) CreateSubmission(submission *model.FormSubmission) error {
	return r.db.Create(submission).Error
}

func (r *formRepository) UpdateSubmission(submission *model.FormSubmission) error {
	return r.db.Save(submission).Error
}

func (r *formRepository) FindSubmissionByID(id uuid.UUID) (*model.FormSubmission, error) {
	var submission model.FormSubmission
	err := r.db.Preload("FormTemplate").First(&submission, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &submission, err
}

func (r *formRepository) ListSubmissions(params dto.ListFormSubmissionsParams) ([]model.FormSubmission, int64, error) {
	var submissions []model.FormSubmission
	var total int64

	q := r.db.Preload("FormTemplate").Model(&model.FormSubmission{})
	if params.Status != "" {
		q = q.Where("status = ?", params.Status)
	}
	if params.FormTemplateID != nil {
		q = q.Where("form_template_id = ?", *params.FormTemplateID)
	}
	if params.Search != "" {
		like := "%" + params.Search + "%"
		q = q.Where("first_name ILIKE ? OR last_name ILIKE ? OR admin_notes ILIKE ?", like, like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}

	err := q.Order("created_at DESC").Limit(limit).Offset(params.Offset).Find(&submissions).Error
	return submissions, total, err
}
