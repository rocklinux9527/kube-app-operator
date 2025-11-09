package repositories

import (
	"github.com/k8s/kube-app-operator/internal/approval/models"
	"gorm.io/gorm"
)

type AppRepo struct {
	db *gorm.DB
}

func NewAppRepo(db *gorm.DB) *AppRepo {
	return &AppRepo{db: db}
}

func (r *AppRepo) Create(a *models.App) error {
	return r.db.Create(a).Error
}

func (r *AppRepo) GetByID(id uint) (*models.App, error) {
	var app models.App
	err := r.db.First(&app, id).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *AppRepo) List() ([]models.App, error) {
	var apps []models.App
	err := r.db.Find(&apps).Error
	return apps, err
}

func (r *AppRepo) Delete(id uint) error {
	return r.db.Delete(&models.App{}, id).Error
}

// Update: 仅更新明确允许更新的字段

func (r *AppRepo) Update(a *models.App) error {
	updates := map[string]interface{}{}

	if a.Name != "" {
		updates["name"] = a.Name
	}
	if a.Namespace != "" {
		updates["namespace"] = a.Namespace
	}
	if a.Image != "" {
		updates["image"] = a.Image
	}
	if a.Replicas != 0 {
		updates["replicas"] = a.Replicas
	}
	if a.TemplateID != 0 {
		updates["template_id"] = a.TemplateID
	}

	if len(updates) == 0 {
		return nil
	}

	return r.db.Model(&models.App{}).Where("id = ?", a.ID).Updates(updates).Error
}
