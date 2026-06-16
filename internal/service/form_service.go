package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"sorint-fleet/internal/dto"
	"sorint-fleet/internal/model"
	"sorint-fleet/internal/repository"
	"sorint-fleet/internal/search"

	"github.com/google/uuid"
)

type FormService interface {
	CreateTemplate(input dto.CreateFormTemplateDto) (*model.FormTemplate, error)
	UpdateTemplate(id uuid.UUID, input dto.UpdateFormTemplateDto) (*model.FormTemplate, error)
	DeleteTemplate(id uuid.UUID) error
	GetTemplateByID(id uuid.UUID) (*model.FormTemplate, error)
	GetTemplateBySlug(slug string) (*model.FormTemplate, error)
	ListTemplates() ([]model.FormTemplate, error)
	CreateSubmission(slug string, input dto.PublicFormSubmissionDto, submittedByID *uuid.UUID) (*model.FormSubmission, error)
	ListSubmissions(params dto.ListFormSubmissionsParams) ([]model.FormSubmission, int64, error)
	GetSubmissionByID(id uuid.UUID) (*model.FormSubmission, error)
	UpdateSubmissionStatus(id uuid.UUID, input dto.UpdateFormSubmissionStatusDto) (*model.FormSubmission, error)
}

type formService struct {
	repo repository.FormRepository
}

func NewFormService(repo repository.FormRepository) FormService {
	return &formService{repo: repo}
}

func (s *formService) CreateTemplate(input dto.CreateFormTemplateDto) (*model.FormTemplate, error) {
	if err := validateFields(input.Fields); err != nil {
		return nil, err
	}

	slug := strings.TrimSpace(input.Slug)
	if slug == "" {
		slug = generateSlug(input.Name)
	}

	existing, err := s.repo.FindTemplateBySlug(slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("slug already exists")
	}

	fieldsJSON, err := json.Marshal(input.Fields)
	if err != nil {
		return nil, err
	}

	active := true
	if input.Active != nil {
		active = *input.Active
	}

	template := &model.FormTemplate{
		Slug:        slug,
		Name:        input.Name,
		Description: input.Description,
		Active:      active,
		Fields:      fieldsJSON,
	}

	if err := s.repo.CreateTemplate(template); err != nil {
		return nil, err
	}
	return template, nil
}

func (s *formService) UpdateTemplate(id uuid.UUID, input dto.UpdateFormTemplateDto) (*model.FormTemplate, error) {
	template, err := s.repo.FindTemplateByID(id)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, errors.New("form template not found")
	}

	if input.Fields != nil {
		if err := validateFields(input.Fields); err != nil {
			return nil, err
		}
		fieldsJSON, err := json.Marshal(input.Fields)
		if err != nil {
			return nil, err
		}
		template.Fields = fieldsJSON
	}
	if input.Name != nil {
		template.Name = *input.Name
	}
	if input.Description != nil {
		template.Description = *input.Description
	}
	if input.Slug != nil {
		slug := strings.TrimSpace(*input.Slug)
		if slug == "" {
			return nil, errors.New("slug cannot be empty")
		}
		if slug != template.Slug {
			existing, err := s.repo.FindTemplateBySlug(slug)
			if err != nil {
				return nil, err
			}
			if existing != nil {
				return nil, errors.New("slug already exists")
			}
			template.Slug = slug
		}
	}
	if input.Active != nil {
		template.Active = *input.Active
	}

	if err := s.repo.UpdateTemplate(template); err != nil {
		return nil, err
	}
	return template, nil
}

func (s *formService) DeleteTemplate(id uuid.UUID) error {
	return s.repo.DeleteTemplate(id)
}

func (s *formService) GetTemplateByID(id uuid.UUID) (*model.FormTemplate, error) {
	template, err := s.repo.FindTemplateByID(id)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, errors.New("form template not found")
	}
	return template, nil
}

func (s *formService) GetTemplateBySlug(slug string) (*model.FormTemplate, error) {
	template, err := s.repo.FindTemplateBySlug(slug)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, errors.New("form template not found")
	}
	return template, nil
}

func (s *formService) ListTemplates() ([]model.FormTemplate, error) {
	return s.repo.FindAllTemplates()
}

