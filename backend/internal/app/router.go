package app

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	accountquotahandler "tree/backend/internal/accountquota/handler"
	accountquotarepo "tree/backend/internal/accountquota/repository"
	accountquotaservice "tree/backend/internal/accountquota/service"
	adminhandler "tree/backend/internal/admin/handler"
	adminrepo "tree/backend/internal/admin/repository"
	adminservice "tree/backend/internal/admin/service"
	authhandler "tree/backend/internal/auth/handler"
	authrepo "tree/backend/internal/auth/repository"
	authservice "tree/backend/internal/auth/service"
	backupconfighandler "tree/backend/internal/backupconfig/handler"
	backupconfigrepo "tree/backend/internal/backupconfig/repository"
	backupconfigservice "tree/backend/internal/backupconfig/service"
	choicescenariohandler "tree/backend/internal/choicescenario/handler"
	choicescenariorepo "tree/backend/internal/choicescenario/repository"
	choicescenarioservice "tree/backend/internal/choicescenario/service"
	"tree/backend/internal/common/contentsafety"
	"tree/backend/internal/common/database"
	commonjwt "tree/backend/internal/common/jwt"
	"tree/backend/internal/common/middleware"
	"tree/backend/internal/common/permission"
	commonredis "tree/backend/internal/common/redis"
	"tree/backend/internal/common/response"
	"tree/backend/internal/common/version"
	"tree/backend/internal/common/wechat"
	contenthandler "tree/backend/internal/content/handler"
	contentrepo "tree/backend/internal/content/repository"
	contentservice "tree/backend/internal/content/service"
	capacityhandler "tree/backend/internal/family/capacity/handler"
	capacityrepo "tree/backend/internal/family/capacity/repository"
	capacityservice "tree/backend/internal/family/capacity/service"
	capacitywechatpay "tree/backend/internal/family/capacity/wechatpay"
	familyhandler "tree/backend/internal/family/core/handler"
	familyrepo "tree/backend/internal/family/core/repository"
	familyservice "tree/backend/internal/family/core/service"
	dissolutionhandler "tree/backend/internal/family/dissolution/handler"
	dissolutionrepo "tree/backend/internal/family/dissolution/repository"
	dissolutionservice "tree/backend/internal/family/dissolution/service"
	invitationhandler "tree/backend/internal/family/invitation/handler"
	invitationrepo "tree/backend/internal/family/invitation/repository"
	invitationservice "tree/backend/internal/family/invitation/service"
	invitationrewardhandler "tree/backend/internal/family/invitationreward/handler"
	invitationrewardrepo "tree/backend/internal/family/invitationreward/repository"
	invitationrewardservice "tree/backend/internal/family/invitationreward/service"
	joinhandler "tree/backend/internal/family/joinrequest/handler"
	joinrepo "tree/backend/internal/family/joinrequest/repository"
	joinservice "tree/backend/internal/family/joinrequest/service"
	memberhandler "tree/backend/internal/family/member/handler"
	memberrepo "tree/backend/internal/family/member/repository"
	memberservice "tree/backend/internal/family/member/service"
	publichandler "tree/backend/internal/family/publicdisplay/handler"
	publicrepo "tree/backend/internal/family/publicdisplay/repository"
	publicservice "tree/backend/internal/family/publicdisplay/service"
	relationshiphandler "tree/backend/internal/family/relationship/handler"
	relationshiprepo "tree/backend/internal/family/relationship/repository"
	relationshipservice "tree/backend/internal/family/relationship/service"
	rolehandler "tree/backend/internal/family/role/handler"
	rolerepo "tree/backend/internal/family/role/repository"
	roleservice "tree/backend/internal/family/role/service"
	transferhandler "tree/backend/internal/family/transfer/handler"
	transferrepo "tree/backend/internal/family/transfer/repository"
	transferservice "tree/backend/internal/family/transfer/service"
	treehandler "tree/backend/internal/family/tree/handler"
	treerepo "tree/backend/internal/family/tree/repository"
	treeservice "tree/backend/internal/family/tree/service"
	launchsplashhandler "tree/backend/internal/launchsplash/handler"
	launchsplashrepo "tree/backend/internal/launchsplash/repository"
	launchsplashservice "tree/backend/internal/launchsplash/service"
	notificationhandler "tree/backend/internal/notification/handler"
	notificationrepo "tree/backend/internal/notification/repository"
	notificationservice "tree/backend/internal/notification/service"
	operationlog "tree/backend/internal/operationlog/service"
	shareconfighandler "tree/backend/internal/shareconfig/handler"
	shareconfigrepo "tree/backend/internal/shareconfig/repository"
	shareconfigservice "tree/backend/internal/shareconfig/service"
	showcasefamilyhandler "tree/backend/internal/showcasefamily/handler"
	showcasefamilyrepo "tree/backend/internal/showcasefamily/repository"
	showcasefamilyservice "tree/backend/internal/showcasefamily/service"
	toolconfighandler "tree/backend/internal/toolconfig/handler"
	toolconfigrepo "tree/backend/internal/toolconfig/repository"
	toolconfigservice "tree/backend/internal/toolconfig/service"
)

func strictRedisEnvironment(env string) bool {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "staging", "production", "prod":
		return true
	default:
		return false
	}
}

