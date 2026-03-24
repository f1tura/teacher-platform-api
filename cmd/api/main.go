// @title Teacher Platform API
// @version 1.0
// @description Backend API for Mini-LMS / Teacher Platform
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"teacher-platform/internal/config"
	"teacher-platform/internal/database"
	"teacher-platform/internal/handlers"
	"teacher-platform/internal/repository"
	"teacher-platform/internal/middleware"
	"github.com/gin-gonic/gin"
	_ "teacher-platform/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	cfg := config.Load()

	dbPool, err := database.NewPool(cfg)
	if err != nil {
		log.Fatal("failed to connect to database: ", err)
	}
	defer dbPool.Close()

	userRepo := repository.NewUserRepository(dbPool)
	courseRepo := repository.NewCourseRepository(dbPool)
	courseHandler := handlers.NewCourseHandler(courseRepo)
	courseStudentRepo := repository.NewCourseStudentRepository(dbPool)
	courseStudentHandler := handlers.NewCourseStudentHandler(courseRepo, userRepo, courseStudentRepo)
	userHandler := handlers.NewUserHandler(userRepo)
	authHandler := handlers.NewAuthHandler(userRepo, cfg)
	lessonRepo := repository.NewLessonRepository(dbPool)
	lessonHandler := handlers.NewLessonHandler(lessonRepo, courseRepo)
	assignmentRepo := repository.NewAssignmentRepository(dbPool)
	assignmentHandler := handlers.NewAssignmentHandler(assignmentRepo, lessonRepo, courseRepo)
	submissionRepo := repository.NewSubmissionRepository(dbPool)
	submissionHandler := handlers.NewSubmissionHandler(submissionRepo, assignmentRepo, lessonRepo, courseRepo, courseStudentRepo)
	progressRepo := repository.NewProgressRepository(dbPool)
	progressHandler := handlers.NewProgressHandler(progressRepo, userRepo, courseStudentRepo)
	attendanceRepo := repository.NewAttendanceRepository(dbPool)
	attendanceHandler := handlers.NewAttendanceHandler(attendanceRepo, lessonRepo, courseRepo, userRepo, courseStudentRepo)

	router := gin.Default()

	router.POST(
	"/courses",
	middleware.AuthMiddleware(cfg.JWTSecret),
	middleware.RequireRoles("teacher", "admin"),
	courseHandler.Create,
	)

	router.GET(
		"/courses",
		middleware.AuthMiddleware(cfg.JWTSecret),
		courseHandler.GetAll,
	)

	router.POST(
	"/courses/:id/lessons",
	middleware.AuthMiddleware(cfg.JWTSecret),
	middleware.RequireRoles("teacher", "admin"),
	lessonHandler.Create,
	)

	router.GET(
		"/courses/:id/lessons",
		middleware.AuthMiddleware(cfg.JWTSecret),
		lessonHandler.GetByCourseID,
	)

	router.GET(
		"/lessons/:id",
		middleware.AuthMiddleware(cfg.JWTSecret),
		lessonHandler.GetByID,
	)

	router.POST(
	"/lessons/:id/attendance",
	middleware.AuthMiddleware(cfg.JWTSecret),
	middleware.RequireRoles("teacher", "admin"),
	attendanceHandler.Create,
	)

	router.GET(
		"/lessons/:id/attendance",
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.RequireRoles("teacher", "admin"),
		attendanceHandler.GetByLessonID,
	)

	router.PATCH(
		"/attendance/:id",
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.RequireRoles("teacher", "admin"),
		attendanceHandler.UpdateByID,
	)
	router.PATCH(
	"/lessons/:id",
	middleware.AuthMiddleware(cfg.JWTSecret),
	middleware.RequireRoles("teacher", "admin"),
	lessonHandler.UpdateByID,
	)

	router.DELETE(
		"/lessons/:id",
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.RequireRoles("teacher", "admin"),
		lessonHandler.DeleteByID,
	)

	router.POST(
	"/lessons/:id/assignments",
	middleware.AuthMiddleware(cfg.JWTSecret),
	middleware.RequireRoles("teacher", "admin"),
	assignmentHandler.Create,
	)

	router.GET(
		"/lessons/:id/assignments",
		middleware.AuthMiddleware(cfg.JWTSecret),
		assignmentHandler.GetByLessonID,
	)

	router.GET(
		"/assignments/:id",
		middleware.AuthMiddleware(cfg.JWTSecret),
		assignmentHandler.GetByID,
	)

	router.PATCH(
	"/submissions/:id/review",
	middleware.AuthMiddleware(cfg.JWTSecret),
	middleware.RequireRoles("teacher", "admin"),
	submissionHandler.Review,
	)

	router.POST(
	"/assignments/:id/submissions",
	middleware.AuthMiddleware(cfg.JWTSecret),
	middleware.RequireRoles("student"),
	submissionHandler.Create,
	)

	router.GET(
		"/assignments/:id/submissions",
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.RequireRoles("teacher", "admin"),
		submissionHandler.GetByAssignmentID,
	)

	router.PATCH(
	"/submissions/:id",
	middleware.AuthMiddleware(cfg.JWTSecret),
	middleware.RequireRoles("student"),
	submissionHandler.UpdateByID,
	)

	router.DELETE(
		"/submissions/:id",
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.RequireRoles("student"),
		submissionHandler.DeleteByID,
	)

	router.PATCH(
	"/assignments/:id",
	middleware.AuthMiddleware(cfg.JWTSecret),
	middleware.RequireRoles("teacher", "admin"),
	assignmentHandler.UpdateByID,
	)

	router.DELETE(
		"/assignments/:id",
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.RequireRoles("teacher", "admin"),
		assignmentHandler.DeleteByID,
	)
	router.GET(
		"/submissions/:id",
		middleware.AuthMiddleware(cfg.JWTSecret),
		submissionHandler.GetByID,
	)
	router.GET(
		"/courses/:id",
		middleware.AuthMiddleware(cfg.JWTSecret),
		courseHandler.GetByID,
	)

	router.PATCH(
	"/courses/:id",
	middleware.AuthMiddleware(cfg.JWTSecret),
	middleware.RequireRoles("teacher", "admin"),
	courseHandler.UpdateByID,
	)

	router.DELETE(
		"/courses/:id",
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.RequireRoles("teacher", "admin"),
		courseHandler.DeleteByID,
	)

	router.POST(
	"/courses/:id/students",
	middleware.AuthMiddleware(cfg.JWTSecret),
	middleware.RequireRoles("teacher", "admin"),
	courseStudentHandler.AddStudent,
	)

	router.GET(
		"/courses/:id/students",
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.RequireRoles("teacher", "admin"),
		courseStudentHandler.GetStudents,
	)

	router.DELETE(
		"/courses/:id/students/:studentId",
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.RequireRoles("teacher", "admin"),
		courseStudentHandler.RemoveStudent,
	)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	router.GET(
	"/students/:id/progress",
	middleware.AuthMiddleware(cfg.JWTSecret),
	progressHandler.GetStudentProgress,
	)
	
	router.GET("/health/db", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := dbPool.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "db down",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "db ok",
		})
	})
	router.GET(
	"/teacher/ping",
	middleware.AuthMiddleware(cfg.JWTSecret),
	middleware.RequireRoles("teacher", "admin"),
	func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "teacher access granted",
		})
	},
	)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.POST("/users", userHandler.Create)
	router.GET("/users", userHandler.GetAll)
	router.GET("/users/:id", userHandler.GetByID)
	router.DELETE("/users/:id", userHandler.DeleteByID)
	router.PATCH("/users/:id", userHandler.UpdateByID)

	router.POST("/auth/register", authHandler.Register)
	router.POST("/auth/login", authHandler.Login)
	router.GET("/auth/me", middleware.AuthMiddleware(cfg.JWTSecret), authHandler.Me)
	log.Println("server started on :" + cfg.AppPort)

	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}