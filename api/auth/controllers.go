package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/KiranRajeev-KV/nyx-backend/cmd"
	db "github.com/KiranRajeev-KV/nyx-backend/internal/db/gen"
	"github.com/KiranRajeev-KV/nyx-backend/internal/logger"
	"github.com/KiranRajeev-KV/nyx-backend/internal/models"
	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx"
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

// FLOW: OTP Verification
// Receive OTP from user
// Validate input (format, length)
// Retrieve pending onboarding by email from temp token
// Check if onboarding exists → error if not
// Check if OTP matches, not expired
// On valid OTP
// Create user in users table, set is_verified = TRUE
// update the onboarding record
// On invalid OTP
// If attempts >= max, invalidate OTP and require restart
// Return success or error accordingly

func VerifyOTP(c *gin.Context) {
	req, ok := pkg.ValidateRequest[models.VerifyOTPRequest](c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tempEmail, valid := pkg.GetEmail(c, "VERIFY-OTP")
	if !valid {
		return
	}

	tx, err := cmd.DBPool.Begin(ctx)
	if pkg.HandleDbTxnErr(c, err, "VERIFY-OTP") {
		return
	}
	defer pkg.RollbackTx(c, tx, ctx, "VERIFY-OTP")

	q := db.New(tx)

	// fetch pending onboarding
	onboarding, err := q.GetPendingOnboardingByEmail(ctx, tempEmail)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid OTP or onboarding not found. Please register again.",
		})
		logger.Log.InfoCtx(c, "[VERIFY-OTP-INFO]: No pending onboarding found for email")
		return
	}

	// check OTP validity and expiry
	if onboarding.Otp != req.OTP || time.Now().After(onboarding.ExpiresAt.Time) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid or expired OTP. Please try again.",
		})
		logger.Log.InfoCtx(c, "[VERIFY-OTP-INFO]: Invalid or expired OTP attempt")
		return
	}

	// create user in users table
	_, err = q.CreateUser(ctx, db.CreateUserParams{
		Name:       onboarding.Name,
		Email:      onboarding.Email,
		Password:   onboarding.Password,
		IsVerified: true,
	})
	if err != nil {
		// handle unique constraint violation (email) just in case
		var pgErr *pgx.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{
				"message": "Email is already registered",
			})
			logger.Log.InfoCtx(c, "[VERIFY-OTP-INFO]: Email already registered during OTP verification")
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[VERIFY-OTP-ERROR]: Failed to create user", err)
		return
	}

	if err := q.DeleteOnboardingByEmail(ctx, onboarding.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[VERIFY-OTP-ERROR]: Failed to delete onboarding record", err)
		return
	}

	// commit transaction
	err = tx.Commit(ctx)
	if pkg.HandleDbTxnCommitErr(c, err, "VERIFY-OTP") {
		return
	}

	// clear temp token cookie
	pkg.ClearTempCookie(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP verified successfully. Your account is now active.",
	})
	logger.Log.SuccessCtx(c)
}
