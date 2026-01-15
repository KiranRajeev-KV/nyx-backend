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
	if pkg.HandleDbTxnErr(c, err, "REGISTER") {
		return
	}
	defer pkg.RollbackTx(c, tx, ctx, "REGISTER")

	q := db.New()

	// check if email is already registered in users table
	exists, err := q.CheckEmailExists(ctx, tx, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[REGISTER-ERROR]: Failed to check existing email", err)
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{
			"message": "Email is already registered",
		})
		logger.Log.InfoCtx(c, "[REGISTER-INFO]: Registration attempt with existing email")
		return
	}

	// check if a pending onboarding already exists
	pending, err := q.CheckPendingOnboarding(ctx, tx, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[REGISTER-ERROR]: Failed to check pending onboarding", err)
		return
	}
	if pending {
		logger.Log.InfoCtx(c, "[REGISTER-INFO]: Updating OTP for pending onboarding")
	}

	// generate OTP + expiry
	otpStr, _, err := pkg.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[REGISTER-ERROR]: Unable to generate OTP", err)
		return
	}
	expiry := time.Now().Add(5 * time.Minute)

	// hash password before storing
	hashedPass, err := pkg.Hash(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[REGISTER-ERROR]: Failed to hash password", err)
		return
	}

	// upsert onboarding with name, email, hashed password, OTP and expiry
	result, err := q.UpsertUserOnboarding(ctx, tx, db.UpsertUserOnboardingParams{
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
		logger.Log.ErrorCtx(c, "[REGISTER-ERROR]: Failed to upsert user onboarding", err)
		return
	}

	// commit transaction
	err = tx.Commit(ctx)
	if pkg.HandleDbTxnCommitErr(c, err, "REGISTER") {
		return
	}

	// set temp token cookie for OTP flow
	tempToken := pkg.CreateTempToken(result.Email)
	pkg.SetTempCookie(c, tempToken)

	// TODO: send OTP via email (do this outside the transaction)
	// otpSlice is commented out for now
	// you can use otpSlice to send the actual code via your emailer

	c.JSON(http.StatusOK, gin.H{
		"message":   "Registration successful, please verify your email using the OTP.",
		"expiry_at": expiry,
	})
	logger.Log.SuccessCtx(c)
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

	q := db.New()

	// fetch pending onboarding
	onboarding, err := q.GetPendingOnboardingByEmail(ctx, tx, tempEmail)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid OTP or onboarding not found. Please register again.",
		})
		logger.Log.InfoCtx(c, "[VERIFY-OTP-INFO]: No pending onboarding found for email")
		return
	}

	// check expiry
	if time.Now().After(onboarding.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "OTP has expired. Please register again.",
		})
		logger.Log.InfoCtx(c, "[VERIFY-OTP-INFO]: Expired OTP attempt")
		return
	}

	// check OTP validity
	if onboarding.Otp != req.OTP {
		// Increment attempts
		err = q.IncrementOnboardingAttempts(ctx, tx, tempEmail)
		if err != nil {
			logger.Log.ErrorCtx(c, "[VERIFY-OTP-ERROR]: Failed to increment attempts", err)
		}

		if onboarding.Attempts.Int32 >= 3 {
			_ = q.DeleteOnboardingByEmail(ctx, tx, tempEmail)
			_ = tx.Commit(ctx)
			pkg.ClearTempCookie(c)
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Too many failed attempts. Please register again.",
			})
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid OTP. Please try again.",
		})
		logger.Log.InfoCtx(c, "[VERIFY-OTP-INFO]: Invalid OTP attempt")
		return
	}

	// create user in users table
	_, err = q.CreateUser(ctx, tx, db.CreateUserParams{
		Name:     onboarding.Name,
		Email:    onboarding.Email,
		Password: onboarding.Password,
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

	if err := q.DeleteOnboardingByEmail(ctx, tx, onboarding.Email); err != nil {
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

func ResendOTP(c *gin.Context) {
	tempEmail, valid := pkg.GetEmail(c, "RESEND-OTP")
	if !valid {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := cmd.DBPool.Begin(ctx)
	if pkg.HandleDbTxnErr(c, err, "RESEND-OTP") {
		return
	}
	defer pkg.RollbackTx(c, tx, ctx, "RESEND-OTP")

	q := db.New()

	// 1. Fetch existing onboarding data
	onboarding, err := q.GetPendingOnboardingByEmail(ctx, tx, tempEmail)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Onboarding session not found. Please register again.",
		})
		logger.Log.InfoCtx(c, "[RESEND-OTP-INFO]: No pending onboarding found for email")
		return
	}

	// 2. Generate new OTP + expiry
	otpStr, _, err := pkg.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[RESEND-OTP-ERROR]: Unable to generate OTP", err)
		return
	}
	expiry := time.Now().Add(5 * time.Minute)

	// 3. Upsert (update) the record with new OTP and reset attempts
	_, err = q.UpsertUserOnboarding(ctx, tx, db.UpsertUserOnboardingParams{
		Name:     onboarding.Name,
		Email:    onboarding.Email,
		Password: onboarding.Password,
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
		logger.Log.ErrorCtx(c, "[RESEND-OTP-ERROR]: Failed to update user onboarding", err)
		return
	}

	// 4. Commit transaction
	err = tx.Commit(ctx)
	if pkg.HandleDbTxnCommitErr(c, err, "RESEND-OTP") {
		return
	}

	// TODO: send OTP via email (outside transaction)

	c.JSON(http.StatusOK, gin.H{
		"message": "A new OTP has been sent to your email.",
	})
	logger.Log.SuccessCtx(c)
}

func LoginUser(c *gin.Context) {
	req, ok := pkg.ValidateRequest[models.LoginUserRequest](c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "LOGIN") {
		return
	}
	defer conn.Release()

	q := db.New()

	// fetch user by email
	user, err := q.GetUserByEmail(ctx, conn, req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid email or password.",
		})
		logger.Log.InfoCtx(c, "[LOGIN-INFO]: Invalid login attempt - email not found")
		return
	}

	// verify password
	if err := pkg.CompareHash(user.Password, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid email or password.",
		})
		logger.Log.InfoCtx(c, "[LOGIN-INFO]: Invalid login attempt - wrong password")
		return
	}

	// create paseto tokens
	accessToken, err := pkg.CreateAuthToken(user.ID.String(), user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[LOGIN-ERROR]: Failed to create auth token", err)
		return
	}

	refreshToken, err := pkg.CreateRefreshToken(user.ID.String(), user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[LOGIN-ERROR]: Failed to create refresh token", err)
		return
	}

	// set refresh token in db
	err = q.SetUserRefreshToken(ctx, conn, db.SetUserRefreshTokenParams{
		ID:           user.ID,
		RefreshToken: pgtype.Text{String: refreshToken, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[LOGIN-ERROR]: Failed to set refresh token in DB", err)
		return
	}

	// set tokens in cookies
	pkg.SetAuthCookie(c, accessToken)
	pkg.SetRefreshCookie(c, refreshToken)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful.",
	})
	logger.Log.SuccessCtx(c)
}

func LogoutUser(c *gin.Context) {
	pkg.NullifyCookies(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful.",
	})
	logger.Log.SuccessCtx(c)
}

func FetchUserSession(c *gin.Context) {
	email, ok := pkg.GetEmail(c, "SESSION")
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if pkg.HandleDbAcquireErr(c, err, "SESSION") {
		return
	}
	defer conn.Release()

	q := db.New()

	result, err := q.FetchUserSession(ctx, conn, email)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "No session found for user",
		})
		logger.Log.WarnCtx(c, "[SESSION-WARN]: User might deleted but cookies exist")
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.ErrorCtx(c, "[SESSION-ERROR]: Failed to fetch user session", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User session obtained successfully",
		"name":    result.Name,
		"email":   result.Email,
		"role":    result.Role,
		"id":      result.ID.String(),
	})
	logger.Log.SuccessCtx(c)
}
