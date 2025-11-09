package handler

import (
	"github.com/k8s/kube-app-operator/internal/approval/config"
	"github.com/k8s/kube-app-operator/internal/approval/models"
	"github.com/k8s/kube-app-operator/internal/approval/repositories"
	"github.com/k8s/kube-app-operator/internal/approval/services"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"sync"
)



var (
    once     sync.Once
    inited   bool
    userSvc  *services.UserService
    dbConn  *gorm.DB
    apprSvc  *services.RequestService
    rdbConn  *redis.Client
)



func initMySQL(dsn string) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}


// InitHandler: 通过路由触发初始化（MySQL、Redis、Repo、Service）

func InitHandler(c *gin.Context) {
	once.Do(func() {
		cfg := config.Load()
		// MySQL
		var err error
		if dbConn, err = initMySQL(cfg.MySQLDSN); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "mysql init failed: " + err.Error()})
			return
		}

		// Redis
		rdbConn = redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})

		// AutoMigrate 你需要的表（User、Request、Approval、RequestHistory）
	       dbConn.AutoMigrate(&models.User{}, &models.Role{}, &models.UserRole{})
               if err := models.InitRoles(dbConn); err != nil {
                  panic(err)
               }
                _ = dbConn.AutoMigrate(&models.Request{}, &models.Approval{}, &models.RequestHistory{},&models.Template{},&models.App{})
		// Repos
		userRepo := repositories.NewUserRepo(dbConn)
		roleRepo := repositories.NewRoleRepo(dbConn)
		approvalRepo := repositories.NewRequestRepo(dbConn, rdbConn)

		// Services
		userSvc = services.NewUserService(userRepo,roleRepo)
		_ = services.NewRequestService(approvalRepo,userRepo)
		inited = true
	})

	c.JSON(http.StatusOK, gin.H{
		"message":       "tables initialized",
		"alreadyInited": inited,
	})
}

