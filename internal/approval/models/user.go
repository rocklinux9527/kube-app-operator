package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)
type User struct {
	ID           string    `gorm:"column:id;primaryKey;size:36" json:"id"`
	Name         string    `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Email        string    `gorm:"column:email;type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(255);not null" json:"-"` // 不在 API 返回中展示
	Roles        []Role `gorm:"many2many:user_roles;" json:"roles"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}


func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return
}

type Role struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	Name      string    `gorm:"uniqueIndex;size:64" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Users []User `gorm:"many2many:user_roles;" json:"-"`
}

func (r *Role) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return
}


type UserRole struct {
	UserID string `gorm:"primaryKey"`
	RoleID string `gorm:"primaryKey"`
}

// InitRoles 预置角色
func InitRoles(db *gorm.DB) error {
	roles := []string{"ADMIN", "OPS", "SRE", "K8S"}
	for _, roleName := range roles {
		var count int64
		if err := db.Model(&Role{}).Where("name = ?", roleName).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := db.Create(&Role{
				ID:   uuid.New().String(),
				Name: roleName,
			}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
