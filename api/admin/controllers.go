package api

import (
	"context"
	"net/http"
	"time"

	"github.com/KiranRajeev-KV/nyx-backend/cmd"
	db "github.com/KiranRajeev-KV/nyx-backend/internal/db/gen"
	"github.com/KiranRajeev-KV/nyx-backend/internal/logger"
	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func FetchAllUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "ADMIN-USERS") {
		return
	}
	defer conn.Release()

	q := db.New()

	users, err := q.FetchAllUsers(ctx, conn)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[ADMIN-USERS-ERROR] Failed to fetch users from DB", err)
		return
	}

	if len(users) == 0 {
		users = []db.FetchAllUsersRow{}
	}

	response := make([]gin.H, len(users))
	for i, user := range users {
		response[i] = gin.H{
			"id":          user.ID,
			"name":        user.Name,
			"email":       user.Email,
			"role":        user.Role,
			"is_banned":   user.IsBanned,
			"trust_score": user.TrustScore,
			"created_at":  user.CreatedAt,
			"updated_at":  user.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Users fetched successfully",
		"data":    response,
	})
	logger.Log.SuccessCtx(c)
}

func BanUser(c *gin.Context) {
	id := c.Param("id")

	targetUUID, exists := pkg.GrabUuid(c, id, "ADMIN-BAN", "userId")
	if !exists {
		return
	}

	// Get the acting admin's ID for audit log
	actorId, ok := pkg.GrabUserId(c, "ADMIN-BAN")
	if !ok {
		return
	}
	actorUUID, exists := pkg.GrabUuid(c, actorId, "ADMIN-BAN", "actorId")
	if !exists {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "ADMIN-BAN") {
		return
	}
	defer conn.Release()

	q := db.New()

	// First check if target user exists
	targetUser, err := q.FetchUserById(ctx, conn, targetUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": "User not found",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[ADMIN-BAN-ERROR] Failed to fetch target user", err)
		return
	}

	// Prevent banning admins
	if targetUser.Role == "ADMIN" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"message": "Cannot ban an admin user",
		})
		logger.Log.WarnCtx(c, "[ADMIN-BAN-WARN] Attempted to ban an admin user")
		return
	}

	// Prevent banning already banned users
	if targetUser.IsBanned {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"message": "User is already banned",
		})
		return
	}

	rowsAffected, err := q.BanUser(ctx, conn, targetUUID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[ADMIN-BAN-ERROR] Failed to ban user", err)
		return
	}

	if rowsAffected == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"message": "Cannot ban this user",
		})
		return
	}

	// Audit log
	err = q.CreateAuditLog(ctx, conn, db.CreateAuditLogParams{
		ActorID:    uuid.NullUUID{UUID: actorUUID, Valid: true},
		Action:     "USER_BANNED",
		TargetType: "USER",
		TargetID:   uuid.NullUUID{UUID: targetUUID, Valid: true},
	})
	if err != nil {
		logger.Log.ErrorCtx(c, "[ADMIN-BAN-ERROR] Failed to create audit log", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User banned successfully",
	})
	logger.Log.SuccessCtx(c)
}

func UnbanUser(c *gin.Context) {
	id := c.Param("id")

	targetUUID, exists := pkg.GrabUuid(c, id, "ADMIN-UNBAN", "userId")
	if !exists {
		return
	}

	actorId, ok := pkg.GrabUserId(c, "ADMIN-UNBAN")
	if !ok {
		return
	}
	actorUUID, exists := pkg.GrabUuid(c, actorId, "ADMIN-UNBAN", "actorId")
	if !exists {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "ADMIN-UNBAN") {
		return
	}
	defer conn.Release()

	q := db.New()

	// Check user exists
	targetUser, err := q.FetchUserById(ctx, conn, targetUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": "User not found",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[ADMIN-UNBAN-ERROR] Failed to fetch target user", err)
		return
	}

	if !targetUser.IsBanned {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"message": "User is not banned",
		})
		return
	}

	err = q.UnbanUser(ctx, conn, targetUUID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[ADMIN-UNBAN-ERROR] Failed to unban user", err)
		return
	}

	// Audit log
	err = q.CreateAuditLog(ctx, conn, db.CreateAuditLogParams{
		ActorID:    uuid.NullUUID{UUID: actorUUID, Valid: true},
		Action:     "USER_UNBANNED",
		TargetType: "USER",
		TargetID:   uuid.NullUUID{UUID: targetUUID, Valid: true},
	})
	if err != nil {
		logger.Log.ErrorCtx(c, "[ADMIN-UNBAN-ERROR] Failed to create audit log", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User unbanned successfully",
	})
	logger.Log.SuccessCtx(c)
}

func PromoteToAdmin(c *gin.Context) {
	id := c.Param("id")

	targetUUID, exists := pkg.GrabUuid(c, id, "ADMIN-PROMOTE", "userId")
	if !exists {
		return
	}

	actorId, ok := pkg.GrabUserId(c, "ADMIN-PROMOTE")
	if !ok {
		return
	}
	actorUUID, exists := pkg.GrabUuid(c, actorId, "ADMIN-PROMOTE", "actorId")
	if !exists {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "ADMIN-PROMOTE") {
		return
	}
	defer conn.Release()

	q := db.New()

	// Check user exists and is a USER
	targetUser, err := q.FetchUserById(ctx, conn, targetUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": "User not found",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[ADMIN-PROMOTE-ERROR] Failed to fetch target user", err)
		return
	}

	if targetUser.Role == "ADMIN" {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"message": "User is already an admin",
		})
		return
	}

	rowsAffected, err := q.PromoteUserToAdmin(ctx, conn, targetUUID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[ADMIN-PROMOTE-ERROR] Failed to promote user", err)
		return
	}

	if rowsAffected == 0 {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"message": "Could not promote user",
		})
		return
	}

	// Audit log
	err = q.CreateAuditLog(ctx, conn, db.CreateAuditLogParams{
		ActorID:    uuid.NullUUID{UUID: actorUUID, Valid: true},
		Action:     "USER_PROMOTED_TO_ADMIN",
		TargetType: "USER",
		TargetID:   uuid.NullUUID{UUID: targetUUID, Valid: true},
	})
	if err != nil {
		logger.Log.ErrorCtx(c, "[ADMIN-PROMOTE-ERROR] Failed to create audit log", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User promoted to admin successfully",
	})
	logger.Log.SuccessCtx(c)
}
