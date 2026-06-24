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
	contenthandler "tree/backend/internal/content/handler"
	contentrepo "tree/backend/internal/content/repository"
	contentservice "tree/backend/internal/content/service"
	familyhandler "tree/backend/internal/family/core/handler"
	familyrepo "tree/backend/internal/family/core/repository"
	familyservice "tree/backend/internal/family/core/service"
	dissolutionhandler "tree/backend/internal/family/dissolution/handler"
	dissolutionrepo "tree/backend/internal/family/dissolution/repository"
	dissolutionservice "tree/backend/internal/family/dissolution/service"
	invitationhandler "tree/backend/internal/family/invitation/handler"
	invitationrepo "tree/backend/internal/family/invitation/repository"
	invitationservice "tree/backend/internal/family/invitation/service"
	joinhandler "tree/backend/internal/family/joinrequest/handler"
	joinrepo "tree/backend/internal/family/joinrequest/repository"
	joinservice "tree/backend/internal/family/joinrequest/service"
	memberhandler "tree/backend/internal/family/member/handler"
	memberrepo "tree/backend/internal/family/member/repository"
	memberservice "tree/backend/internal/family/member/service"
	messagehandler "tree/backend/internal/family/message/handler"
	messagerepo "tree/backend/internal/family/message/repository"
	messageservice "tree/backend/internal/family/message/service"
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
	operationlog "tree/backend/internal/operationlog/service"
)

