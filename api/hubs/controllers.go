package api

import (
	"context"
	"net/http"
	"time"

	"github.com/KiranRajeev-KV/nyx-backend/cmd"
	db "github.com/KiranRajeev-KV/nyx-backend/internal/db/gen"
	"github.com/KiranRajeev-KV/nyx-backend/internal/logger"
	"github.com/KiranRajeev-KV/nyx-backend/internal/models"
	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func CreateHub(c *gin.Context) {
	req, ok := pkg.ValidateRequest[models.CreateHubRequest](c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := cmd.DBPool.Begin(ctx)
	if pkg.HandleDbTxnErr(c, err, "HUB-CREATE") {
		return
	}
	defer pkg.RollbackTx(c, tx, ctx, "HUB-CREATE")

	q := db.New()

	hub, err := q.CreateHub(ctx, tx, db.CreateHubParams{
		Name:      req.Name,
		Address:   pgtype.Text{String: req.Address, Valid: true},
		Contact:   pgtype.Text{String: req.Contact, Valid: true},
		Longitude: pgtype.Text{String: req.Longitude, Valid: req.Longitude != ""},
		Latitude:  pgtype.Text{String: req.Latitude, Valid: req.Latitude != ""},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[HUB-CREATE-ERROR]: Failed to create hub in DB", err)
		return
	}

	err = tx.Commit(ctx)
	if pkg.HandleDbTxnCommitErr(c, err, "HUB-CREATE") {
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Hub created successfully",
		"data":    hub,
	})
	logger.Log.SuccessCtx(c)
}

func FetchHubs(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "HUBS") {
		return
	}
	defer conn.Release()

	q := db.New()

	hubs, err := q.FetchAllHubs(ctx, conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[HUBS-ERROR]: Failed to fetch hubs from DB", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Hubs fetched successfully",
		"data":    hubs,
	})
	logger.Log.SuccessCtx(c)
}

func FetchHubById(c *gin.Context) {
	id := c.Param("id")
	hubId, ok := pkg.GrabUuid(c, id, "HUB-BY-ID", "id")
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "HUB-BY-ID") {
		return
	}
	defer conn.Release()

	q := db.New()

	hub, err := q.FetchHubByID(ctx, conn, hubId)
	if err != nil {
		if err.Error() == "no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Hub not found",
			})
			logger.Log.WarnCtx(c, "[HUB-BY-ID-WARN]: Hub not found")
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[HUB-BY-ID-ERROR]: Failed to fetch hub from DB", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Hub fetched successfully",
		"data":    hub,
	})
	logger.Log.SuccessCtx(c)
}

func UpdateHub(c *gin.Context) {
	id := c.Param("id")
	hubId, ok := pkg.GrabUuid(c, id, "HUB-UPDATE", "id")
	if !ok {
		return
	}

	req, ok := pkg.ValidateRequest[models.UpdateHubRequest](c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := cmd.DBPool.Begin(ctx)
	if pkg.HandleDbTxnErr(c, err, "HUB-UPDATE") {
		return
	}
	defer pkg.RollbackTx(c, tx, ctx, "HUB-UPDATE")

	q := db.New()

	// Build update parameters
	params := db.UpdateHubParams{
		ID: hubId,
	}

	if req.Name != nil {
		params.Name = pgtype.Text{String: *req.Name, Valid: true}
	}
	if req.Address != nil {
		params.Address = pgtype.Text{String: *req.Address, Valid: true}
	}
	if req.Contact != nil {
		params.Contact = pgtype.Text{String: *req.Contact, Valid: true}
	}
	if req.Longitude != nil {
		params.Longitude = pgtype.Text{String: *req.Longitude, Valid: *req.Longitude != ""}
	}
	if req.Latitude != nil {
		params.Latitude = pgtype.Text{String: *req.Latitude, Valid: *req.Latitude != ""}
	}

	hub, err := q.UpdateHub(ctx, tx, params)
	if err != nil {
		if err.Error() == "no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Hub not found",
			})
			logger.Log.WarnCtx(c, "[HUB-UPDATE-WARN]: Hub not found")
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[HUB-UPDATE-ERROR]: Failed to update hub in DB", err)
		return
	}

	err = tx.Commit(ctx)
	if pkg.HandleDbTxnCommitErr(c, err, "HUB-UPDATE") {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Hub updated successfully",
		"data":    hub,
	})
	logger.Log.SuccessCtx(c)
}

func DeleteHub(c *gin.Context) {
	id := c.Param("id")
	hubId, ok := pkg.GrabUuid(c, id, "HUB-DELETE", "id")
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := cmd.DBPool.Begin(ctx)
	if pkg.HandleDbTxnErr(c, err, "HUB-DELETE") {
		return
	}
	defer pkg.RollbackTx(c, tx, ctx, "HUB-DELETE")

	q := db.New()

	// Check if hub has linked items
	itemCount, err := q.CheckHubLinkedItems(ctx, tx, uuid.NullUUID{UUID: hubId, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[HUB-DELETE-ERROR]: Failed to check hub linked items", err)
		return
	}

	if itemCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"message": "Cannot delete hub: it has associated items",
		})
		logger.Log.WarnCtx(c, "[HUB-DELETE-WARN]: Attempted to delete hub with associated items")
		return
	}

	// Delete the hub
	err = q.DeleteHub(ctx, tx, hubId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[HUB-DELETE-ERROR]: Failed to delete hub from DB", err)
		return
	}

	err = tx.Commit(ctx)
	if pkg.HandleDbTxnCommitErr(c, err, "HUB-DELETE") {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Hub deleted successfully",
	})
	logger.Log.SuccessCtx(c)
}
