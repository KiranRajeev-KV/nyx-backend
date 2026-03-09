package api

import (
	"context"
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

func CreateClaim(c *gin.Context) {
	req, ok := pkg.ValidateRequest[models.CreateClaimRequest](c)
	if !ok {
		return
	}

	userId, ok := pkg.GrabUserId(c, "CLAIMS")
	if !ok {
		return
	}

	userUUID, exists := pkg.GrabUuid(c, userId, "CLAIMS", "userId")
	if !exists {
		return
	}

	itemUUID, exists := pkg.GrabUuid(c, req.ItemID, "CLAIMS", "itemId")
	if !exists {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "CLAIMS") {
		return
	}
	defer conn.Release()

	q := db.New()

	// Check if item exists and is FOUND with OPEN or PENDING_CLAIM status
	item, err := q.GetItemByID(ctx, conn, itemUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": "Item not found",
			})
			logger.Log.WarnCtx(c, "[CLAIMS-WARN] Attempt to claim non-existent item")
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[CLAIMS-ERROR] Failed to fetch item for claim validation", err)
		return
	}

	// Business rule: Can only claim FOUND items
	if item.Type != db.ItemTypeFOUND {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Only FOUND items can be claimed",
		})
		logger.Log.WarnCtx(c, "[CLAIMS-WARN] Attempt to claim non-FOUND item")
		return
	}

	// Business rule: Can only claim items with OPEN or PENDING_CLAIM status
	if item.Status != db.ItemStatusOPEN && item.Status != db.ItemStatusPENDINGCLAIM {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "This item is not available for claiming",
		})
		logger.Log.WarnCtx(c, "[CLAIMS-WARN] Attempt to claim item with invalid status")
		return
	}

	// Business rule: Cannot claim own items
	if item.UserID == userUUID {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "You cannot claim your own item",
		})
		logger.Log.WarnCtx(c, "[CLAIMS-WARN] Attempt to claim own item")
		return
	}

	// Business rule: Check if user has already claimed this item
	existingClaim, err := q.CheckExistingClaim(ctx, conn, db.CheckExistingClaimParams{
		ItemID:     itemUUID,
		ClaimantID: userUUID,
	})
	if err == nil {
		// If claim exists
		if existingClaim.Status == db.ClaimStatusREJECTED {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "You have already claimed this item and it was rejected. You cannot claim it again.",
			})
			logger.Log.WarnCtx(c, "[CLAIMS-WARN] Attempt to re-claim rejected item")
			return
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "You have already claimed this item",
			})
			logger.Log.WarnCtx(c, "[CLAIMS-WARN] Attempt to duplicate claim on same item")
			return
		}
	} else if err != pgx.ErrNoRows {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[CLAIMS-ERROR] Failed to check existing claim", err)
		return
	}

	// All validations passed, create claim in transaction
	tx, err := cmd.DBPool.Begin(ctx)
	if pkg.HandleDbTxnErr(c, err, "CLAIMS") {
		return
	}
	defer pkg.RollbackTx(c, tx, ctx, "CLAIMS")

	// Create claim
	var proofImageUrl pgtype.Text
	if req.ProofImageUrl != nil {
		proofImageUrl = pgtype.Text{String: *req.ProofImageUrl, Valid: true}
	}

	newClaim, err := q.CreateClaim(ctx, tx, db.CreateClaimParams{
		ItemID:        itemUUID,
		ClaimantID:    userUUID,
		ProofText:     pgtype.Text{String: req.ProofText, Valid: true},
		ProofImageUrl: proofImageUrl,
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[CLAIMS-ERROR] Failed to create claim", err)
		return
	}

	// Update item status to PENDING_CLAIM
	_, err = q.UpdateItemStatus(ctx, tx, db.UpdateItemStatusParams{
		ID:     itemUUID,
		Status: db.ItemStatusPENDINGCLAIM,
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[CLAIMS-ERROR] Failed to update item status", err)
		return
	}

	err = tx.Commit(ctx)
	if pkg.HandleDbTxnCommitErr(c, err, "CLAIMS") {
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Claim submitted successfully",
		"data":    newClaim,
	})
	logger.Log.SuccessCtx(c)
}

