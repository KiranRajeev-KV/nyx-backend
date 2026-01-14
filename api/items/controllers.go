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

	q := db.New(conn)

	var items []db.FetchItemsByTypeRow

	if !exists || typeParam == "" {
		// Fetch all items
		allItems, err := q.FetchAllItems(ctx)
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
		items, err = q.FetchItemsByType(ctx, db.ItemType(typeParam))
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

func CreateItem(c *gin.Context) {}

func FetchItemByID(c *gin.Context) {}

func UpdateItemByID(c *gin.Context) {}

func DeleteItemByID(c *gin.Context) {}
