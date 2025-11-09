package repositories

import (
	"fmt"
	"github.com/k8s/kube-app-operator/internal/approval/models"
	"gorm.io/gorm"
	"strings"
)

type TemplateRepo struct {
	db *gorm.DB
}

func NewTemplateRepo(db *gorm.DB) *TemplateRepo {
	return &TemplateRepo{db: db}
}



func (r *TemplateRepo) GetTemplateByName(name string) (*models.Template, error) {
	var template models.Template
	if err := r.db.Where("name = ?", name).First(&template).Error; err != nil {
		return nil, fmt.Errorf("获取模板失败: %v", err)
	}
	return &template, nil
}

func (r *TemplateRepo) Create(t *models.Template) error {
	return r.db.Create(t).Error
}

func (r *TemplateRepo) GetByID(id uint) (*models.Template, error) {
	var t models.Template
	err := r.db.First(&t, id).Error
	return &t, err
}

func (r *TemplateRepo) List() ([]models.Template, error) {
	var templates []models.Template
	err := r.db.Find(&templates).Error
	return templates, err
}

func (r *TemplateRepo) Update(template *models.Template) error {
	return r.db.Model(&models.Template{}).
		Where("id = ?", template.ID).
		Omit("created_at"). // 忽略这个字段
		Updates(template).Error
}
func (r *TemplateRepo) Delete(id uint) error {
	var apps []models.App
	// 查找绑定该模板的应用
	if err := r.db.Select("name").Where("template_id = ?", id).Find(&apps).Error; err != nil {
		return err
	}
	if len(apps) > 0 {
		// 拼接应用名称
		names := make([]string, len(apps))
		for i, app := range apps {
			names[i] = app.Name
		}
		return fmt.Errorf("当前有应用绑定此模板：%s，请先解除绑定后再删除", strings.Join(names, ", "))
	}
	// 没有绑定才删除
	return r.db.Delete(&models.Template{}, id).Error
}
