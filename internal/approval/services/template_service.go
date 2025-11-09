package services

import (
	"github.com/k8s/kube-app-operator/internal/approval/models"
	repo "github.com/k8s/kube-app-operator/internal/approval/repositories"
)



type TemplateService struct {
	repo *repo.TemplateRepo
}


func NewTemplateService(repo *repo.TemplateRepo) *TemplateService {
	return &TemplateService{repo: repo}
}


func (s *TemplateService) CreateTemplate(t *models.Template) error {
	return s.repo.Create(t)
}

func (s *TemplateService) GetTemplate(id uint) (*models.Template, error) {
	return s.repo.GetByID(id)
}

func (s *TemplateService) ListTemplates() ([]models.Template, error) {
	return s.repo.List()
}

func (s *TemplateService) UpdateTemplate(t *models.Template) error {
	return s.repo.Update(t)
}

func (s *TemplateService) DeleteTemplate(id uint) error {
	return s.repo.Delete(id)
}

