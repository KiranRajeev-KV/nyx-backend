package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/KiranRajeev-KV/nyx-backend/cmd"
	db "github.com/KiranRajeev-KV/nyx-backend/internal/db/gen"
	"github.com/KiranRajeev-KV/nyx-backend/internal/logger"
	"github.com/KiranRajeev-KV/nyx-backend/internal/models"
	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/gin-gonic/gin"
)

func FetchAuditLogs(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "AUDIT-LOGS") {
		return
	}
	defer conn.Release()

	q := db.New()

	rows, err := q.FetchAuditLogs(ctx, conn)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[AUDIT-LOGS-ERROR] Failed to fetch audit logs from DB", err)
		return
	}

	auditLogs := make([]models.AuditLogResponse, 0, len(rows))
	for _, row := range rows {
		var actor *models.ActorResponse
		if len(row.Actor) > 0 && string(row.Actor) != "null" {
			var a models.ActorResponse
			if err := json.Unmarshal(row.Actor, &a); err == nil {
				actor = &a
			}
		}

		log := models.AuditLogResponse{
			ID:         row.ID,
			Action:     row.Action,
			TargetType: string(row.TargetType),
			CreatedAt:  row.CreatedAt.Time,
		}

		if row.ActorID.Valid {
			log.ActorID = &row.ActorID.UUID
		}
		if actor != nil {
			log.Actor = actor
		}
		if row.TargetID.Valid {
			log.TargetID = &row.TargetID.UUID
		}

		auditLogs = append(auditLogs, log)
	}

	if auditLogs == nil {
		auditLogs = []models.AuditLogResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Audit logs fetched successfully",
		"data":    auditLogs,
	})
	logger.Log.SuccessCtx(c)
}
