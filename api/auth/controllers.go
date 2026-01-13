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
	"github.com/jackc/pgx/v5/pgtype"
)

// FLOW: User Registration
// Receive name, email, password (hashed from frontend)
// Validate input (format, strength, uniqueness)
// Check if a verified user exists → error if yes
// Check if a pending onboarding already exists
// If yes, you can update OTP and expiry
// OR reject until old OTP expires
// Generate a secure random OTP + expiry timestamp
// Insert (or update) user_onboarding row
// Send OTP via email
// User submits OTP for verification
// Check OTP validity, expiration, and attempt count
// On valid OTP
// Create user in users table, set is_verified = TRUE
// update the onboarding record
// Return success

func RegisterUser(c *gin.Context) {
	req, ok := pkg.ValidateRequest[models.RegisterUserRequest](c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := cmd.DBPool.Begin(ctx)
	if pkg.HandleDbTxnErr(c, err, "AUTH") {
		return
	}
	defer pkg.RollbackTx(c, tx, ctx, "AUTH")

	q := db.New(tx)

	// check if email is already registered in users table
	exists, err := q.CheckEmailExists(ctx, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[AUTH-ERROR]: Failed to check existing email", err)
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{
			"message": "Email is already registered",
		})
		logger.Log.InfoCtx(c, "[AUTH-INFO]: Registration attempt with existing email")
		return
	}

	// check if a pending onboarding already exists
	pending, err := q.CheckPendingOnboarding(ctx, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[AUTH-ERROR]: Failed to check pending onboarding", err)
		return
	}
	if pending {
		logger.Log.InfoCtx(c, "[AUTH-INFO]: Updating OTP for pending onboarding")
	}

	// generate OTP + expiry
	otpStr, _, err := pkg.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[AUTH-ERROR]: Unable to generate OTP", err)
		return
	}
	expiry := time.Now().Add(5 * time.Minute)

	// hash password before storing
	hashedPass, err := pkg.Hash(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[AUTH-ERROR]: Failed to hash password", err)
		return
	}

	// upsert onboarding with name, email, hashed password, OTP and expiry
	result, err := q.UpsertUserOnboarding(ctx, db.UpsertUserOnboardingParams{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPass,
		Otp:      otpStr,
		ExpiresAt: pgtype.Timestamptz{
			Time:  expiry,
			Valid: true,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[AUTH-ERROR]: Failed to upsert user onboarding", err)
		return
	}

	// commit transaction
	err = tx.Commit(ctx)
	if pkg.HandleDbTxnCommitErr(c, err, "AUTH") {
		return
	}

	// set temp token cookie for OTP flow
	tempToken := pkg.CreateTempToken(result.Email)
	pkg.SetTempCookie(c, tempToken)

	// TODO: send OTP via email (do this outside the transaction)
	// otpSlice is commented out for now
	// you can use otpSlice to send the actual code via your emailer

	c.JSON(http.StatusOK, gin.H{
		"message":   "Registration successful, please verify your email with the OTP sent.",
		"expiry_at": expiry,
	})
	logger.Log.InfoCtx(c, "[AUTH-SUCCESS]: User onboarded, OTP sent")
}
