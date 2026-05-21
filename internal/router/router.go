package router

import (
	"sorint-fleet/internal/controller"
	"sorint-fleet/internal/gotenberg"
	"sorint-fleet/internal/middleware"
	"sorint-fleet/internal/repository"
	"sorint-fleet/internal/service"
	"sorint-fleet/internal/ws"

	"github.com/gin-gonic/gin"
)

func Setup() *gin.Engine {
	r := gin.Default()

	r.Use(corsMiddleware())
	gClient  := gotenberg.NewClient("")

	userRepo := repository.NewUserRepository()
	vehicleRepo := repository.NewVehicleRepository()

	refreshRepo := repository.NewRefreshTokenRepository()
	resetRepo := repository.NewPasswordResetRepository()
	emailChangeRepo := repository.NewEmailChangeRepository()

	assignmentRepo := repository.NewVehicleAssignmentRepository()
	assignmentSvc := service.NewVehicleAssignmentService(assignmentRepo)
	assignmentCtrl := controller.NewVehicleAssignmentController(assignmentSvc)

	authSvc := service.NewAuthService(userRepo, refreshRepo, resetRepo)
	vehicleSvc := service.NewVehicleService(vehicleRepo, userRepo, assignmentRepo)
	userSvc := service.NewUserService(userRepo)
	profileSvc := service.NewProfileService(userRepo, emailChangeRepo)


	pdfGen   := gotenberg.NewGenerator(gClient)
	pdfSvc   := service.NewPDFService(pdfGen, "")

	authCtrl := controller.NewAuthController(authSvc)
	vehicleCtrl := controller.NewVehicleController(vehicleSvc,pdfSvc)
	userCtrl := controller.NewUserController(userSvc)
	profileCtrl := controller.NewProfileController(profileSvc)

	r.GET("/ws", ws.ServeWS)

	v1 := r.Group("/api")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authCtrl.Register)
			auth.POST("/login", authCtrl.Login)
			auth.POST("/refresh", authCtrl.Refresh)
			auth.POST("/logout", authCtrl.Logout)
			auth.POST("/google", authCtrl.Google)
			auth.POST("/change-password", middleware.Auth(), authCtrl.ChangePassword)
			auth.POST("/request-reset", authCtrl.RequestPasswordReset)
			auth.POST("/reset-password", authCtrl.ResetPassword)
		}

		profile := v1.Group("/profile", middleware.Auth())
		{
			profile.GET("", profileCtrl.GetProfile)
			profile.PATCH("", profileCtrl.UpdateProfile)
			profile.POST("/request-email-change", profileCtrl.RequestEmailChange)
			profile.POST("/change-password", profileCtrl.ChangePassword)
			profile.POST("/disable", profileCtrl.DisableAccount)
		}

		v1.POST("/confirm-email", profileCtrl.ConfirmEmailChange)

		users := v1.Group("/users", middleware.Auth(), middleware.RequireRole("admin"))
		{
			users.GET("", userCtrl.List)
			users.GET("/pending", userCtrl.ListPending)
			users.GET("/:id", userCtrl.GetByID)
			users.PATCH("/:id/role", userCtrl.UpdateRole)
			users.POST("/:id/approve", userCtrl.Approve)
			users.POST("/:id/reject", userCtrl.Reject)
			users.POST("/:id/enable", userCtrl.Enable)
			users.POST("/:id/disable", userCtrl.Disable)
			users.GET("/:id/history", assignmentCtrl.UserHistory)
		}

		vehicles := v1.Group("/vehicles", middleware.Auth(), middleware.RequireRole("admin"))
		{
			vehicles.GET("", vehicleCtrl.List)
			vehicles.GET("/:id", vehicleCtrl.GetByID)

			vehicles.POST("", middleware.RequireRole("admin"), vehicleCtrl.Create)
			vehicles.PATCH("/:id", middleware.RequireRole("admin"), vehicleCtrl.Update)
			vehicles.PATCH("/:id/assign", middleware.RequireRole("admin"), vehicleCtrl.Assign)
			vehicles.PATCH("/:id/unassign", middleware.RequireRole("admin"), vehicleCtrl.Unassign)
			vehicles.DELETE("/:id", middleware.RequireRole("admin"), vehicleCtrl.Delete)
			vehicles.POST("/import", middleware.RequireRole("admin"), vehicleCtrl.ImportExcel)
			vehicles.GET("/:id/history", assignmentCtrl.VehicleHistory)
			vehicles.GET("/:id/assignment-pdf", vehicleCtrl.AssignmentPDF)
		}

		vehicleMeta := v1.Group("/vehicle-meta", middleware.Auth(), middleware.RequireRole("admin"))
		{
			vehicleMeta.GET("/brands", vehicleCtrl.Brands)
			vehicleMeta.GET("/models", vehicleCtrl.ModelsByBrand)
		}
	}

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
