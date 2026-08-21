// Package bootstrap wires together infra, third-party packages and feature
// modules, then hands back a ready-to-run gin.Engine.
package bootstrap

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"mission-note/internal/app/auth"
	"mission-note/internal/app/flashcard"
	"mission-note/internal/app/shadowing"
	"mission-note/internal/config"
	"mission-note/internal/core/middleware"
	"mission-note/internal/core/ratelimiting"
	"mission-note/internal/core/validation"
	"mission-note/internal/infra/db"
	"mission-note/internal/pkg/openai"
)

// corsMiddleware allows any origin to call the API — permissive on purpose so
// local test pages/frontends (served from a different port or file://) can
// call the API during development.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// App holds everything that needs to be closed on shutdown.
type App struct {
	Router *gin.Engine
	Port   string
	close  func()
}

// Close releases resources acquired during Init (e.g. the DB pool).
func (a *App) Close() {
	if a.close != nil {
		a.close()
	}
}

// Init loads config, connects infra, and registers every feature module.
func Init(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	if err := validation.RegisterCustomValidators(); err != nil {
		return nil, err
	}

	pool, err := db.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	aiClient := openai.New(cfg.OpenAIAPIKey, cfg.OpenAIBaseURL, cfg.OpenAIModel, cfg.OpenAISTTModel, cfg.AIMockImages)

	// =========================================================================
	// 1. GLOBAL MIDDLEWARES (ส่วนการตั้งค่าความปลอดภัยระดับระบบ)
	// =========================================================================
	router := gin.Default()

	// อนุญาตให้ Frontend ยิง API ข้าม Origin ได้ (CORS)
	// และจำกัดจำนวนการยิง Request เพื่อป้องกัน Spam (Rate Limiting)
	router.Use(corsMiddleware())
	router.Use(ratelimiting.PerIP(10, 20))

	// เสิร์ฟไฟล์ไอคอน tag/badge แบบ static จากเครื่อง (ยังไม่มีไฟล์จริงตอนนี้
	// แค่เปิด path ไว้ก่อน — วางไฟล์ลงโฟลเดอร์นี้แล้วเข้าถึงได้ที่ /assets/icons/<ชื่อไฟล์>)
	router.Static("/assets/icons", "./assets/icons")

	// Base Path หลักของ API ทั้งหมด
	api := router.Group("/api/v1")

	// =========================================================================
	// 2. PUBLIC ROUTES (กลุ่ม API สาธารณะ - ใครๆ ก็เข้าถึงได้ ไม่ต้องใช้ Token)
	// =========================================================================
	// โมดูลระบบยืนยันตัวตน (เช่น POST /register, POST /login)
	auth.RegisterModule(api, pool, cfg.JWTSecret)

	// =========================================================================
	// 3. PROTECTED ROUTES (กลุ่ม API ลับ - บังคับตรวจ JWT Token ทุกเส้นผ่าน RequireAuth)
	// =========================================================================
	protected := api.Group("")
	protected.Use(middleware.RequireAuth(cfg.JWTSecret))

	// โมดูลการเรียนรู้ (ต้องยืนยันตัวตนก่อนเพื่อดึง/บันทึกข้อมูลเฉพาะบุคคล)
	flashcard.RegisterModule(protected, pool, aiClient)
	shadowing.RegisterModule(protected, pool, aiClient)

	app := &App{
		Router: router,
		Port:   cfg.AppPort,
		close:  func() { pool.Close() },
	}

	return app, nil
}
