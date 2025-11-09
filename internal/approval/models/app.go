package models

import "time"

type App struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(64);unique;not null" json:"name"`
	Namespace   string    `json:"namespace"`
	Image       string    `json:"image"`
	Replicas    int32     `json:"replicas"`
	TemplateID  uint      `json:"template_id"`
	Template    Template  `gorm:"foreignKey:TemplateID" json:"template"`
	CreatedAt   time.Time `gorm:"<-:create" json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}


// 返回前端的简化结构

type AppResponse struct {
	ID         uint      `json:"id"`
	Name       string    `json:"name"`
	Namespace  string    `json:"namespace"`
	TemplateID uint      `json:"template_id"`
	Image     string      `json:"image"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Replicas    int32 `json:"replicas"`
}

// 转换函数：从数据库模型转换成响应模型

func (a *App) ToResponse() *AppResponse {
	return &AppResponse{
		ID:         a.ID,
		Name:       a.Name,
		Namespace:  a.Namespace,
		TemplateID: a.TemplateID,
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
		Image: a.Image,
		Replicas: a.Replicas,
	}
}
