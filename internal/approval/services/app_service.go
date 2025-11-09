package services

import (
	"errors"

	"github.com/k8s/kube-app-operator/internal/approval/models"
	repo "github.com/k8s/kube-app-operator/internal/approval/repositories"
)

type AppService struct {
	appRepo      *repo.AppRepo
	templateRepo *repo.TemplateRepo
}

func NewAppService(appRepo *repo.AppRepo, templateRepo *repo.TemplateRepo) *AppService {
	return &AppService{appRepo: appRepo, templateRepo: templateRepo}
}

// 创建应用

func (s *AppService) CreateApp(a *models.App) (*models.AppResponse, error) {
	if _, err := s.templateRepo.GetByID(a.TemplateID); err != nil {
		return nil, errors.New("关联模板不存在")
	}
	if err := s.appRepo.Create(a); err != nil {
		return nil, err
	}
	return a.ToResponse(), nil
}

// 获取单个应用

func (s *AppService) GetApp(id uint) (*models.AppResponse, error) {
	app, err := s.appRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return app.ToResponse(), nil
}

// 列出所有应用

func (s *AppService) ListApps() ([]models.AppResponse, error) {
	apps, err := s.appRepo.List()
	if err != nil {
		return nil, err
	}

	responses := make([]models.AppResponse, len(apps))
	for i, a := range apps {
		responses[i] = *a.ToResponse()
	}
	return responses, nil
}

// 删除应用

func (s *AppService) DeleteApp(id uint) error {
	return s.appRepo.Delete(id)
}

// 更新应用

func (s *AppService) UpdateApp(a *models.App) (*models.AppResponse, error) {
	if a.TemplateID != 0 {
		if _, err := s.templateRepo.GetByID(a.TemplateID); err != nil {
			return nil, errors.New("关联模板不存在")
		}
	}
	if err := s.appRepo.Update(a); err != nil {
		return nil, err
	}
	// 返回最新数据
	newApp, err := s.appRepo.GetByID(a.ID)
	if err != nil {
		return nil, err
	}
	return newApp.ToResponse(), nil
}
