package router

import (
	"sorint-fleet/internal/controller"
	"sorint-fleet/internal/gotenberg"
	"sorint-fleet/internal/middleware"
	"sorint-fleet/internal/repository"
	"sorint-fleet/internal/service"
	"sorint-fleet/internal/session"
	"sorint-fleet/internal/ws"

	"github.com/gin-gonic/gin"
)

func Setup(sessionStore *session.Store) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware())

	gClient := gotenberg.NewClient("")

	userRepo := repository.NewUserRepository()
	vehicleRepo := repository.NewVehicleRepository()
	resetRepo := repository.NewPasswordResetRepository()
	emailChangeRepo := repository.NewEmailChangeRepository()
	assignmentRepo := repository.NewVehicleAssignmentRepository()
	formRepo := repository.NewFormRepository()

	authSvc := service.NewAuthService(userRepo, resetRepo, sessionStore)
	vehicleSvc := service.NewVehicleService(vehicleRepo, userRepo, assignmentRepo)
	userSvc := service.NewUserService(userRepo, sessionStore)
	profileSvc := service.NewProfileService(userRepo, emailChangeRepo, sessionStore)
	formSvc := service.NewFormService(formRepo)
	assignmentSvc := service.NewVehicleAssignmentService(assignmentRepo)
	pdfGen := gotenberg.NewGenerator(gClient)
	pdfSvc := service.NewPDFService(pdfGen, "")

	authCtrl := controller.NewAuthController(authSvc)
	vehicleCtrl := controller.NewVehicleController(vehicleSvc, pdfSvc)
	userCtrl := controller.NewUserController(userSvc)
	profileCtrl := controller.NewProfileController(profileSvc)
	formCtrl := controller.NewFormController(formSvc)
	assignmentCtrl := controller.NewVehicleAssignmentController(assignmentSvc)
    requestCtrl := controller.NewRequestController()

	auth := middleware.Auth(sessionStore)
	adminOnly := middleware.RequireRole("admin")

	r.GET("/ws", ws.ServeWS)

	v1 := r.Group("/api")
	{
		authG := v1.Group("/auth")
		{
			authG.POST("/register", authCtrl.Register)
			authG.POST("/login", authCtrl.Login)
			authG.POST("/logout", auth, authCtrl.Logout)
			authG.POST("/google", authCtrl.Google)
			authG.POST("/change-password", auth, authCtrl.ChangePassword)
			authG.POST("/request-reset", authCtrl.RequestPasswordReset)
			authG.POST("/reset-password", authCtrl.ResetPassword)
		}

		v1.POST("/confirm-email", profileCtrl.ConfirmEmailChange)

		profileG := v1.Group("/profile", auth)
		{
			profileG.GET("", profileCtrl.GetProfile)
			profileG.PATCH("", profileCtrl.UpdateProfile)
			profileG.POST("/request-email-change", profileCtrl.RequestEmailChange)
			profileG.POST("/change-password", profileCtrl.ChangePassword)
			profileG.POST("/disable", profileCtrl.DisableAccount)
		}

		usersG := v1.Group("/users", auth, adminOnly)
		{
			usersG.GET("", userCtrl.List)
			usersG.GET("/pending", userCtrl.ListPending)
			usersG.GET("/:id", userCtrl.GetByID)
			usersG.PATCH("/:id/role", userCtrl.UpdateRole)
			usersG.POST("/:id/approve", userCtrl.Approve)
			usersG.POST("/:id/reject", userCtrl.Reject)
			usersG.POST("/:id/enable", userCtrl.Enable)
			usersG.POST("/:id/disable", userCtrl.Disable)
			usersG.GET("/:id/history", assignmentCtrl.UserHistory)
		}

		vehiclesG := v1.Group("/vehicles", auth, adminOnly)
		{
			vehiclesG.GET("", vehicleCtrl.List)
			vehiclesG.GET("/:id", vehicleCtrl.GetByID)
			vehiclesG.GET("/recommend", vehicleCtrl.Recommend)
			vehiclesG.POST("", vehicleCtrl.Create)
			vehiclesG.PATCH("/:id", vehicleCtrl.Update)
			vehiclesG.PATCH("/:id/assign", vehicleCtrl.Assign)
			vehiclesG.PATCH("/:id/unassign", vehicleCtrl.Unassign)
			vehiclesG.DELETE("/:id", vehicleCtrl.Delete)
			vehiclesG.POST("/import", vehicleCtrl.ImportExcel)
			vehiclesG.GET("/:id/history", assignmentCtrl.VehicleHistory)
			vehiclesG.GET("/:id/assignment-pdf", vehicleCtrl.AssignmentPDF)
		}

		metaG := v1.Group("/vehicle-meta", auth, adminOnly)
		{
			metaG.GET("/brands", vehicleCtrl.Brands)
			metaG.GET("/models", vehicleCtrl.ModelsByBrand)
		}

		formsG := v1.Group("/forms")
		{
			formsG.GET(":slug", formCtrl.GetTemplateBySlug)
			formsG.POST(":slug/submissions", formCtrl.SubmitPublicForm)
		}

		requestsG := v1.Group("/requests", auth)
		{
			requestsG.POST("", requestCtrl.Create)
			requestsG.GET("/search", requestCtrl.Search)
		}

		adminFormsG := v1.Group("/admin/forms", auth, adminOnly)
		{
			adminFormsG.GET("", formCtrl.ListTemplates)
			adminFormsG.POST("", formCtrl.CreateTemplate)
			adminFormsG.PATCH(":id", formCtrl.UpdateTemplate)
			adminFormsG.DELETE(":id", formCtrl.DeleteTemplate)
		}

		adminSubmissionsG := v1.Group("/admin/form-submissions", auth, adminOnly)
		{
			adminSubmissionsG.GET("", formCtrl.ListSubmissions)
			adminSubmissionsG.GET(":id", formCtrl.GetSubmissionByID)
			adminSubmissionsG.PATCH(":id/status", formCtrl.UpdateSubmissionStatus)
		}
	}

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