func (s *Server) RegisterRoutes() {
	api := s.engine.Group("/api")
	api.GET("/health", s.health)
	api.GET("/version", s.version)
	s.registerShareConfigRoutes(api)
	s.registerShowcaseFamilyRoutes(api)
	s.registerToolConfigRoutes(api)
	s.registerLaunchSplashRoutes(api)

	s.registerAdminRoutes(api)
	s.registerUserAuthRoutes(api)
	s.registerChoiceScenarioRoutes(api)
	s.registerFamilyRoutes(api)
	s.registerContentRoutes(api)
	s.registerCapacityRoutes(api)
	s.registerInvitationRewardRoutes(api)
}

func (s *Server) registerInvitationRewardRoutes(api *gin.RouterGroup) {
	db, err := database.Init(context.Background(), s.cfg.MySQL)
	if err != nil {
		s.logger.Error("init mysql for invitation rewards", zap.Error(err))
		return
	}
	_, userAuth := s.buildUserAuth()
	if userAuth == nil {
		return
	}
	service := invitationrewardservice.New(db, invitationrewardrepo.New(db), permission.NewFamilyPermissionService(db), accountquotaservice.NewService(db, accountquotarepo.NewRepository(db)), operationlog.NewGormService(db))
	h := invitationrewardhandler.New(service)
	protected := api.Group("/families/:familyId/invitation-reward")
	protected.Use(userAuth)
	protected.GET("", h.Summary)
	protected.POST("/share", h.Share)
	apiProtected := api.Group("/invitation-rewards")
	apiProtected.Use(userAuth)
	apiProtected.POST("/:referralCode/qualify", h.Qualify)
}

func (s *Server) registerCapacityRoutes(api *gin.RouterGroup) {
	db, err := database.Init(context.Background(), s.cfg.MySQL)
	if err != nil {
		s.logger.Error("init mysql for capacity", zap.Error(err))
		return
	}
	_, userAuth := s.buildUserAuth()
	if userAuth == nil {
		return
	}
	jwtManager, err := commonjwt.NewManager(s.cfg.JWT, s.cfg.App.Name)
	if err != nil {
		return
	}
	var blacklist commonredis.TokenBlacklist = commonredis.NoopTokenBlacklist{}
	permissionService := permission.NewFamilyPermissionService(db)
	userRepo := authrepo.NewGormUserRepository(db)
	quotaService := accountquotaservice.NewService(db, accountquotarepo.NewRepository(db))
	logs := operationlog.NewGormService(db)
	var provider *capacitywechatpay.Provider
	if s.cfg.WechatPay.Enabled {
		provider, err = capacitywechatpay.New(context.Background(), *s.cfg)
		if err != nil {
			s.logger.Error("init wechat pay", zap.Error(err))
			return
		}
	}
	wechatClient := wechat.NewMiniProgramClient(s.cfg.Wechat)
	notifications := notificationservice.NewWithSubscribe(notificationrepo.New(db), userRepo, wechatClient, s.cfg.Wechat, s.logger)
	service := capacityservice.NewWithNotifications(db, capacityrepo.NewRepository(db), permissionService, provider, logs, userRepo, quotaService, s.cfg.Wechat.MiniAppID, s.logger, s.cfg.Capacity.OrderTimeoutMinutes, notifications)
	service.StartOrderExpiryWorker(context.Background())
	h := capacityhandler.New(service, provider, s.cfg.Wechat.MiniAppID, s.cfg.WechatPay.MerchantID)
	api.GET("/capacity-products", h.Products)
	protected := api.Group("/families/:familyId/capacity")
	protected.Use(userAuth)
	protected.GET("/orders", h.Orders)
	protected.POST("/orders", h.CreateOrder)
	protected.POST("/orders/:orderNo/pay", h.ResumePayment)
	protected.POST("/orders/:orderNo/reconcile", h.ReconcilePayment)
	protected.POST("/orders/:orderNo/refund", h.Refund)
	notificationHandler := notificationhandler.New(notifications)
	userNotifications := api.Group("/notifications")
	userNotifications.Use(userAuth)
	userNotifications.GET("", notificationHandler.List)
	userNotifications.POST("/:notificationId/read", notificationHandler.MarkRead)
	userNotifications.POST("/read-all", notificationHandler.MarkAllRead)
	_ = jwtManager
	_ = blacklist
	api.POST("/payments/wechat/notify", h.PaymentNotify)
	api.POST("/payments/wechat/refund-notify", h.RefundNotify)
}

func (s *Server) registerChoiceScenarioRoutes(api *gin.RouterGroup) {
	db, err := database.Init(context.Background(), s.cfg.MySQL)
	if err != nil {
		s.logger.Error("init mysql for choice scenarios", zap.Error(err))
		return
	}
	_, userAuth := s.buildUserAuth()
	if userAuth == nil {
		s.logger.Error("choice scenario routes disabled")
		return
	}
	userRepo := authrepo.NewGormUserRepository(db)
	logs := operationlog.NewGormService(db)
	contentSafety := contentsafety.NewService(contentsafety.NewClient(s.cfg.Wechat), userRepo, logs, s.cfg.Wechat.MiniAppID)
	service := choicescenarioservice.New(db, choicescenariorepo.NewRepository(db), contentSafety, logs)
	handler := choicescenariohandler.New(service)
	api.GET("/choice-scenarios", handler.List)
	protected := api.Group("/choice-scenarios")
	protected.Use(userAuth)
	protected.POST("", handler.Create)
	protected.POST("/:scenarioId/use", handler.Use)
}

