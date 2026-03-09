package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/KiranRajeev-KV/nyx-backend/cmd"
	db "github.com/KiranRajeev-KV/nyx-backend/internal/db/gen"
	"github.com/KiranRajeev-KV/nyx-backend/internal/logger"
	"github.com/KiranRajeev-KV/nyx-backend/internal/models"
	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/KiranRajeev-KV/nyx-backend/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func FetchItems(c *gin.Context) {
	// params: type="LOST" | "FOUND" (optional)
	// status must be OPEN or PENDING_CLAIM (default both)

	typeParam, exists := c.GetQuery("type")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "ITEMS") {
		return
	}
	defer conn.Release()

	q := db.New()

	var items []db.FetchItemsByTypeRow

	if !exists || typeParam == "" {
		// Fetch all items
		allItems, err := q.FetchAllItems(ctx, conn)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Oops! Something happened. Please try again later",
			})
			logger.Log.ErrorCtx(c, "[ITEMS-ERROR] Failed to fetch all items from DB", err)
			return
		}

		// Convert FetchAllItemsRow to FetchItemsByTypeRow for consistent response
		items = make([]db.FetchItemsByTypeRow, len(allItems))
		for i, item := range allItems {
			items[i] = db.FetchItemsByTypeRow(item)
		}
	} else {
		// Validate type parameter
		if typeParam != "LOST" && typeParam != "FOUND" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Invalid 'type' query parameter. Must be 'LOST' or 'FOUND'",
			})
			logger.Log.WarnCtx(c, "[ITEMS-WARN] Invalid 'type' query parameter")
			return
		}

		// Fetch items by type
		items, err = q.FetchItemsByType(ctx, conn, db.ItemType(typeParam))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Oops! Something happened. Please try again later",
			})
			logger.Log.ErrorCtx(c, "[ITEMS-ERROR] Failed to fetch items from DB", err)
			return
		}
	}

	if len(items) == 0 {
		items = []db.FetchItemsByTypeRow{}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Items fetched successfully",
		"data":    items,
	})
	logger.Log.SuccessCtx(c)
}

func CreateItem(c *gin.Context) {
	req, ok := pkg.ValidateRequest[models.CreateItemRequest](c)
	if !ok {
		return
	}

	userId, ok := pkg.GrabUserId(c, "ITEMS")
	if !ok {
		return
	}

	userUUID, exists := pkg.GrabUuid(c, userId, "ITEMS", "userId")
	if !exists {
		return
	}

	var hubUUID uuid.NullUUID
	if req.Type == "FOUND" {
		if req.HubId == nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Hub ID is required for FOUND items",
			})
			return
		}
		parsedHubUUID, exists := pkg.GrabUuid(c, *req.HubId, "ITEMS", "hubId")
		if !exists {
			return
		}
		hubUUID = uuid.NullUUID{UUID: parsedHubUUID, Valid: true}
	} else {
		// For LOST items, HubID must be NULL
		hubUUID = uuid.NullUUID{Valid: false}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "ITEMS") {
		return
	}
	defer conn.Release()

	q := db.New()

	var timeAt time.Time
	if req.TimeAt != "" {
		var err error
		timeAt, err = time.Parse(time.RFC3339, req.TimeAt)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Invalid time format for TimeAt",
			})
			logger.Log.WarnCtx(c, "[ITEMS-WARN] Invalid time format")
			return
		}
	}

	newItem, err := q.CreateItem(ctx, conn, db.CreateItemParams{
		UserID:              userUUID,
		IsAnonymous:         req.IsAnonymous,
		HubID:               hubUUID,
		Name:                req.Name,
		Description:         pgtype.Text{String: req.Description, Valid: true},
		Type:                db.ItemType(req.Type),
		LocationDescription: pgtype.Text{String: req.Location, Valid: true},
		TimeAt:              pgtype.Timestamptz{Time: timeAt, Valid: !timeAt.IsZero()},
		Latitude:            pgtype.Text{String: req.Latitude, Valid: true},
		Longitude:           pgtype.Text{String: req.Longitude, Valid: true},
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[ITEMS-ERROR] Failed to create item in DB", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Item created successfully",
		"data":    newItem,
	})
	logger.Log.SuccessCtx(c)
}