func (s *formService) CreateSubmission(slug string, input dto.PublicFormSubmissionDto, submittedByID *uuid.UUID) (*model.FormSubmission, error) {
	template, err := s.repo.FindTemplateBySlug(slug)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, errors.New("form template not found")
	}
	if !template.Active {
		return nil, errors.New("form template is inactive")
	}

	dataJSON, err := json.Marshal(input.Data)
	if err != nil {
		return nil, err
	}

	submission := &model.FormSubmission{
		FormTemplateID: template.ID,
		FirstName:      input.FirstName,
		LastName:       input.LastName,
		Data:           dataJSON,
		Status:         model.SubmissionStatusPending,
		SubmittedByID:  submittedByID,
	}

	if err := s.repo.CreateSubmission(submission); err != nil {
		return nil, err
	}

	lname := strings.ToLower(template.Slug)
	lname2 := strings.ToLower(template.Name)
	if strings.Contains(lname, "request") || strings.Contains(lname, "change") || strings.Contains(lname2, "request") || strings.Contains(lname2, "change") {
		parts := []string{}
		if submission.FirstName != "" || submission.LastName != "" {
			parts = append(parts, submission.FirstName+" "+submission.LastName)
		}
		for _, v := range input.Data {
			switch t := v.(type) {
			case string:
				parts = append(parts, t)
			case []interface{}:
				for _, e := range t {
					parts = append(parts, fmt.Sprint(e))
				}
			default:
				parts = append(parts, fmt.Sprint(t))
			}
		}
		queryText := strings.Join(parts, " ")
		req := &model.Request{
			UserID:    submittedByID,
			QueryText: queryText,
			CreatedAt: time.Now(),
		}
		repo := repository.NewRequestRepository()
		if err := repo.Create(req); err == nil {
			_ = search.IndexRequest(*req)
		}
	}
	return submission, nil
}

func (s *formService) ListSubmissions(params dto.ListFormSubmissionsParams) ([]model.FormSubmission, int64, error) {
	return s.repo.ListSubmissions(params)
}

func (s *formService) GetSubmissionByID(id uuid.UUID) (*model.FormSubmission, error) {
	submission, err := s.repo.FindSubmissionByID(id)
	if err != nil {
		return nil, err
	}
	if submission == nil {
		return nil, errors.New("form submission not found")
	}
	return submission, nil
}

func (s *formService) UpdateSubmissionStatus(id uuid.UUID, input dto.UpdateFormSubmissionStatusDto) (*model.FormSubmission, error) {
	submission, err := s.repo.FindSubmissionByID(id)
	if err != nil {
		return nil, err
	}
	if submission == nil {
		return nil, errors.New("form submission not found")
	}

	submission.Status = model.FormSubmissionStatus(input.Status)
	if input.AdminNotes != nil {
		submission.AdminNotes = *input.AdminNotes
	}

	if err := s.repo.UpdateSubmission(submission); err != nil {
		return nil, err
	}
	return submission, nil
}

func validateFields(fields []dto.FormFieldDto) error {
	for i, field := range fields {
		fieldName := strings.TrimSpace(field.Name)
		if fieldName == "" {
			return errors.New("field name cannot be empty")
		}
		field.Label = strings.TrimSpace(field.Label)
		if field.Label == "" {
			return errors.New("field label cannot be empty")
		}
		types := map[string]bool{"text": true, "textarea": true, "select": true, "radio": true, "checkbox": true, "email": true}
		if !types[field.Type] {
			return errors.New("invalid field type for field " + fieldName)
		}
		if (field.Type == "select" || field.Type == "radio" || field.Type == "checkbox") && len(field.Options) == 0 {
			return errors.New("field options are required for " + field.Type + " field " + fieldName)
		}
		if len(field.Options) > 0 {
			for _, option := range field.Options {
				if strings.TrimSpace(option) == "" {
					return errors.New("field options cannot be empty")
				}
			}
		}
		fields[i] = field
	}
	return nil
}

func generateSlug(value string) string {
	slug := strings.ToLower(strings.TrimSpace(value))
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = uuid.NewString()
	}
	return slug
}
