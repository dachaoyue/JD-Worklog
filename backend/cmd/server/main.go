package main

import (
	"log"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/example/worklog-system/internal/config"
	"github.com/example/worklog-system/internal/db"
	"github.com/example/worklog-system/internal/handlers"
	"github.com/example/worklog-system/internal/middleware"
	"github.com/example/worklog-system/internal/seed"
)

func main() {
	cfg := config.Load()
	dbConn := db.Connect(cfg)
	seed.Run(dbConn)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	jwt := middleware.NewJWT(cfg.JWTSecret)
	authH := &handlers.AuthHandler{DB: dbConn, JWT: jwt}
	tfaH := &handlers.TwoFactorAuthHandler{DB: dbConn, JWT: jwt}
	projH := &handlers.ProjectHandler{DB: dbConn}
	tsH := &handlers.TimesheetHandler{DB: dbConn}
	repH := &handlers.ReportHandler{DB: dbConn}
	acctH := &handlers.AccountHandler{DB: dbConn}
	usersH := &handlers.UsersHandler{DB: dbConn}
	aiH := &handlers.AIHandler{DB: dbConn, Cfg: cfg}
	scheduleH := &handlers.WorkScheduleHandler{DB: dbConn}

	api := r.Group("/api")
	{
		api.POST("/login", authH.Login)

		auth := api.Group("")
		auth.Use(jwt.AuthRequired())
		{
			auth.POST("/login/2fa", authH.Login2FA)
			auth.POST("/2fa/setup", tfaH.Setup)
			auth.POST("/2fa/verify", tfaH.Verify)

			tfa := auth.Group("")
			tfa.Use(jwt.TFARequired())
			{
				admin := tfa.Group("/admin")
				admin.Use(middleware.RequireRole("admin"))
				{
					admin.GET("/allprojects", projH.AllList)
					admin.POST("/projects", projH.Create)
					admin.PUT("/projects/:id", projH.Update)
					admin.DELETE("/projects/:id", projH.Delete)
					admin.POST("/reports/project-totals", repH.ProjectTotals)
					// CSV 导出
					admin.GET("/reports/project-totals.csv", repH.ProjectTotalsCSV)
					admin.GET("/reports/project-export-xlsx", repH.ProjectExportXLSX)
					admin.POST("/users", usersH.Create)
					admin.GET("/users", usersH.List)
					admin.PUT("/users/:id/nickname", usersH.UpdateNickname)
					admin.POST("/users/:id/require-2fa", usersH.Require2FA)
					// admin.DELETE("/users/:id", usersH.Delete)
					// 管理员重置用户密码
					admin.POST("/users/reset-password", usersH.ResetPassword)

					// Backfill routes
					admin.POST("/timesheets/backfill", handlers.BackfillTimesheets(dbConn))
					admin.GET("/timesheets/backfill/history", handlers.GetBackfillHistory(dbConn))
					admin.DELETE("/timesheets/backfill/:id", handlers.DeleteBackfill(dbConn))

					// 工作排期（甘特图）
					admin.GET("/work-schedules/board", scheduleH.GetBoard)
					admin.POST("/work-schedules/tasks", scheduleH.CreateTask)
					admin.PUT("/work-schedules/tasks/:id", scheduleH.UpdateTask)
					admin.DELETE("/work-schedules/tasks/:id", scheduleH.DeleteTask)
					admin.POST("/work-schedules/calendar/toggle", scheduleH.ToggleCalendarDay)
				}
				tfa.GET("/projects", projH.List)
				tfa.GET("/timesheets/mine", tsH.ListMine)
				tfa.POST("/timesheets", tsH.Create)
				tfa.DELETE("/timesheets/:id", tsH.Delete)
				// 新增：更新本人工时
				tfa.PUT("/timesheets/:id", tsH.Update)
				// 新增：按日期获取本人工时
				tfa.POST("/timesheets/mine/by-date", tsH.ListMineByDate)

				// 用户自助改密
				tfa.POST("/change-password", acctH.ChangePassword)

				// 获取当前用户信息
				tfa.GET("/me", acctH.GetMe)

				// 用户自助禁用 2FA
				tfa.POST("/2fa/disable", acctH.Disable2FA)

				// AI report generation
				tfa.POST("/ai/generate-report", aiH.GenerateReport)

				// Work Plan routes
				tfa.POST("/work-plans/list", handlers.GetWorkPlans(dbConn))
				tfa.POST("/work-plans/mine", handlers.GetMyWorkPlans(dbConn))
				tfa.POST("/work-plans/by-project", handlers.GetWorkPlansByProject(dbConn))
				tfa.POST("/work-plans/create", handlers.CreateWorkPlan(dbConn))
				tfa.POST("/work-plans/update", handlers.UpdateWorkPlan(dbConn))
				tfa.POST("/work-plans/delete", handlers.DeleteWorkPlan(dbConn))
			}
		}
	}

	// API 路由...
	// r.Static("/assets", "./frontend/dist/assets")
	// r.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}
		// 非 API 请求直接交给 Nginx，不返回前端页面
		c.JSON(404, gin.H{"error": "not found"})
	})

	log.Println("server starting on :10081")
	if err := r.Run(":10081"); err != nil {
		panic(err)
	}
}