func (s *Server) health(ctx response.Context) {
	response.Success(ctx, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (s *Server) version(ctx response.Context) {
	response.Success(ctx, http.StatusOK, version.Current(s.cfg.App.Env))
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
	managementDB, err := database.Init(context.Background(), s.cfg.MySQL)
	if err != nil {
		s.logger.Error("init mysql for admin management", zap.Error(err))
		return
	}
	management := adminhandler.NewManagementHandler(
		adminservice.NewManagementService(
			adminrepo.NewManagementRepository(managementDB, s.cfg.Wechat.MiniAppID),
			accountquotaservice.NewService(managementDB, accountquotarepo.NewRepository(managementDB)),
		),
	)
	protected.GET("/dashboard", management.Dashboard)
	protected.GET("/users", management.Users)
	protected.GET("/users/:userId", management.User)
	protected.POST("/users/:userId/status", management.SetUserStatus)
	protected.GET("/families", management.Families)
	protected.GET("/families/:familyId", management.Family)
	protected.PUT("/families/:familyId", management.UpdateFamily)
	protected.POST("/families/:familyId/status", management.SetFamilyStatus)
	protected.GET("/families/:familyId/members", management.Members)
	protected.PUT("/families/:familyId/members/:memberId", management.UpdateMember)
	protected.DELETE("/families/:familyId/members/:memberId", management.DeleteMember)
	protected.POST("/families/:familyId/members/:memberId/relatives", management.CreateRelative)
	protected.GET("/families/:familyId/tree", management.Tree)
	protected.GET("/admin-users", management.Admins)
	protected.POST("/admin-users", management.CreateAdmin)
	protected.PUT("/admin-users/:adminId", management.UpdateAdmin)
	protected.POST("/admin-users/:adminId/status", management.SetAdminStatus)
	protected.GET("/operation-logs", management.Logs)
	protected.GET("/recovery-items", management.RecoveryItems)
	protected.POST("/families/:familyId/members/:memberId/restore", management.RestoreMember)
	protected.POST("/families/:familyId/relationships/:relationshipId/restore", management.RestoreRelationship)
	var capacityProvider *capacitywechatpay.Provider
	if s.cfg.WechatPay.Enabled {
		capacityProvider, err = capacitywechatpay.New(context.Background(), *s.cfg)
		if err != nil {
			s.logger.Error("init wechat pay for admin refund review", zap.Error(err))
		}
	}
	adminWechatClient := wechat.NewMiniProgramClient(s.cfg.Wechat)
	adminUserRepo := authrepo.NewGormUserRepository(managementDB)
	adminNotifications := notificationservice.NewWithSubscribe(notificationrepo.New(managementDB), adminUserRepo, adminWechatClient, s.cfg.Wechat, s.logger)
	capacityAdminService := capacityservice.NewWithNotifications(managementDB, capacityrepo.NewRepository(managementDB), nil, capacityProvider, operationlog.NewGormService(managementDB), nil, nil, s.cfg.Wechat.MiniAppID, s.logger, s.cfg.Capacity.OrderTimeoutMinutes, adminNotifications)
	capacityAdminHandler := capacityhandler.New(capacityAdminService, capacityProvider, s.cfg.Wechat.MiniAppID, s.cfg.WechatPay.MerchantID)
	protected.GET("/capacity-products", capacityAdminHandler.AdminProducts)
	protected.PUT("/capacity-products/:productKey", capacityAdminHandler.AdminUpdateProduct)
	protected.GET("/capacity-orders", capacityAdminHandler.AdminOrders)
	protected.POST("/capacity-orders/:orderNo/reconcile", capacityAdminHandler.AdminReconcilePayment)
	protected.POST("/capacity-refunds/:refundNo/approve", capacityAdminHandler.AdminApproveRefund)
	protected.POST("/capacity-refunds/:refundNo/reject", capacityAdminHandler.AdminRejectRefund)
	protected.POST("/capacity-refunds/:refundNo/resolve", capacityAdminHandler.AdminReconcileRefund)
	protected.POST("/capacity-refunds/:refundNo/retry", capacityAdminHandler.AdminRetryRefund)
	backupConfigHandler := backupconfighandler.New(backupconfigservice.New(backupconfigrepo.New(managementDB), operationlog.NewGormService(managementDB)))
	protected.GET("/backup-operation-config", backupConfigHandler.Overview)
	protected.PUT("/backup-operation-config", backupConfigHandler.Update)
	rewardAdminService := invitationrewardservice.New(managementDB, invitationrewardrepo.New(managementDB), permission.NewFamilyPermissionService(managementDB), accountquotaservice.NewService(managementDB, accountquotarepo.NewRepository(managementDB)), operationlog.NewGormService(managementDB))
	rewardAdminHandler := invitationrewardhandler.New(rewardAdminService)
	protected.GET("/family-invitation-reward-config", rewardAdminHandler.AdminGet)
	protected.PUT("/family-invitation-reward-config", rewardAdminHandler.AdminUpdate)

	quotaHandler := s.buildQuotaHandler(managementDB)
	protected.GET("/account-quota-configs", quotaHandler.ListConfigs)
	protected.PUT("/account-quota-configs/:tier", quotaHandler.UpdateConfig)
	protected.POST("/account-quota-configs/:tier/impact-preview", quotaHandler.PreviewImpact)
	protected.GET("/account-feature-overrides/:featureKey", quotaHandler.ListFeatureOverrides)
	protected.PUT("/account-feature-overrides/:featureKey", quotaHandler.UpdateFeatureOverrides)
	protected.DELETE("/account-feature-overrides/:featureKey/:overrideId", quotaHandler.DeleteFeatureOverride)
	shareConfigHandler := s.buildShareConfigHandler(managementDB)
	protected.GET("/share-configs", shareConfigHandler.AdminList)
	protected.PUT("/share-configs/:configKey", shareConfigHandler.AdminUpdate)
	qualityConfigHandler := publichandler.NewHandler(publicservice.NewService(publicrepo.NewRepository(managementDB), nil, nil, nil))
	protected.GET("/quality-family-eligibility-config", qualityConfigHandler.AdminQualityEligibilityConfig)
	protected.PUT("/quality-family-eligibility-config", qualityConfigHandler.AdminUpdateQualityEligibilityConfig)
	toolConfigHandler := s.buildToolConfigHandler(managementDB)
	protected.GET("/tool-configs", toolConfigHandler.AdminList)
	protected.POST("/tool-configs", toolConfigHandler.AdminCreate)
	protected.PUT("/tool-configs/:toolKey", toolConfigHandler.AdminUpdate)
	launchSplashHandler := s.buildLaunchSplashHandler(managementDB)
	protected.GET("/launch-splash-config", launchSplashHandler.AdminGet)
	protected.PUT("/launch-splash-config", launchSplashHandler.AdminUpdate)
	showcaseFamilyHandler := s.buildShowcaseFamilyHandler(managementDB)
	protected.GET("/showcase-families", showcaseFamilyHandler.AdminList)
	protected.GET("/showcase-family-display-settings", showcaseFamilyHandler.AdminGetDisplaySettings)
	protected.PUT("/showcase-family-display-settings", showcaseFamilyHandler.AdminUpdateDisplaySettings)
	protected.POST("/showcase-families", showcaseFamilyHandler.AdminCreate)
	protected.PUT("/showcase-families/:showcaseFamilyId", showcaseFamilyHandler.AdminUpdate)
	protected.POST("/showcase-families/:showcaseFamilyId/publish", showcaseFamilyHandler.AdminPublish)
	protected.POST("/showcase-families/:showcaseFamilyId/unpublish", showcaseFamilyHandler.AdminUnpublish)
	protected.DELETE("/showcase-families/:showcaseFamilyId", showcaseFamilyHandler.AdminDelete)
	protected.GET("/showcase-families/:showcaseFamilyId/tree", showcaseFamilyHandler.AdminTree)
	protected.POST("/showcase-families/:showcaseFamilyId/members", showcaseFamilyHandler.AdminCreateMember)
	protected.PUT("/showcase-families/:showcaseFamilyId/members/:memberId", showcaseFamilyHandler.AdminUpdateMember)
	protected.DELETE("/showcase-families/:showcaseFamilyId/members/:memberId", showcaseFamilyHandler.AdminDeleteMember)
	protected.POST("/showcase-families/:showcaseFamilyId/members/:memberId/relatives", showcaseFamilyHandler.AdminAddRelative)
	protected.POST("/showcase-families/:showcaseFamilyId/relationships", showcaseFamilyHandler.AdminCreateRelationship)
	protected.DELETE("/showcase-families/:showcaseFamilyId/relationships/:relationshipId", showcaseFamilyHandler.AdminDeleteRelationship)
}

func (s *Server) buildLaunchSplashHandler(db *gorm.DB) *launchsplashhandler.Handler {
	return launchsplashhandler.New(launchsplashservice.New(launchsplashrepo.New(db), operationlog.NewGormService(db)))
}

func (s *Server) registerShareConfigRoutes(api *gin.RouterGroup) {
	db, err := database.Init(context.Background(), s.cfg.MySQL)
	if err != nil {
		s.logger.Error("init mysql for share configs", zap.Error(err))
		return
	}
	api.GET("/share-configs", s.buildShareConfigHandler(db).PublicList)
}

func (s *Server) buildShareConfigHandler(db *gorm.DB) *shareconfighandler.Handler {
	return shareconfighandler.New(shareconfigservice.New(shareconfigrepo.New(db), operationlog.NewGormService(db)))
}

func (s *Server) registerToolConfigRoutes(api *gin.RouterGroup) {
	db, err := database.Init(context.Background(), s.cfg.MySQL)
	if err != nil {
		s.logger.Error("init mysql for tool configs", zap.Error(err))
		return
	}
	api.GET("/tool-configs", s.buildToolConfigHandler(db).PublicList)
}

func (s *Server) registerLaunchSplashRoutes(api *gin.RouterGroup) {
	db, err := database.Init(context.Background(), s.cfg.MySQL)
	if err != nil {
		s.logger.Error("init mysql for launch splash", zap.Error(err))
		return
	}
	api.GET("/launch-splash-config", s.buildLaunchSplashHandler(db).Public)
}

func (s *Server) registerShowcaseFamilyRoutes(api *gin.RouterGroup) {
	db, err := database.Init(context.Background(), s.cfg.MySQL)
	if err != nil {
		s.logger.Error("init mysql for showcase families", zap.Error(err))
		return
	}
	api.GET("/showcase-families", s.buildShowcaseFamilyHandler(db).PublicList)
	api.GET("/showcase-families/:showcaseFamilyId/tree", s.buildShowcaseFamilyHandler(db).PublicTree)
}

func (s *Server) buildToolConfigHandler(db *gorm.DB) *toolconfighandler.Handler {
	return toolconfighandler.New(toolconfigservice.New(toolconfigrepo.New(db), operationlog.NewGormService(db)))
}

func (s *Server) buildShowcaseFamilyHandler(db *gorm.DB) *showcasefamilyhandler.Handler {
	return showcasefamilyhandler.New(showcasefamilyservice.New(showcasefamilyrepo.New(db), operationlog.NewGormService(db)))
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
		if strictRedisEnvironment(s.cfg.App.Env) {
			s.logger.Error("redis is required for admin authentication", zap.Error(err))
			return nil, nil
		}
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
	auth.POST("/login-phone", authHandler.LoginPhone)
	auth.POST("/wechat-mini/login", authHandler.WechatMiniLogin)
	auth.POST("/wechat-mini/phone-login", authHandler.WechatMiniPhoneLogin)
	if db, err := database.Init(context.Background(), s.cfg.MySQL); err == nil {
		jwtManager, jwtErr := commonjwt.NewManager(s.cfg.JWT, s.cfg.App.Name)
		if jwtErr == nil {
			blacklist := commonredis.TokenBlacklist(commonredis.NoopTokenBlacklist{})
			if client, redisErr := commonredis.NewClient(context.Background(), s.cfg.Redis); redisErr == nil {
				blacklist = commonredis.NewTokenBlacklist(client)
			}
			assist := authhandler.NewAssistanceHandler(authservice.NewAssistanceService(db, authrepo.NewGormUserRepository(db), authrepo.NewGormAssistanceCodeRepository(db), jwtManager, blacklist, operationlog.NewGormService(db), s.cfg.JWT.UserSecret))
			auth.POST("/assistance/redeem", assist.Redeem)
			protectedAssist := api.Group("/users/me/assistance-codes")
			protectedAssist.Use(userAuth)
			protectedAssist.GET("", assist.List)
			protectedAssist.POST("", assist.Generate)
			protectedAssist.POST("/:codeId/revoke", assist.Revoke)
		}
	}

	protected := auth.Group("")
	protected.Use(userAuth, middleware.RequireNoAssistance())
	protected.POST("/logout", authHandler.Logout)
	protected.POST("/wechat-mini/bind-phone", authHandler.BindWechatPhone)
	protected.POST("/cancel-account/wechat-reauth", authHandler.CancelAccountByWechatReauth)

	users := api.Group("/users/me")
	users.Use(userAuth, middleware.RequireNoAssistance())
	users.GET("", authHandler.GetMe)
	users.PATCH("/profile", authHandler.UpdateProfile)
	users.POST("/avatar", authHandler.UploadAvatar)
	if db, err := database.Init(context.Background(), s.cfg.MySQL); err == nil {
		users.GET("/capabilities", s.buildQuotaHandler(db).Capabilities)
	} else {
		s.logger.Error("init mysql for user capabilities", zap.Error(err))
	}

	api.Static("/static/avatars", authservice.AvatarStorageRoot())
}

func (s *Server) buildQuotaHandler(db *gorm.DB) *accountquotahandler.Handler {
	quotaRepo := accountquotarepo.NewRepository(db)
	quotaService := accountquotaservice.NewService(db, quotaRepo)
	return accountquotahandler.NewHandler(quotaService)
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
		if strictRedisEnvironment(s.cfg.App.Env) {
			s.logger.Error("redis is required for user authentication", zap.Error(err))
			return nil, nil
		}
		s.logger.Warn("init redis user token blacklist failed, using noop blacklist", zap.Error(err))
	} else {
		blacklist = commonredis.NewTokenBlacklist(redisClient)
	}

	userRepo := authrepo.NewGormUserRepository(db)
	codeRepo := authrepo.NewGormVerificationCodeRepository(db)
	operationLog := operationlog.NewGormService(db)
	wechatClient := wechat.NewMiniProgramClient(s.cfg.Wechat)
	contentSafety := contentsafety.NewService(contentsafety.NewClient(s.cfg.Wechat), userRepo, operationLog, s.cfg.Wechat.MiniAppID)
	service := authservice.NewPhoneAuthService(db, userRepo, codeRepo, wechatClient, jwtManager, blacklist, operationLog, contentSafety, s.cfg)
	handler := authhandler.NewAuthHandler(service)

	return handler, middleware.UserAuthWithStateLookup(jwtManager, blacklist, func(ctx context.Context, userID uint64) (middleware.UserAuthState, error) {
		user, err := userRepo.FindByID(ctx, userID)
		if err != nil {
			return middleware.UserAuthState{}, err
		}
		return middleware.UserAuthState{Status: user.Status, PhoneVerified: user.PhoneVerified}, nil
	})
}

func (s *Server) registerFamilyRoutes(api *gin.RouterGroup) {
	handler, memberHandler, relationshipHandler, treeHandler, invitationHandler, joinHandler, publicApplicationHandler, roleHandler, transferHandler, dissolutionHandler, userAuth, adminAuth := s.buildFamilyCore()
	if handler == nil || memberHandler == nil || relationshipHandler == nil || treeHandler == nil ||
		invitationHandler == nil || joinHandler == nil || publicApplicationHandler == nil ||
		roleHandler == nil || transferHandler == nil ||
		dissolutionHandler == nil || userAuth == nil || adminAuth == nil {
		s.logger.Error("family core routes disabled")
		return
	}

	api.GET("/families/:familyId/public", handler.PublicDetail)
	api.GET("/public/platform-stats", handler.PublicPlatformStats)
	api.GET("/public/families", handler.ListPublicFamilyShowcase)
	api.GET("/public/families/:familyId/tree", treeHandler.PublicTree)
	api.GET("/invitations/:inviteToken", invitationHandler.Detail)

	families := api.Group("/families")
	families.Use(userAuth, middleware.RestrictAssistanceToFamilyWork())
	families.POST("", handler.Create)
	families.GET("", handler.List)
	families.GET("/:familyId", handler.Detail)
	families.PUT("/:familyId", handler.Update)
	families.GET("/:familyId/operation-logs", handler.OperationLogs)
	families.GET("/:familyId/archive-export", handler.ExportArchive)
	families.POST("/:familyId/leave", handler.Leave)
	families.POST("/:familyId/dissolution-requests", handler.CreateDissolutionRequest)
	families.GET("/:familyId/dissolution-requests/current", handler.CurrentDissolutionRequest)
	families.POST("/:familyId/dissolution-requests/:requestId/cancel", handler.CancelDissolutionRequest)
	families.POST("/:familyId/dissolution/finalize", handler.FinalizeDissolution)
	families.POST("/:familyId/dissolution/restore", handler.RestoreDissolution)
	families.POST("/:familyId/members", memberHandler.Create)
	families.GET("/:familyId/members", memberHandler.List)
	families.GET("/:familyId/recoverable-members", memberHandler.ListRecoverable)
	families.GET("/:familyId/members/:memberId", memberHandler.Detail)
	families.PUT("/:familyId/members/:memberId", memberHandler.Update)
	families.POST("/:familyId/members/:memberId/profile-confirmation", memberHandler.ConfirmProfile)
	families.DELETE("/:familyId/members/:memberId", memberHandler.Delete)
	families.POST("/:familyId/members/:memberId/restore", memberHandler.Restore)
	families.POST("/:familyId/members/:memberId/discard-recovery", memberHandler.DiscardRecovery)
	families.POST("/:familyId/members/:memberId/bind-user", memberHandler.BindUser)
	families.POST("/:familyId/members/:memberId/unbind-user", memberHandler.UnbindUser)
	families.POST("/:familyId/members/:memberId/invite", invitationHandler.Create)
	families.GET("/:familyId/invitations", invitationHandler.ListFamily)
	families.POST("/:familyId/members/:memberId/set-admin", roleHandler.SetAdmin)
	families.POST("/:familyId/members/:memberId/unset-admin", roleHandler.UnsetAdmin)
	families.POST("/:familyId/relationships", relationshipHandler.Create)
	families.POST("/:familyId/relationships/place-existing", relationshipHandler.PlaceExisting)
	families.PUT("/:familyId/relationships/:relationshipId", relationshipHandler.Update)
	families.DELETE("/:familyId/relationships/:relationshipId", relationshipHandler.Delete)
	families.GET("/:familyId/tree", treeHandler.PrivateTree)
	families.GET("/:familyId/join-requests", joinHandler.ListFamily)
	families.POST("/:familyId/join-requests/:requestId/approve", joinHandler.Approve)
	families.POST("/:familyId/join-requests/:requestId/reject", joinHandler.Reject)
	families.POST("/:familyId/join-requests/:requestId/cancel", joinHandler.Cancel)
	families.POST("/:familyId/public-applications", publicApplicationHandler.Submit)
	families.GET("/:familyId/public-quality-eligibility", publicApplicationHandler.QualityEligibility)
	families.GET("/:familyId/public-applications", publicApplicationHandler.ListFamily)
	families.POST("/:familyId/public-applications/:applicationId/cancel", publicApplicationHandler.Cancel)
	families.POST("/:familyId/public-display/enable", publicApplicationHandler.EnableUser)
	families.POST("/:familyId/public-display/disable", publicApplicationHandler.DisableUser)
	families.POST("/:familyId/take-down-public", publicApplicationHandler.TakeDownUser)
	families.POST("/:familyId/founder-transfer-requests", transferHandler.Create)
	families.GET("/:familyId/founder-transfer-requests/current", transferHandler.Current)
	families.POST("/:familyId/founder-transfer-requests/:requestId/cancel", transferHandler.Cancel)

	invitations := api.Group("/invitations")
	invitations.Use(userAuth)
	invitations.POST("/:invitationId/accept", invitationHandler.Accept)
	invitations.POST("/:invitationId/reject", invitationHandler.Reject)
	invitations.POST("/:invitationId/cancel", invitationHandler.Cancel)
	invitations.POST("/:invitationId/regenerate", invitationHandler.Regenerate)

	users := api.Group("/users/me")
	users.Use(userAuth)
	users.GET("/invitations", invitationHandler.ListMine)
	users.GET("/join-requests", joinHandler.ListMine)

	admin := api.Group("/admin")
	admin.Use(adminAuth)
	admin.GET("/family-public-applications", publicApplicationHandler.ListAdmin)
	admin.POST("/family-public-applications/:applicationId/approve", publicApplicationHandler.Approve)
	admin.POST("/family-public-applications/:applicationId/reject", publicApplicationHandler.Reject)
	admin.POST("/families/:familyId/take-down-public", publicApplicationHandler.TakeDown)
	admin.GET("/founder-transfer-requests", transferHandler.ListAdmin)
	admin.POST("/founder-transfer-requests/:requestId/approve", transferHandler.Approve)
	admin.POST("/founder-transfer-requests/:requestId/reject", transferHandler.Reject)
	admin.GET("/dissolution-requests", dissolutionHandler.ListAdmin)
	admin.POST("/dissolution-requests/:requestId/approve", dissolutionHandler.Approve)
	admin.POST("/dissolution-requests/:requestId/reject", dissolutionHandler.Reject)
	admin.POST("/families/:familyId/restore", dissolutionHandler.Restore)
	admin.POST("/families/:familyId/finalize-dissolution", dissolutionHandler.Finalize)
}

func (s *Server) registerContentRoutes(api *gin.RouterGroup) {
	handler, adminAuth := s.buildContent()
	if handler == nil || adminAuth == nil {
		s.logger.Error("content routes disabled")
		return
	}

	content := api.Group("/content")
	content.GET("/categories", handler.PublicCategories)
	content.GET("/articles", handler.PublicArticles)
	content.GET("/articles/:articleId", handler.PublicArticleDetail)
	api.Static("/static/content", contentservice.ContentImageStorageRoot())

	admin := api.Group("/admin/content")
	admin.Use(adminAuth)
	admin.GET("/categories", handler.AdminCategories)
	admin.POST("/categories", handler.CreateCategory)
	admin.PUT("/categories/:categoryId", handler.UpdateCategory)
	admin.DELETE("/categories/:categoryId", handler.DeleteCategory)
	admin.POST("/images", handler.UploadImage)
	admin.GET("/articles", handler.AdminArticles)
	admin.POST("/articles", handler.CreateArticle)
	admin.GET("/articles/:articleId", handler.AdminArticleDetail)
	admin.PUT("/articles/:articleId", handler.UpdateArticle)
	admin.POST("/articles/:articleId/publish", handler.PublishArticle)
	admin.POST("/articles/:articleId/unpublish", handler.UnpublishArticle)
	admin.DELETE("/articles/:articleId", handler.DeleteArticle)
}

func (s *Server) buildContent() (*contenthandler.Handler, gin.HandlerFunc) {
	ctx := context.Background()

	db, err := database.Init(ctx, s.cfg.MySQL)
	if err != nil {
		s.logger.Error("init mysql for content", zap.Error(err))
		return nil, nil
	}

	jwtManager, err := commonjwt.NewManager(s.cfg.JWT, s.cfg.App.Name)
	if err != nil {
		s.logger.Error("init jwt manager for content", zap.Error(err))
		return nil, nil
	}

	var blacklist commonredis.TokenBlacklist = commonredis.NoopTokenBlacklist{}
	redisClient, err := commonredis.NewClient(ctx, s.cfg.Redis)
	if err != nil {
		if strictRedisEnvironment(s.cfg.App.Env) {
			s.logger.Error("redis is required for admin content authentication", zap.Error(err))
			return nil, nil
		}
		s.logger.Warn("init redis content token blacklist failed, using noop blacklist", zap.Error(err))
	} else {
		blacklist = commonredis.NewTokenBlacklist(redisClient)
	}

	repo := contentrepo.NewRepository(db)
	uow := contentrepo.NewUnitOfWork(db, repo)
	service := contentservice.NewService(repo, uow)
	return contenthandler.NewHandler(service), middleware.AdminAuth(jwtManager, blacklist)
}

func (s *Server) buildFamilyCore() (*familyhandler.FamilyHandler, *memberhandler.MemberHandler, *relationshiphandler.RelationshipHandler, *treehandler.TreeHandler, *invitationhandler.Handler, *joinhandler.Handler, *publichandler.Handler, *rolehandler.Handler, *transferhandler.Handler, *dissolutionhandler.Handler, gin.HandlerFunc, gin.HandlerFunc) {
	ctx := context.Background()
	db, err := database.Init(ctx, s.cfg.MySQL)
	if err != nil {
		s.logger.Error("init mysql for family core", zap.Error(err))
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil
	}
	jwtManager, err := commonjwt.NewManager(s.cfg.JWT, s.cfg.App.Name)
	if err != nil {
		s.logger.Error("init jwt manager for family core", zap.Error(err))
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil
	}

	var blacklist commonredis.TokenBlacklist = commonredis.NoopTokenBlacklist{}
	var treeCache commonredis.TreeCache = commonredis.NoopTreeCache{}
	redisClient, err := commonredis.NewClient(ctx, s.cfg.Redis)
	if err != nil {
		if strictRedisEnvironment(s.cfg.App.Env) {
			s.logger.Error("redis is required for family authentication", zap.Error(err))
			return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil
		}
		s.logger.Warn("init redis family token blacklist failed, using noop blacklist", zap.Error(err))
	} else {
		blacklist = commonredis.NewTokenBlacklist(redisClient)
		treeCache = commonredis.NewTreeCache(redisClient)
	}

	repo := familyrepo.NewFamilyRepository(db)
	userRepo := authrepo.NewGormUserRepository(db)
	wechatClient := wechat.NewMiniProgramClient(s.cfg.Wechat)
	notifications := notificationservice.NewWithSubscribe(notificationrepo.New(db), userRepo, wechatClient, s.cfg.Wechat, s.logger)
	operationLog := operationlog.NewGormService(db)
	contentSafety := contentsafety.NewService(contentsafety.NewClient(s.cfg.Wechat), userRepo, operationLog, s.cfg.Wechat.MiniAppID)
	permissionService := permission.NewFamilyPermissionService(db)
	quotaRepo := accountquotarepo.NewRepository(db)
	quotaService := accountquotaservice.NewService(db, quotaRepo)
	service := familyservice.NewFamilyServiceWithNotifications(db, repo, permissionService, quotaService, contentSafety, notifications, treeCache)
	memberRepository := memberrepo.NewMemberRepository(db)
	memberService := memberservice.NewMemberService(db, memberRepository, permissionService, quotaService, contentSafety, treeCache)
	relationshipRepository := relationshiprepo.NewRepository(db)
	relationshipUnitOfWork := relationshiprepo.NewUnitOfWork(db, relationshipRepository)
	relationshipService := relationshipservice.NewRelationshipService(relationshipRepository, relationshipUnitOfWork, permissionService, quotaService, contentSafety)
	treeRepository := treerepo.NewRepository(db)
	treeService := treeservice.NewTreeService(treeRepository, treeCache, permissionService)
	invitationRepository := invitationrepo.NewRepository(db)
	invitationUnitOfWork := invitationrepo.NewUnitOfWork(db, invitationRepository)
	invitationService := invitationservice.NewServiceWithNotifications(invitationRepository, invitationUnitOfWork, permissionService, quotaService, contentSafety, notifications, treeCache)
	joinRepository := joinrepo.NewRepository(db)
	joinUnitOfWork := joinrepo.NewUnitOfWork(db, joinRepository)
	joinRequestService := joinservice.NewServiceWithNotifications(joinRepository, joinUnitOfWork, permissionService, quotaService, contentSafety, notifications)
	publicRepository := publicrepo.NewRepository(db)
	publicUnitOfWork := publicrepo.NewUnitOfWork(db, publicRepository)
	publicApplicationService := publicservice.NewServiceWithNotifications(publicRepository, publicUnitOfWork, permissionService, contentSafety, notifications)
	roleRepository := rolerepo.NewRepository(db)
	roleUnitOfWork := rolerepo.NewUnitOfWork(db, roleRepository)
	roleService := roleservice.NewService(roleRepository, roleUnitOfWork)
	transferRepository := transferrepo.NewRepository(db)
	transferUnitOfWork := transferrepo.NewUnitOfWork(db, transferRepository)
	transferService := transferservice.NewServiceWithNotifications(transferRepository, transferUnitOfWork, quotaService, contentSafety, notifications)
	dissolutionRepository := dissolutionrepo.NewRepository(db)
	dissolutionUnitOfWork := dissolutionrepo.NewUnitOfWork(db, dissolutionRepository)
	dissolutionService := dissolutionservice.NewServiceWithNotifications(dissolutionRepository, dissolutionUnitOfWork, quotaService, s.cfg.Family.DissolutionCooldownDays, notifications)
	return familyhandler.NewFamilyHandler(service),
		memberhandler.NewMemberHandler(memberService),
		relationshiphandler.NewRelationshipHandler(relationshipService),
		treehandler.NewTreeHandler(treeService),
		invitationhandler.NewHandler(invitationService),
		joinhandler.NewHandler(joinRequestService),
		publichandler.NewHandler(publicApplicationService),
		rolehandler.NewHandler(roleService),
		transferhandler.NewHandler(transferService),
		dissolutionhandler.NewHandler(dissolutionService),
		middleware.UserAuthWithStateLookup(jwtManager, blacklist, func(ctx context.Context, userID uint64) (middleware.UserAuthState, error) {
			user, err := userRepo.FindByID(ctx, userID)
			if err != nil {
				return middleware.UserAuthState{}, err
			}
			return middleware.UserAuthState{Status: user.Status, PhoneVerified: user.PhoneVerified}, nil
		}),
		middleware.AdminAuth(jwtManager, blacklist)
}
