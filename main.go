package main

import (
	"log"
	"release-manager/internal/config"
	"release-manager/internal/handler"
	"release-manager/internal/middleware"
	"release-manager/internal/repository"
	"release-manager/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger := config.InitLogger(cfg.LogLevel)
	defer logger.Sync()

	// 初始化数据库
	db, err := config.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// 自动迁移
	if err := repository.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化BOS客户端
	bosClient, err := config.InitBOS(cfg.BOS)
	if err != nil {
		log.Fatalf("Failed to init BOS client: %v", err)
	}

	// 初始化Repository
	repos := repository.NewRepositories(db)

	// 初始化Service
	services := service.NewServices(cfg, repos, bosClient, logger)

	// 初始化Handler
	handlers := handler.NewHandlers(services, logger)

	// 设置Gin
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(logger))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 注册路由
	registerRoutes(r, handlers, services.Auth, logger)

	// 启动服务
	logger.Info("Server starting on " + cfg.ServerAddr)
	if err := r.Run(cfg.ServerAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func registerRoutes(r *gin.Engine, h *handler.Handlers, authSvc *service.AuthService, logger *config.Logger) {
	api := r.Group("/api")

	// 公开路由
	auth := api.Group("/auth")
	{
		auth.POST("/login", h.Auth.Login)
	}

	// 需要认证的路由
	protected := api.Group("")
	protected.Use(middleware.JWTAuth(authSvc, logger))
	{
		// 认证相关
		protected.POST("/auth/logout", h.Auth.Logout)
		protected.GET("/auth/profile", h.Auth.Profile)

		// 目录和版本
		protected.GET("/directories", h.Version.GetDirectories)
		protected.GET("/directories/:id/versions", h.Version.GetVersions)
		protected.GET("/versions/:id/download-url", h.Version.GetDownloadURL)
		protected.GET("/versions/compare", h.Version.CompareVersions)

		// 编译任务
		protected.POST("/build/submit", h.Build.Submit)
		protected.GET("/build/tasks", h.Build.GetTasks)
		protected.GET("/build/tasks/:id", h.Build.GetTask)

		// 网盘
		protected.GET("/drive/personal", h.Drive.GetPersonalFiles)
		protected.GET("/drive/public", h.Drive.GetPublicFiles)
		protected.POST("/drive/upload", h.Drive.Upload)
		protected.DELETE("/drive/files/:id", h.Drive.Delete)
		protected.GET("/drive/files/:id/url", h.Drive.GetFileURL)
	}

	// 管理员路由
	admin := api.Group("/admin")
	admin.Use(middleware.JWTAuth(authSvc, logger))
	admin.Use(middleware.AdminOnly())
	{
		admin.GET("/users", h.Admin.GetUsers)
		admin.PUT("/users/:id/admin", h.Admin.SetAdmin)
		admin.POST("/directories", h.Admin.CreateDirectory)
		admin.PUT("/directories/:id", h.Admin.UpdateDirectory)
		admin.DELETE("/directories/:id", h.Admin.DeleteDirectory)
		admin.POST("/versions", h.Admin.UploadVersion)
		admin.DELETE("/versions/:id", h.Admin.DeleteVersion)
		admin.POST("/baseline", h.Admin.SetBaseline)
		admin.GET("/logs", h.Admin.GetLogs)
	}
}
