package http

import (
	"deadline-bot/pkg/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Services  map[string]interface{} `json:"services"`
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		api.GET("/health", h.Health)
		api.GET("/health/live", h.Liveness)
		api.GET("/health/ready", h.Readiness)
	}
}

func (h *HealthHandler) Health(c *gin.Context) {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Services:  make(map[string]interface{}),
	}

	dbStatus := h.checkDatabase()
	response.Services["database"] = dbStatus

	if dbStatus["status"] != "ok" {
		response.Status = "degraded"
		utils.RespondWithError(c, http.StatusServiceUnavailable, "Service degraded")
		return
	}

	utils.RespondWithSuccess(c, response)
}

func (h *HealthHandler) Liveness(c *gin.Context) {
	utils.RespondWithSuccess(c, map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
	})
}

func (h *HealthHandler) Readiness(c *gin.Context) {
	dbStatus := h.checkDatabase()

	if dbStatus["status"] != "ok" {
		utils.RespondWithError(c, http.StatusServiceUnavailable, "Service not ready")
		return
	}

	utils.RespondWithSuccess(c, map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now(),
		"services": map[string]interface{}{
			"database": dbStatus,
		},
	})
}

func (h *HealthHandler) checkDatabase() map[string]interface{} {
	start := time.Now()

	sqlDB, err := h.db.DB()
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"error":   err.Error(),
			"latency": time.Since(start).Milliseconds(),
		}
	}

	if err := sqlDB.Ping(); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"error":   err.Error(),
			"latency": time.Since(start).Milliseconds(),
		}
	}

	return map[string]interface{}{
		"status":  "ok",
		"latency": time.Since(start).Milliseconds(),
	}
}