func FetchItemById(c *gin.Context) {
	id := c.Param("id")

	itemUUID, exists := pkg.GrabUuid(c, id, "ITEMS", "itemId")
	if !exists {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "ITEMS") {
		return
	}
	defer conn.Release()

	q := db.New()

	item, err := q.FetchItemByID(ctx, conn, itemUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": "Item not found",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[ITEMS-ERROR] Failed to fetch item by ID from DB", err)
		return
	}

	var userObj any
	if item.User != nil {
		if err := json.Unmarshal(item.User, &userObj); err != nil {
			logger.Log.ErrorCtx(c, "[ITEMS-ERROR] Failed to unmarshal user JSON", err)
		}
	}

	var hubObj any
	if item.Hub != nil {
		if err := json.Unmarshal(item.Hub, &hubObj); err != nil {
			logger.Log.ErrorCtx(c, "[ITEMS-ERROR] Failed to unmarshal hub JSON", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item fetched successfully",
		"data": gin.H{
			"id":                   item.ID,
			"is_anonymous":         item.IsAnonymous,
			"name":                 item.Name,
			"image_url_redacted":   item.ImageUrlRedacted,
			"description":          item.Description,
			"status":               item.Status,
			"type":                 item.Type,
			"location_description": item.LocationDescription,
			"time_at":              item.TimeAt,
			"latitude":             item.Latitude,
			"longitude":            item.Longitude,
			"created_at":           item.CreatedAt,
			"updated_at":           item.UpdatedAt,
			"user":                 userObj,
			"hub":                  hubObj,
		},
	})
	logger.Log.SuccessCtx(c)
}

func FetchAllItemsByUserId(c *gin.Context) {
	userId, ok := pkg.GrabUserId(c, "ITEMS")
	if !ok {
		return
	}

	userUUID, exists := pkg.GrabUuid(c, userId, "ITEMS", "userId")
	if !exists {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "ITEMS") {
		return
	}
	defer conn.Release()

	q := db.New()

	items, err := q.FetchAllItemsByUserId(ctx, conn, userUUID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[ITEMS-ERROR] Failed to fetch items by user ID from DB", err)
		return
	}

	response := make([]gin.H, len(items))
	for i, item := range items {
		var hubObj any
		if item.Hub != nil {
			if err := json.Unmarshal(item.Hub, &hubObj); err != nil {
				logger.Log.ErrorCtx(c, "[ITEMS-ERROR] Failed to unmarshal hub JSON", err)
			}
		}

		response[i] = gin.H{
			"id":                   item.ID,
			"is_anonymous":         item.IsAnonymous,
			"name":                 item.Name,
			"image_url_redacted":   item.ImageUrlRedacted,
			"description":          item.Description,
			"status":               item.Status,
			"type":                 item.Type,
			"location_description": item.LocationDescription,
			"time_at":              item.TimeAt,
			"latitude":             item.Latitude,
			"longitude":            item.Longitude,
			"created_at":           item.CreatedAt,
			"updated_at":           item.UpdatedAt,
			"user":                 item.User,
			"hub":                  hubObj,
		}

		// Build full image URL if image_url_original is set
		if item.ImageUrlOriginal.Valid && item.ImageUrlOriginal.String != "" {
			response[i]["image_url_original"] = storage.S3.GetPublicURL(item.ImageUrlOriginal.String)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Items fetched successfully",
		"data":    response,
	})
	logger.Log.SuccessCtx(c)
}

func UpdateItemById(c *gin.Context) {
	id := c.Param("id")

	itemUUID, exists := pkg.GrabUuid(c, id, "ITEMS", "itemId")
	if !exists {
		return
	}

	userId, ok := pkg.GrabUserId(c, "ITEMS")
	if !ok {
		return
	}

	userUUID, exists := pkg.GrabUuid(c, userId, "ITEMS", "userId")
	if !exists {
		return
	}

	req, ok := pkg.ValidateRequest[models.UpdateItemRequest](c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "ITEMS") {
		return
	}
	defer conn.Release()

	q := db.New()

	params := db.UpdateItemByIdParams{
		ID:     itemUUID,
		UserID: userUUID,
	}

	if req.Name != nil {
		params.Name = pgtype.Text{String: *req.Name, Valid: true}
	} else {
		params.Name = pgtype.Text{Valid: false}
	}

	if req.Description != nil {
		params.Description = pgtype.Text{String: *req.Description, Valid: true}
	} else {
		params.Description = pgtype.Text{Valid: false}
	}

	if req.Location != nil {
		params.LocationDescription = pgtype.Text{String: *req.Location, Valid: true}
	} else {
		params.LocationDescription = pgtype.Text{Valid: false}
	}

	if req.Latitude != nil {
		params.Latitude = pgtype.Text{String: *req.Latitude, Valid: true}
	} else {
		params.Latitude = pgtype.Text{Valid: false}
	}

	if req.Longitude != nil {
		params.Longitude = pgtype.Text{String: *req.Longitude, Valid: true}
	} else {
		params.Longitude = pgtype.Text{Valid: false}
	}

	if req.HubId != nil {
		hubUUID, err := uuid.Parse(*req.HubId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Invalid Hub ID",
			})
			return
		}
		params.HubID = uuid.NullUUID{UUID: hubUUID, Valid: true}
	} else {
		params.HubID = uuid.NullUUID{Valid: false}
	}

	if req.TimeAt != nil {
		t, err := time.Parse(time.RFC3339, *req.TimeAt)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Invalid time format for TimeAt",
			})
			return
		}
		params.TimeAt = pgtype.Timestamptz{Time: t, Valid: true}
	} else {
		params.TimeAt = pgtype.Timestamptz{Valid: false}
	}

	updatedItem, err := q.UpdateItemById(ctx, conn, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": "Item not found or you are not authorized to update it",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[ITEMS-ERROR] Failed to update item in DB", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item updated successfully",
		"data":    updatedItem,
	})
	logger.Log.SuccessCtx(c)
}