func FetchUserClaims(c *gin.Context) {
	userId, ok := pkg.GrabUserId(c, "CLAIMS")
	if !ok {
		return
	}

	userUUID, exists := pkg.GrabUuid(c, userId, "CLAIMS", "userId")
	if !exists {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "CLAIMS") {
		return
	}
	defer conn.Release()

	q := db.New()

	claims, err := q.FetchClaimsByUser(ctx, conn, userUUID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[CLAIMS-ERROR] Failed to fetch user claims", err)
		return
	}

	if len(claims) == 0 {
		claims = []db.FetchClaimsByUserRow{}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User claims fetched successfully",
		"data":    claims,
	})
	logger.Log.SuccessCtx(c)
}

func FetchClaimsByItem(c *gin.Context) {
	itemId := c.Param("id")

	itemUUID, exists := pkg.GrabUuid(c, itemId, "CLAIMS", "itemId")
	if !exists {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "CLAIMS") {
		return
	}
	defer conn.Release()

	q := db.New()

	claims, err := q.FetchClaimsByItem(ctx, conn, itemUUID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[CLAIMS-ERROR] Failed to fetch claims by item", err)
		return
	}

	if len(claims) == 0 {
		claims = []db.FetchClaimsByItemRow{}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Claims for item fetched successfully",
		"data":    claims,
	})
	logger.Log.SuccessCtx(c)
}

func FetchAllClaims(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "CLAIMS") {
		return
	}
	defer conn.Release()

	q := db.New()

	claims, err := q.FetchAllClaims(ctx, conn)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[CLAIMS-ERROR] Failed to fetch all claims", err)
		return
	}

	if len(claims) == 0 {
		claims = []db.FetchAllClaimsRow{}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All claims fetched successfully",
		"data":    claims,
	})
	logger.Log.SuccessCtx(c)
}

func ProcessClaim(c *gin.Context) {
	claimId := c.Param("id")

	claimUUID, exists := pkg.GrabUuid(c, claimId, "CLAIMS", "claimId")
	if !exists {
		return
	}

	req, ok := pkg.ValidateRequest[models.ProcessClaimRequest](c)
	if !ok {
		return
	}

	adminUserId, ok := pkg.GrabUserId(c, "CLAIMS")
	if !ok {
		return
	}

	adminUserUUID, exists := pkg.GrabUuid(c, adminUserId, "CLAIMS", "adminUserId")
	if !exists {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "CLAIMS") {
		return
	}
	defer conn.Release()

	q := db.New()

	// Get claim details first
	claim, err := q.FetchClaimByID(ctx, conn, claimUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": "Claim not found",
			})
			logger.Log.WarnCtx(c, "[CLAIMS-WARN] Attempt to process non-existent claim")
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[CLAIMS-ERROR] Failed to fetch claim for processing", err)
		return
	}

	// Only allow processing of PENDING claims
	if claim.Status != db.ClaimStatusPENDING {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "This claim has already been processed",
		})
		logger.Log.WarnCtx(c, "[CLAIMS-WARN] Attempt to process already processed claim")
		return
	}

	// Process claim in transaction
	tx, err := cmd.DBPool.Begin(ctx)
	if pkg.HandleDbTxnErr(c, err, "CLAIMS") {
		return
	}
	defer pkg.RollbackTx(c, tx, ctx, "CLAIMS")

	// Update claim
	var claimStatus db.ClaimStatus
	if req.Status == "APPROVED" {
		claimStatus = db.ClaimStatusAPPROVED
	} else if req.Status == "REJECTED" {
		claimStatus = db.ClaimStatusREJECTED
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid status. Must be APPROVED or REJECTED",
		})
		return
	}

	updatedClaim, err := q.ProcessClaim(ctx, tx, db.ProcessClaimParams{
		ID:          claimUUID,
		Status:      claimStatus,
		AdminNotes:  pgtype.Text{String: req.AdminNotes, Valid: true},
		ProcessedBy: uuid.NullUUID{UUID: adminUserUUID, Valid: true},
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[CLAIMS-ERROR] Failed to process claim", err)
		return
	}

	// Handle item status based on claim processing result
	if claimStatus == db.ClaimStatusAPPROVED {
		// If approved, set item to RESOLVED
		_, err = q.UpdateItemStatus(ctx, tx, db.UpdateItemStatusParams{
			ID:     claim.ItemID,
			Status: db.ItemStatusRESOLVED,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Oops! Something happened. Please try again later",
			})
			logger.Log.ErrorCtx(c, "[CLAIMS-ERROR] Failed to update item status to RESOLVED", err)
			return
		}
	} else if claimStatus == db.ClaimStatusREJECTED {
		// If rejected, check if there are other pending claims
		pendingCount, err := q.GetPendingClaimsCount(ctx, tx, claim.ItemID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Oops! Something happened. Please try again later",
			})
			logger.Log.ErrorCtx(c, "[CLAIMS-ERROR] Failed to get pending claims count", err)
			return
		}

		// If no other pending claims, set item back to OPEN
		var newStatus db.ItemStatus
		if pendingCount > 0 {
			newStatus = db.ItemStatusPENDINGCLAIM
		} else {
			newStatus = db.ItemStatusOPEN
		}

		_, err = q.UpdateItemStatus(ctx, tx, db.UpdateItemStatusParams{
			ID:     claim.ItemID,
			Status: newStatus,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Oops! Something happened. Please try again later",
			})
			logger.Log.ErrorCtx(c, "[CLAIMS-ERROR] Failed to update item status after rejection", err)
			return
		}
	}

	err = tx.Commit(ctx)
	if pkg.HandleDbTxnCommitErr(c, err, "CLAIMS") {
		return
	}

	// Prepare response data
	response := gin.H{
		"id":          updatedClaim.ID,
		"status":      updatedClaim.Status,
		"admin_notes": updatedClaim.AdminNotes,
		"updated_at":  updatedClaim.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Claim processed successfully",
		"data":    response,
	})
	logger.Log.SuccessCtx(c)
}

