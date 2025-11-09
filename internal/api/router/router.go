package router

import (
    "github.com/gin-gonic/gin"
    "github.com/k8s/kube-app-operator/internal/api/handler"
    "github.com/k8s/kube-app-operator/internal/middleware"
    "github.com/redis/go-redis/v9"
    "k8s.io/apimachinery/pkg/runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"

    repo "github.com/k8s/kube-app-operator/internal/approval/repositories"
    "github.com/k8s/kube-app-operator/internal/approval/services"
    "gorm.io/gorm"
)


var (
    db  *gorm.DB
    rdb *redis.Client
)

// 提供一个初始化方法给 main.go 调用

func InitDependencies(database *gorm.DB, redisClient *redis.Client) {
    db = database
    rdb = redisClient
}


// 注册所有路由

func RegisterRoutes(r *gin.Engine, k8sClient client.Client, scheme *runtime.Scheme) {
    // k8s resource create and delete
    v1 := r.Group("/api/v1", middleware.JWTAuthMiddleware())
    {
        v1.POST("/apps/create", handler.NewCreateKubeAppHandler(k8sClient, scheme))
        v1.POST("/apps/delete", handler.NewDeleteKubeAppHandler(k8sClient, scheme))
    }

    // init mysql
    r.GET("/init", handler.InitHandler)

    // ====== 初始化 service + handler ======

    // user
    userRepo := repo.NewUserRepo(db)
    roleRepo := repo.NewRoleRepo(db)
    userSvc := services.NewUserService(userRepo,roleRepo)
    userHandler := handler.NewUserHandler(userSvc)

    // template
    templateRepo := repo.NewTemplateRepo(db)
    templateSvc := services.NewTemplateService(templateRepo)
    templateHandler := handler.NewTemplateHandler(templateSvc)

    // app
    appRepo := repo.NewAppRepo(db)
    appSvc := services.NewAppService(appRepo,templateRepo)
    appHandler := handler.NewAppHandler(appSvc)

    // approval
    requestRepo := repo.NewRequestRepo(db, rdb)
    requestSvc := services.NewRequestService(requestRepo,userRepo)
    requestHandler := handler.NewRequestHandler(requestSvc)
    // 审批流接口
    approval := r.Group("/approvals",middleware.JWTAuthMiddleware())
    {
      approval.GET("/single/:id", requestHandler.FindByID)
      approval.POST("/create", requestHandler.CreateRequest)
      approval.DELETE("/delete", requestHandler.DeleteRequestList)
      approval.POST("/:id/approve", requestHandler.ApproveRequest)
      approval.POST("/:id/reject", requestHandler.RejectRequest)
       // 请求查询进度
      approval.GET("/list", requestHandler.ListRequests)
      approval.POST("/batch", requestHandler.BatchFindByIDs)
     }
    // 用户管理接口
    users := r.Group("/users")
    {
        users.POST("/create", userHandler.CreateUser)
        users.GET("/query", userHandler.ListUsers,middleware.JWTAuthMiddleware())
        users.GET("/:id", userHandler.GetUser,middleware.JWTAuthMiddleware())
        users.PUT("/update/:id", userHandler.UpdateUser)
        users.DELETE("/delete/:id", userHandler.DeleteUser,middleware.JWTAuthMiddleware())
        users.POST("/:id/roles", userHandler.AssignRolesToUser,middleware.JWTAuthMiddleware())
        users.DELETE("/:id/roles", userHandler.RemoveRoles)
        users.POST("/login", userHandler.Login)
        users.GET("/info", middleware.JWTAuthMiddleware(), userHandler.GetUserInfo)
        users.POST("/logout",handler.UserLogout)
    }
    // k8s 系统资源管理接口

    kubes := r.Group("/kube",middleware.JWTAuthMiddleware())
    {
        kubes.GET("/namespace/query",handler.ListNamespaces)
        kubes.GET("/deployment/query",handler.GetKubeDeployments)
        kubes.POST("/rollout/restart",handler.RolloutRestart)
        kubes.GET("/service/query", handler.GetKubeServices)
        kubes.GET("/ingress/query", handler.GetKubeIngress)
        kubes.GET("/pvc/query", handler.GetKubePVCS)
        kubes.GET("/pod/query", handler.GetKubePods)
        kubes.POST("/pod/restart",handler.RestartKubePod)

      //  kubes.DELETE("/:id/roles", userHandler.RemoveRoles)
    }
    // 模板管理接口（template CRUD）
    templates := r.Group("/templates",middleware.JWTAuthMiddleware())
    {
        templates.POST("/create", templateHandler.Create)
        templates.GET("/list", templateHandler.List)
        templates.GET("/:id", templateHandler.Get)
        templates.PUT("/update/:id", templateHandler.Update)
        templates.DELETE("/delete/:id", templateHandler.Delete)
    }
    // 模板管理接口（APP CRUD）
    apps := r.Group("/apps",middleware.JWTAuthMiddleware())
    {
        apps.POST("/create", appHandler.Create)
        apps.DELETE("/delete/:id", appHandler.Delete)
        apps.GET("/:id", appHandler.Get)
        apps.PUT("/update/:id", appHandler.Update)
        apps.GET("/list", appHandler.List)
    }
}

