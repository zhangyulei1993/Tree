package app

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	adminhandler "tree/backend/internal/admin/handler"
	adminrepo "tree/backend/internal/admin/repository"
	adminservice "tree/backend/internal/admin/service"
	authhandler "tree/backend/internal/auth/handler"
	authrepo "tree/backend/internal/auth/repository"
	authservice "tree/backend/internal/auth/service"
	"tree/backend/internal/common/database"
	commonjwt "tree/backend/internal/common/jwt"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/permission"
	commonredis "tree/backend/internal/common/redis"
	"tree/backend/internal/common/response"
	"tree/backend/internal/common/wechat"
	familyhandler "tree/backend/internal/family/core/handler"
	familyrepo "tree/backend/internal/family/core/repository"
	familyservice "tree/backend/internal/family/core/service"
	operationlog "tree/backend/internal/operationlog/service"
)

func (s *Server) RegisterRoutes() {
	api := s.engine.Group("/api")
	api.GET("/health", s.health)

	s.registerAdminRoutes(api)
	s.registerUserAuthRoutes(api)
	s.registerFamilyRoutes(api)
}

func (s *Server) health(ctx response.Context) {
	response.Success(ctx, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (s *Server) registerAdminRoutes(api *gin.RouterGroup) {
	authHandler, adminAuth := s.buildAdminAuth()
	if authHandler == nil || adminAuth == nil {
		s.logger.Error("admin auth routes disabled")
		return
	}

	admin := api.Group("/admin")
	admin.POST("/auth/login", authHandler.Login)

	protected := admin.Group("")
	protected.Use(adminAuth)
	protected.POST("/auth/logout", authHandler.Logout)
	protected.GET("/me", authHandler.Me)
	protected.PUT("/me/password", authHandler.ChangePassword)
}

func (s *Server) buildAdminAuth() (*adminhandler.AuthHandler, gin.HandlerFunc) {
	ctx := context.Background()

	db, err := database.Init(ctx, s.cfg.MySQL)
	if err != nil {
		s.logger.Error("init mysql for admin auth", zap.Error(err))
		return nil, nil
	}

	jwtManager, err := commonjwt.NewManager(s.cfg.JWT, s.cfg.App.Name)
	if err != nil {
		s.logger.Error("init jwt manager", zap.Error(err))
		return nil, nil
	}

	var blacklist commonredis.TokenBlacklist = commonredis.NoopTokenBlacklist{}
	redisClient, err := commonredis.NewClient(ctx, s.cfg.Redis)
	if err != nil {
		s.logger.Warn("init redis token blacklist failed, using noop blacklist", zap.Error(err))
	} else {
		blacklist = commonredis.NewTokenBlacklist(redisClient)
	}

	repo := adminrepo.NewGormAdminUserRepository(db)
	operationLog := operationlog.NewGormService(db)
	authService := adminservice.NewAdminAuthService(repo, jwtManager, blacklist, operationLog, s.cfg.AdminSecurity)
	authHandler := adminhandler.NewAuthHandler(authService)

	return authHandler, middleware.AdminAuth(jwtManager, blacklist)
}

func (s *Server) registerUserAuthRoutes(api *gin.RouterGroup) {
	authHandler, userAuth := s.buildUserAuth()
	if authHandler == nil || userAuth == nil {
		s.logger.Error("user auth routes disabled")
		return
	}

	auth := api.Group("/auth")
	auth.POST("/send-code", authHandler.SendCode)
	auth.POST("/register-phone", authHandler.RegisterPhone)
	auth.POST("/login-phone", authHandler.LoginPhone)
	auth.POST("/wechat-mini/login", authHandler.WechatMiniLogin)

	protected := auth.Group("")
	protected.Use(userAuth)
	protected.POST("/logout", authHandler.Logout)
	protected.POST("/wechat-mini/bind-phone", authHandler.BindPhone)
	protected.POST("/change-phone", authHandler.ChangePhone)
	protected.POST("/cancel-account", authHandler.CancelAccount)
}

func (s *Server) buildUserAuth() (*authhandler.AuthHandler, gin.HandlerFunc) {
	ctx := context.Background()

	db, err := database.Init(ctx, s.cfg.MySQL)
	if err != nil {
		s.logger.Error("init mysql for user auth", zap.Error(err))
		return nil, nil
	}

	jwtManager, err := commonjwt.NewManager(s.cfg.JWT, s.cfg.App.Name)
	if err != nil {
		s.logger.Error("init jwt manager for user auth", zap.Error(err))
		return nil, nil
	}

	var blacklist commonredis.TokenBlacklist = commonredis.NoopTokenBlacklist{}
	redisClient, err := commonredis.NewClient(ctx, s.cfg.Redis)
	if err != nil {
		s.logger.Warn("init redis user token blacklist failed, using noop blacklist", zap.Error(err))
	} else {
		blacklist = commonredis.NewTokenBlacklist(redisClient)
	}

	userRepo := authrepo.NewGormUserRepository(db)
	codeRepo := authrepo.NewGormVerificationCodeRepository(db)
	operationLog := operationlog.NewGormService(db)
	wechatClient := wechat.NewMiniProgramClient(s.cfg.Wechat)
	service := authservice.NewPhoneAuthService(db, userRepo, codeRepo, wechatClient, jwtManager, blacklist, operationLog, s.cfg)
	handler := authhandler.NewAuthHandler(service)

	return handler, middleware.UserAuth(jwtManager, blacklist)
}

func (s *Server) registerFamilyRoutes(api *gin.RouterGroup) {
	handler, userAuth := s.buildFamilyCore()
	if handler == nil || userAuth == nil {
		s.logger.Error("family core routes disabled")
		return
	}

	api.GET("/families/:familyId/public", handler.PublicDetail)

	families := api.Group("/families")
	families.Use(userAuth)
	families.POST("", handler.Create)
	families.GET("", handler.List)
	families.GET("/:familyId", handler.Detail)
	families.PUT("/:familyId", handler.Update)
	families.POST("/:familyId/dissolution-requests", handler.CreateDissolutionRequest)
	families.GET("/:familyId/dissolution-requests/current", handler.CurrentDissolutionRequest)
	families.POST("/:familyId/dissolution-requests/:requestId/cancel", handler.CancelDissolutionRequest)
}

func (s *Server) buildFamilyCore() (*familyhandler.FamilyHandler, gin.HandlerFunc) {
	ctx := context.Background()
	db, err := database.Init(ctx, s.cfg.MySQL)
	if err != nil {
		s.logger.Error("init mysql for family core", zap.Error(err))
		return nil, nil
	}
	jwtManager, err := commonjwt.NewManager(s.cfg.JWT, s.cfg.App.Name)
	if err != nil {
		s.logger.Error("init jwt manager for family core", zap.Error(err))
		return nil, nil
	}

	var blacklist commonredis.TokenBlacklist = commonredis.NoopTokenBlacklist{}
	redisClient, err := commonredis.NewClient(ctx, s.cfg.Redis)
	if err != nil {
		s.logger.Warn("init redis family token blacklist failed, using noop blacklist", zap.Error(err))
	} else {
		blacklist = commonredis.NewTokenBlacklist(redisClient)
	}

	repo := familyrepo.NewFamilyRepository(db)
	permissionService := permission.NewFamilyPermissionService(db)
	service := familyservice.NewFamilyService(db, repo, permissionService)
	return familyhandler.NewFamilyHandler(service), middleware.UserAuth(jwtManager, blacklist)
}