func UploadClaimProofImage(c *gin.Context) {
	id := c.Param("id")

	claimUUID, exists := pkg.GrabUuid(c, id, "CLAIMS", "claimId")
	if !exists {
		return
	}

	userId, ok := pkg.GrabUserId(c, "CLAIMS")
	if !ok {
		return
	}

	userUUID, exists := pkg.GrabUuid(c, userId, "CLAIMS", "userId")
	if !exists {
		return
	}

	req, okValidate := pkg.ValidateRequest[models.UploadClaimProofImageRequest](c)
	if !okValidate {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "CLAIMS") {
		return
	}
	defer conn.Release()

	q := db.New()

	// Verify claim exists and belongs to the user
	claim, err := q.FetchClaimByID(ctx, conn, claimUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": "Claim not found",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[CLAIMS-ERROR] Failed to fetch claim by ID", err)
		return
	}

	// Verify the user is the claimant
	if claim.ClaimantID != userUUID {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"message": "You are not authorized to upload proof images for this claim",
		})
		logger.Log.WarnCtx(c, "[CLAIMS-WARN] Unauthorized proof image upload attempt")
		return
	}

	// Generate unique object key
	fileExt := ".jpg"
	switch req.ContentType {
	case "image/png":
		fileExt = ".png"
	case "image/webp":
		fileExt = ".webp"
	}

	imageUUID := uuid.New().String()
	objectKey := fmt.Sprintf("claims/%s/proof_%s%s", claimUUID.String(), imageUUID, fileExt)

	// Generate presigned URL (valid for 15 minutes)
	presignedUrl, err := storage.S3.GeneratePresignedPutURL(ctx, objectKey, 15*time.Minute)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to generate upload URL",
		})
		logger.Log.ErrorCtx(c, "[CLAIMS-ERROR] Failed to generate presigned URL", err)
		return
	}

	// Update claim record with proof image key
	_, err = q.UpdateClaimProofImage(ctx, conn, db.UpdateClaimProofImageParams{
		ID:            claimUUID,
		ProofImageUrl: pgtype.Text{String: objectKey, Valid: true},
		ClaimantID:    userUUID,
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later",
		})
		logger.Log.ErrorCtx(c, "[CLAIMS-ERROR] Failed to update claim with proof image key", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Upload URL generated successfully",
		"data": models.UploadClaimProofImageResponse{
			PresignedUrl: presignedUrl,
			ObjectKey:    objectKey,
		},
	})
	logger.Log.SuccessCtx(c)
}