func DeleteItemById(c *gin.Context) {
	id := c.Param("id")

	itemUUID, exists := pkg.GrabUuid(c, id, "ITEMS", "itemId")
	if !exists {
		return
	}

	userId, ok := pkg.GrabUserId(c, "ITEMS")
	if !ok {
		return
	}

	userUUID, exists := pkg.GrabUuid(c, userId, "ITEMS", "userId")
	if !exists {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "ITEMS") {
		return
	}
	defer conn.Release()

	q := db.New()

	_, err = q.SoftDeleteItemById(ctx, conn, db.SoftDeleteItemByIdParams{
		ID:     itemUUID,
		UserID: userUUID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": "Item not found or you are not authorized to delete it",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[ITEMS-ERROR] Failed to delete item in DB", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item deleted successfully",
	})
	logger.Log.SuccessCtx(c)
}

func UpdateItemStatus(c *gin.Context) {
	id := c.Param("id")

	itemUUID, exists := pkg.GrabUuid(c, id, "ITEMS", "itemId")
	if !exists {
		return
	}

	req, ok := pkg.ValidateRequest[models.UpdateItemStatusRequest](c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "ITEMS") {
		return
	}
	defer conn.Release()

	q := db.New()

	updatedItem, err := q.UpdateItemStatusById(ctx, conn, db.UpdateItemStatusByIdParams{
		ID:     itemUUID,
		Status: db.ItemStatus(req.Status),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": "Item not found",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[ITEMS-ERROR] Failed to update item status in DB", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item status updated successfully",
		"data":    updatedItem,
	})
	logger.Log.SuccessCtx(c)
}

func UploadItemImage(c *gin.Context) {
	id := c.Param("id")

	itemUUID, exists := pkg.GrabUuid(c, id, "ITEMS", "itemId")
	if !exists {
		return
	}

	userId, ok := pkg.GrabUserId(c, "ITEMS")
	if !ok {
		return
	}

	userUUID, exists := pkg.GrabUuid(c, userId, "ITEMS", "userId")
	if !exists {
		return
	}

	req, okValidate := pkg.ValidateRequest[models.UploadItemImageRequest](c)
	if !okValidate {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "ITEMS") {
		return
	}
	defer conn.Release()

	q := db.New()

	// Verify the user owns the item
	item, err := q.FetchItemByID(ctx, conn, itemUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": "Item not found",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[ITEMS-ERROR] Failed to fetch item by ID", err)
		return
	}

	// Unmarshal user object to get the ID and verify ownership
	// The User JSON object contains the ID
	var userObj map[string]interface{}
	if err := json.Unmarshal(item.User, &userObj); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[ITEMS-ERROR] Failed to unmarshal user JSON", err)
		return
	}

	// Verify ownership
	if userStr, ok := userObj["id"].(string); !ok || userStr != userUUID.String() {
		// Only admins or the owner can upload images for this item
		isAdmin := false
		if role, ok := c.Get("role"); ok && role == "ADMIN" {
			isAdmin = true
		}

		if !isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "You are not authorized to upload images for this item",
			})
			logger.Log.WarnCtx(c, "[ITEMS-WARN] Unauthorized image upload attempt")
			return
		}
	}

	// Generate a unique object key
	// Example: items/uuid/image_original_1234.jpg
	fileExt := ".jpg"
	switch req.ContentType {
	case "image/png":
		fileExt = ".png"
	case "image/webp":
		fileExt = ".webp"
	}

	imageUUID := uuid.New().String()
	objectKey := fmt.Sprintf("items/%s/image_original_%s%s", itemUUID.String(), imageUUID, fileExt)

	// Assume we want the presigned URL to be valid for 15 minutes
	presignedUrl, err := storage.S3.GeneratePresignedPutURL(ctx, objectKey, 15*time.Minute)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to generate upload URL",
		})
		logger.Log.ErrorCtx(c, "[ITEMS-ERROR] Failed to generate presigned URL", err)
		return
	}

	// We can update the Item record so we know what image path to expect
	// Or we can wait for a webhook to confirm the upload.
	// For now, we just update it immediately.
	_, err = q.UpdateItemImageOriginal(ctx, conn, db.UpdateItemImageOriginalParams{
		ID:               itemUUID,
		UserID:           userUUID,
		ImageUrlOriginal: pgtype.Text{String: objectKey, Valid: true},
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[ITEMS-ERROR] Failed to update item with image key", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Upload URL generated successfully",
		"data": models.UploadItemImageResponse{
			PresignedUrl: presignedUrl,
			ObjectKey:    objectKey,
		},
	})
	logger.Log.SuccessCtx(c)
}
