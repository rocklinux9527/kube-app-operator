package repositories

import (
	"errors"

	"github.com/k8s/kube-app-operator/internal/approval/models"
	"gorm.io/gorm"
)

type RoleRepo struct {
	db *gorm.DB
}

func NewRoleRepo(db *gorm.DB) *RoleRepo {

	return &RoleRepo{db: db}
}

// FindByNames 根据角色名查询角色

func (r *RoleRepo) FindByNames(names []string) ([]models.Role, error) {
    if len(names) == 0 {
        return []models.Role{}, nil
    }

    var roles []models.Role
    if err := r.db.Where("name IN ?", names).Find(&roles).Error; err != nil {
        return nil, err
    }
    return roles, nil
}


func (r *RoleRepo) GetByName(name string) (*models.Role, error) {
	var role models.Role
	if err := r.db.Where("name = ?", name).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepo) Create(role *models.Role) error {
	return r.db.Create(role).Error
}

func (r *UserRepo) UpdateRoles(user *models.User, roles []models.Role) error {
	return r.db.Model(user).Association("Roles").Replace(roles)
}

func (r *UserRepo) RemoveRoles(userID string, roleNames []string) error {
	var user models.User
	if err := r.db.First(&user, "id = ?", userID).Error; err != nil {
		return err
	}

	var roles []models.Role
	if err := r.db.Where("name IN ?", roleNames).Find(&roles).Error; err != nil {
		return err
	}

	if len(roles) == 0 {
		return errors.New("no valid roles found")
	}

	// 删除关联关系

	if err := r.db.Model(&user).Association("Roles").Delete(roles); err != nil {
		return err
	}

	return nil
}

// 判断用户是否具备某个角色

func (r *UserRepo) UserHasRole(userName, role string) (bool, error) {
	var user models.User
	if err := r.db.Preload("Roles").Where("name = ?", userName).First(&user).Error; err != nil {
		return false, err
	}
	for _, r := range user.Roles {
		if r.Name == role {
			return true, nil
		}
	}
	return false, nil
}