func (s *Server) RegisterRoutes() {
	api := s.engine.Group("/api")
	api.GET("/health", s.health)

	s.registerAdminRoutes(api)
	s.registerUserAuthRoutes(api)
	s.registerFamilyRoutes(api)
	s.registerContentRoutes(api)
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
	auth.POST("/wechat-mini/phone-login", authHandler.WechatMiniPhoneLogin)

	protected := auth.Group("")
	protected.Use(userAuth)
	protected.POST("/logout", authHandler.Logout)
	protected.POST("/wechat-mini/bind-phone", authHandler.BindPhone)
	protected.POST("/change-phone", authHandler.ChangePhone)
	protected.POST("/cancel-account", authHandler.CancelAccount)

	users := api.Group("/users/me")
	users.Use(userAuth)
	users.GET("", authHandler.GetMe)
	users.PATCH("/profile", authHandler.UpdateProfile)
	users.POST("/avatar", authHandler.UploadAvatar)

	api.Static("/static/avatars", authservice.AvatarStorageRoot())
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
	handler, memberHandler, relationshipHandler, treeHandler, invitationHandler, joinHandler, publicApplicationHandler, visitorMessageHandler, roleHandler, transferHandler, dissolutionHandler, userAuth, adminAuth := s.buildFamilyCore()
	if handler == nil || memberHandler == nil || relationshipHandler == nil || treeHandler == nil ||
		invitationHandler == nil || joinHandler == nil || publicApplicationHandler == nil ||
		visitorMessageHandler == nil || roleHandler == nil || transferHandler == nil ||
		dissolutionHandler == nil || userAuth == nil || adminAuth == nil {
		s.logger.Error("family core routes disabled")
		return
	}

	api.GET("/families/:familyId/public", handler.PublicDetail)
	api.GET("/public/families", handler.ListPublicFamilies)
	api.GET("/public/families/:familyId/tree", treeHandler.PublicTree)
	api.POST("/public/families/:familyId/visitor-messages", visitorMessageHandler.CreatePublic)
	api.GET("/public/families/:familyId/visitor-messages", visitorMessageHandler.ListPublic)
	api.GET("/invitations/:inviteToken", invitationHandler.Detail)

	families := api.Group("/families")
	families.Use(userAuth)
	families.POST("", handler.Create)
	families.GET("", handler.List)
	families.GET("/:familyId", handler.Detail)
	families.PUT("/:familyId", handler.Update)
	families.POST("/:familyId/dissolution-requests", handler.CreateDissolutionRequest)
	families.GET("/:familyId/dissolution-requests/current", handler.CurrentDissolutionRequest)
	families.POST("/:familyId/dissolution-requests/:requestId/cancel", handler.CancelDissolutionRequest)
	families.POST("/:familyId/members", memberHandler.Create)
	families.GET("/:familyId/members", memberHandler.List)
	families.GET("/:familyId/members/:memberId", memberHandler.Detail)
	families.PUT("/:familyId/members/:memberId", memberHandler.Update)
	families.DELETE("/:familyId/members/:memberId", memberHandler.Delete)
	families.POST("/:familyId/members/:memberId/bind-user", memberHandler.BindUser)
	families.POST("/:familyId/members/:memberId/unbind-user", memberHandler.UnbindUser)
	families.POST("/:familyId/members/:memberId/invite", invitationHandler.Create)
	families.POST("/:familyId/members/:memberId/set-admin", roleHandler.SetAdmin)
	families.POST("/:familyId/members/:memberId/unset-admin", roleHandler.UnsetAdmin)
	families.POST("/:familyId/relationships", relationshipHandler.Create)
	families.PUT("/:familyId/relationships/:relationshipId", relationshipHandler.Update)
	families.DELETE("/:familyId/relationships/:relationshipId", relationshipHandler.Delete)
	families.GET("/:familyId/tree", treeHandler.PrivateTree)
	families.POST("/:familyId/join-requests", joinHandler.Create)
	families.GET("/:familyId/join-requests", joinHandler.ListFamily)
	families.POST("/:familyId/join-requests/:requestId/approve", joinHandler.Approve)
	families.POST("/:familyId/join-requests/:requestId/reject", joinHandler.Reject)
	families.POST("/:familyId/join-requests/:requestId/cancel", joinHandler.Cancel)
	families.POST("/:familyId/public-applications", publicApplicationHandler.Submit)
	families.GET("/:familyId/public-applications", publicApplicationHandler.ListFamily)
	families.POST("/:familyId/public-applications/:applicationId/cancel", publicApplicationHandler.Cancel)
	families.POST("/:familyId/founder-transfer-requests", transferHandler.Create)
	families.POST("/:familyId/founder-transfer-requests/:requestId/cancel", transferHandler.Cancel)

	invitations := api.Group("/invitations")
	invitations.Use(userAuth)
	invitations.POST("/:invitationId/accept", invitationHandler.Accept)
	invitations.POST("/:invitationId/reject", invitationHandler.Reject)
	invitations.POST("/:invitationId/cancel", invitationHandler.Cancel)

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
	admin.GET("/visitor-messages", visitorMessageHandler.ListAdmin)
	admin.POST("/visitor-messages/:messageId/approve", visitorMessageHandler.Approve)
	admin.POST("/visitor-messages/:messageId/reject", visitorMessageHandler.Reject)
	admin.DELETE("/visitor-messages/:messageId", visitorMessageHandler.Delete)
	admin.GET("/founder-transfer-requests", transferHandler.ListAdmin)
	admin.POST("/founder-transfer-requests/:requestId/approve", transferHandler.Approve)
	admin.POST("/founder-transfer-requests/:requestId/reject", transferHandler.Reject)
	admin.GET("/dissolution-requests", dissolutionHandler.ListAdmin)
	admin.POST("/dissolution-requests/:requestId/approve", dissolutionHandler.Approve)
	admin.POST("/dissolution-requests/:requestId/reject", dissolutionHandler.Reject)
	admin.POST("/families/:familyId/restore", dissolutionHandler.Restore)
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

	admin := api.Group("/admin/content")
	admin.Use(adminAuth)
	admin.GET("/categories", handler.AdminCategories)
	admin.POST("/categories", handler.CreateCategory)
	admin.PUT("/categories/:categoryId", handler.UpdateCategory)
	admin.DELETE("/categories/:categoryId", handler.DeleteCategory)
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
		s.logger.Warn("init redis content token blacklist failed, using noop blacklist", zap.Error(err))
	} else {
		blacklist = commonredis.NewTokenBlacklist(redisClient)
	}

	repo := contentrepo.NewRepository(db)
	uow := contentrepo.NewUnitOfWork(db, repo)
	service := contentservice.NewService(repo, uow)
	return contenthandler.NewHandler(service), middleware.AdminAuth(jwtManager, blacklist)
}

