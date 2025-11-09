package repositories

import (
	"context"
	"errors"
	"github.com/k8s/kube-app-operator/internal/approval/models"
	"gorm.io/gorm"
	"time"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}


// Create user 并绑定角色（事务）

func (r *UserRepo) CreateUserWithRoles(user *models.User, roleNames []string) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        // 插入用户
        if err := tx.Create(user).Error; err != nil {
            return err
        }

        // 查找角色
        var roles []models.Role
        if err := tx.Where("name IN ?", roleNames).Find(&roles).Error; err != nil {
            return err
        }
        if len(roles) == 0 {
            return errors.New("no valid roles found")
        }

        // 绑定用户和角色
        if err := tx.Model(user).Association("Roles").Append(roles); err != nil {
            return err
        }

        return nil
    })
}


// Get user 查询用户名称

func (r *UserRepo) GetByName(ctx context.Context, name string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Preload("Roles").Where("name = ?", name).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}



// Create 插入用户（要求 ID 非空）
func (r *UserRepo) Create(u *models.User) error {
	if u == nil {
		return errors.New("user nil")
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now().UTC()
	}
	u.UpdatedAt = time.Now().UTC()
	return r.db.Create(u).Error
}

// GetByID 查找单个用户，找不到返回 (nil, nil)
func (r *UserRepo) GetByID(id string) (*models.User, error) {
	var u models.User
	if err := r.db.Preload("Roles").First(&u, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}



// GetByEmail 根据邮箱查询用户
func (r *UserRepo) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Preload("Roles").Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}


// Update 更新用户记录（智能更新，只更新非零值字段）

func (r *UserRepo) Update(u *models.User) error {
	if u == nil || u.ID == "" {
		return errors.New("invalid user")
	}
	u.UpdatedAt = time.Now().UTC()
	return r.db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(u).Error
}


// UpdateUserWithRoles 更新用户和角色（支持清空角色）

func (r *UserRepo) UpdateUserWithRoles(user *models.User, roleNames []string) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        // 更新基本信息
        if err := tx.Model(&models.User{}).
            Where("id = ?", user.ID).
            Updates(map[string]interface{}{
                "name":          user.Name,
                "email":         user.Email,
                "password_hash": user.PasswordHash,
                "updated_at":    user.UpdatedAt,
            }).Error; err != nil {
            return err
        }

        // 处理角色逻辑
        if roleNames != nil { // nil 表示“不更新角色”，[] 表示“清空角色”
            var roles []models.Role
            if len(roleNames) > 0 {
                if err := tx.Where("name IN ?", roleNames).Find(&roles).Error; err != nil {
                    return err
                }
                if len(roles) != len(roleNames) {
                    return errors.New("some roles not found")
                }
            }
            // Replace 会自动清空旧的关系，写入新的（即使 roles 是空切片也能清空）
            if err := tx.Model(user).Association("Roles").Replace(&roles); err != nil {
                return err
            }
        }

        return nil
    })
}


// Delete 删除用户，先检查是否存在
func (r *UserRepo) Delete(id string) error {
    if id == "" {
        return errors.New("id empty")
    }

    var u models.User
    if err := r.db.First(&u, "id = ?", id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return errors.New("user not found")
        }
        return err
    }

    return r.db.Delete(&models.User{}, "id = ?", id).Error
}


// List 分页查询用户
func (r *UserRepo) List(page, limit int) ([]models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	var users []models.User
	var total int64
	if err := r.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := r.db.Preload("Roles").Order("created_at desc").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
