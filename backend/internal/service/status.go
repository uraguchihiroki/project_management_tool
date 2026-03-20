package service

import (
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

var colorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

type StatusService interface {
	Get(id uuid.UUID) (*model.Status, error)
	Create(orgID uuid.UUID, name, color, statusType string, order int) (*model.Status, error)
	Update(id uuid.UUID, name, color string, order int) (*model.Status, error)
	Delete(id uuid.UUID) error
}

func (s *statusService) Get(id uuid.UUID) (*model.Status, error) {
	return s.statusRepo.FindByID(id)
}

type statusService struct {
	statusRepo repository.StatusRepository
}

func NewStatusService(statusRepo repository.StatusRepository) StatusService {
	return &statusService{statusRepo: statusRepo}
}

func (s *statusService) Create(orgID uuid.UUID, name, color, statusType string, order int) (*model.Status, error) {
	if name == "" {
		return nil, fmt.Errorf("ステータス名は必須です")
	}
	if len(name) > 50 {
		return nil, fmt.Errorf("ステータス名は50文字以内で指定してください")
	}
	if color == "" || !colorRegex.MatchString(color) {
		return nil, fmt.Errorf("色は#RRGGBB形式で指定してください")
	}
	if statusType != "issue" && statusType != "project" {
		statusType = "issue"
	}
	statusID := uuid.New()
	key := "sts-" + statusID.String()
	status := &model.Status{
		ID:             statusID,
		Key:            key,
		OrganizationID: &orgID,
		Name:           name,
		Color:          color,
		Order:          order,
		Type:           statusType,
	}
	if err := s.statusRepo.Create(status); err != nil {
		return nil, err
	}
	return status, nil
}

func (s *statusService) Update(id uuid.UUID, name, color string, order int) (*model.Status, error) {
	status, err := s.statusRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if status.StatusKey == "sts_start" || status.StatusKey == "sts_goal" {
		return nil, fmt.Errorf("システムステータスは変更できません")
	}
	if name != "" {
		if len(name) > 50 {
			return nil, fmt.Errorf("ステータス名は50文字以内で指定してください")
		}
		status.Name = name
	}
	if color != "" {
		if !colorRegex.MatchString(color) {
			return nil, fmt.Errorf("色は#RRGGBB形式で指定してください")
		}
		status.Color = color
	}
	status.Order = order
	if err := s.statusRepo.Update(status); err != nil {
		return nil, err
	}
	return status, nil
}

func (s *statusService) Delete(id uuid.UUID) error {
	status, err := s.statusRepo.FindByID(id)
	if err != nil {
		return err
	}
	if status.StatusKey == "sts_start" || status.StatusKey == "sts_goal" {
		return fmt.Errorf("システムステータスは削除できません")
	}
	count, err := s.statusRepo.CountInUse(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("このステータスは使用中のため削除できません")
	}
	return s.statusRepo.Delete(id)
}
