package services

import (
	"context"
	"errors"
	"github.com/k8s/kube-app-operator/internal/approval/models"
	"github.com/k8s/kube-app-operator/internal/approval/repositories"
	utils "github.com/k8s/kube-app-operator/internal/middleware"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserService struct {
    userRepo *repositories.UserRepo
    roleRepo *repositories.RoleRepo
}


func NewUserService(userRepo *repositories.UserRepo, roleRepo *repositories.RoleRepo) *UserService {
	return &UserService{
          userRepo: userRepo,
          roleRepo: roleRepo,
        }
}

//Get user name 查询用户

func (s *UserService) GetByNameUser(ctx context.Context, name string) (*models.User, error) {
    return s.userRepo.GetByName(ctx, name)
}



// 调用 repo 创建用户并绑定角色

func (s *UserService) CreateUserWithRoles(name, email, password string, roleNames []string) (*models.User, error) {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }
    // 默认角色逻辑
    if len(roleNames) == 0 {
        roleNames = []string{"ops"}
    }
    user := &models.User{
        Name:         name,
        Email:        email,
        PasswordHash: string(hashedPassword),
        CreatedAt:    time.Now().UTC(),
        UpdatedAt:    time.Now().UTC(),
    }

    // 调用 repo 完成事务
    if err := s.userRepo.CreateUserWithRoles(user, roleNames); err != nil {
        return nil, err
    }

    return user, nil
}

// Login 用户登录并返回 JWT
func (s *UserService) Login(ctx context.Context, email, password string) (string, error) {
    user, err := s.userRepo.GetByEmail(email)
    if err != nil {
        return "", err
    }
    if user == nil {
        return "", errors.New("user not found")
    }

    // 校验密码
    if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
        return "", errors.New("invalid credentials")
    }
    // 生成 JWT
    return utils.GenerateToken(user.ID, user.Email)
}

// 创建create user

func (s *UserService) CreateUser(name, email, password string, roleNames []string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var roles []models.Role
	if len(roleNames) == 0 {
		role, err := s.roleRepo.GetByName("OPS")
		if err != nil {
			return nil, err
		}
		if role == nil {
			role = &models.Role{Name: "OPS"}
			if err := s.roleRepo.Create(role); err != nil {
				return nil, err
			}
		}
		roles = append(roles, *role)
	} else {
		for _, roleName := range roleNames {
			role, err := s.roleRepo.GetByName(roleName)
			if err != nil {
				return nil, err
			}
			if role == nil {
				return nil, errors.New("role not found: " + roleName)
			}
			roles = append(roles, *role)
		}
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Roles:        roles,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}


// 验证密码（例如登录时用）

func (s *UserService) CheckPassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}



// GetUser 根据 id 获取用户
func (s *UserService) GetUser(ctx context.Context, id string) (*models.User, error) {
	return s.userRepo.GetByID(id)
}



// UpdateUser 更新用户，只更新非空字段 + 可选密码
func (s *UserService) UpdateUser(ctx context.Context, u *models.User, newPassword string, roleNames []string) (*models.User, error) {
    oldUser, err := s.userRepo.GetByID(u.ID)
    if err != nil {
        return nil, err
    }
    if oldUser == nil {
        return nil, errors.New("user not found")
    }

    // 更新密码
    if newPassword != "" {
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
        if err != nil {
            return nil, err
        }
        oldUser.PasswordHash = string(hashedPassword)
    }

    // 更新字段（只覆盖非空）
    if u.Name != "" {
        oldUser.Name = u.Name
    }
    if u.Email != "" {
        oldUser.Email = u.Email
    }

    // 更新角色
    if roleNames != nil {
        roles, err := s.roleRepo.FindByNames(roleNames)
        if err != nil {
            return nil, err
        }
        oldUser.Roles = roles
    }

    oldUser.UpdatedAt = time.Now().UTC()

    // repo 层统一保存 user + roles
    if err := s.userRepo.UpdateUserWithRoles(oldUser, roleNames); err != nil {
        return nil, err
    }

    return oldUser, nil
}


// DeleteUser 删除用户，先检查是否存在
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
    if id == "" {
        return errors.New("id empty")
    }

    user, err := s.userRepo.GetByID(id)
    if err != nil {
        return err
    }
    if user == nil {
        return errors.New("user not found")
    }
    return s.userRepo.Delete(id)
}
// ListUsers 分页查询
func (s *UserService) ListUsers(ctx context.Context, page, limit int) ([]models.User, int64, error) {
	return s.userRepo.List(page, limit)
}


// assign Role

func (s *UserService) AssignRoles(ctx context.Context, userID string, roleNames []string) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 查找这些角色
	roles, err := s.roleRepo.FindByNames(roleNames)
	if err != nil {
		return nil, err
	}

	// 设置用户角色
	if err := s.userRepo.UpdateRoles(user, roles); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) RemoveRoles(ctx context.Context, userID string, roleNames []string) error {
	if userID == "" {
		return errors.New("invalid user id")
	}
	if len(roleNames) == 0 {
		return errors.New("no roles provided")
	}

	return s.userRepo.RemoveRoles(userID, roleNames)
}