func (s *Server) buildFamilyCore() (*familyhandler.FamilyHandler, *memberhandler.MemberHandler, *relationshiphandler.RelationshipHandler, *treehandler.TreeHandler, *invitationhandler.Handler, *joinhandler.Handler, *publichandler.Handler, *messagehandler.Handler, *rolehandler.Handler, *transferhandler.Handler, *dissolutionhandler.Handler, gin.HandlerFunc, gin.HandlerFunc) {
	ctx := context.Background()
	db, err := database.Init(ctx, s.cfg.MySQL)
	if err != nil {
		s.logger.Error("init mysql for family core", zap.Error(err))
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil
	}
	jwtManager, err := commonjwt.NewManager(s.cfg.JWT, s.cfg.App.Name)
	if err != nil {
		s.logger.Error("init jwt manager for family core", zap.Error(err))
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil
	}

	var blacklist commonredis.TokenBlacklist = commonredis.NoopTokenBlacklist{}
	var treeCache commonredis.TreeCache = commonredis.NoopTreeCache{}
	redisClient, err := commonredis.NewClient(ctx, s.cfg.Redis)
	if err != nil {
		s.logger.Warn("init redis family token blacklist failed, using noop blacklist", zap.Error(err))
	} else {
		blacklist = commonredis.NewTokenBlacklist(redisClient)
		treeCache = commonredis.NewTreeCache(redisClient)
	}

	repo := familyrepo.NewFamilyRepository(db)
	permissionService := permission.NewFamilyPermissionService(db)
	service := familyservice.NewFamilyService(db, repo, permissionService)
	memberRepository := memberrepo.NewMemberRepository(db)
	memberService := memberservice.NewMemberService(db, memberRepository, permissionService)
	relationshipRepository := relationshiprepo.NewRepository(db)
	relationshipUnitOfWork := relationshiprepo.NewUnitOfWork(db, relationshipRepository)
	relationshipService := relationshipservice.NewRelationshipService(relationshipRepository, relationshipUnitOfWork, permissionService)
	treeRepository := treerepo.NewRepository(db)
	treeService := treeservice.NewTreeService(treeRepository, treeCache, permissionService)
	invitationRepository := invitationrepo.NewRepository(db)
	invitationUnitOfWork := invitationrepo.NewUnitOfWork(db, invitationRepository)
	invitationService := invitationservice.NewService(invitationRepository, invitationUnitOfWork, permissionService)
	joinRepository := joinrepo.NewRepository(db)
	joinUnitOfWork := joinrepo.NewUnitOfWork(db, joinRepository)
	joinRequestService := joinservice.NewService(joinRepository, joinUnitOfWork, permissionService)
	publicRepository := publicrepo.NewRepository(db)
	publicUnitOfWork := publicrepo.NewUnitOfWork(db, publicRepository)
	publicApplicationService := publicservice.NewService(publicRepository, publicUnitOfWork, permissionService)
	messageRepository := messagerepo.NewRepository(db)
	messageUnitOfWork := messagerepo.NewUnitOfWork(db, messageRepository)
	visitorMessageService := messageservice.NewService(messageRepository, messageUnitOfWork)
	roleRepository := rolerepo.NewRepository(db)
	roleUnitOfWork := rolerepo.NewUnitOfWork(db, roleRepository)
	roleService := roleservice.NewService(roleRepository, roleUnitOfWork)
	transferRepository := transferrepo.NewRepository(db)
	transferUnitOfWork := transferrepo.NewUnitOfWork(db, transferRepository)
	transferService := transferservice.NewService(transferRepository, transferUnitOfWork)
	dissolutionRepository := dissolutionrepo.NewRepository(db)
	dissolutionUnitOfWork := dissolutionrepo.NewUnitOfWork(db, dissolutionRepository)
	dissolutionService := dissolutionservice.NewService(dissolutionRepository, dissolutionUnitOfWork)
	return familyhandler.NewFamilyHandler(service),
		memberhandler.NewMemberHandler(memberService),
		relationshiphandler.NewRelationshipHandler(relationshipService),
		treehandler.NewTreeHandler(treeService),
		invitationhandler.NewHandler(invitationService),
		joinhandler.NewHandler(joinRequestService),
		publichandler.NewHandler(publicApplicationService),
		messagehandler.NewHandler(visitorMessageService),
		rolehandler.NewHandler(roleService),
		transferhandler.NewHandler(transferService),
		dissolutionhandler.NewHandler(dissolutionService),
		middleware.UserAuth(jwtManager, blacklist),
		middleware.AdminAuth(jwtManager, blacklist)
}
