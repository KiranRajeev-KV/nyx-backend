package pkg

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	paseto "aidanwoods.dev/go-paseto"

	"github.com/KiranRajeev-KV/nyx-backend/cmd"
	db "github.com/KiranRajeev-KV/nyx-backend/internal/db/gen"
	"github.com/KiranRajeev-KV/nyx-backend/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

/*
	JTI stores the email
	Audience stores the userId
*/

const (
	RefreshTokenValidTime = time.Hour * 24 * 90
	AuthTokenValidTime    = time.Hour * 1
	TempTokenValidTime    = time.Minute * 5
	privateKeyPath        = "app.rsa"
	publicKeyPath         = "app.pub.rsa"
)

var (
	VerifyKey paseto.V4AsymmetricPublicKey
	SignKey   paseto.V4AsymmetricSecretKey
)

type Role string

const (
	RoleUser  Role = "USER"
	RoleAdmin Role = "ADMIN"
)

func InitPaseto() error {
	privateKeyBinary, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return err
	}
	privateKeyHex := hex.EncodeToString(privateKeyBinary)

	publicKeyBinary, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return err
	}
	publicKeyHex := hex.EncodeToString(publicKeyBinary)

	// Verify using public key
	VerifyKey, err = paseto.NewV4AsymmetricPublicKeyFromHex(publicKeyHex)
	if err != nil {
		return fmt.Errorf("Error in public-paseto: %w", err)
	}
	// Sign using private key
	SignKey, err = paseto.NewV4AsymmetricSecretKeyFromHex(privateKeyHex)
	if err != nil {
		return fmt.Errorf("Error in private-paseto: %w", err)
	}
	return nil
}

func CreateAuthToken(userId, email string, role db.UserRole) (string, error) {
	token := paseto.NewToken()

	token.SetJti(email)
	token.SetAudience(userId)
	token.SetIssuer("NYX-BACKEND")
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(time.Now().Add(AuthTokenValidTime))
	token.SetSubject("access_token")
	token.Set("role", role)

	signed := token.V4Sign(SignKey, nil)
	return signed, nil
}

func CreateRefreshToken(userId, email string, role db.UserRole) (string, error) {

	token := paseto.NewToken()
	token.SetJti(email)
	token.SetAudience(userId)
	token.SetIssuer("NYX-BACKEND")
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(time.Now().Add(RefreshTokenValidTime))
	token.SetSubject("refresh_token")
	token.Set("role", role)

	signed := token.V4Sign(SignKey, nil)
	return signed, nil
}

func CreateTempToken(email string) string {
	token := paseto.NewToken()

	token.SetJti(email)
	token.SetIssuer("NYX-BACKEND")
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(time.Now().Add(TempTokenValidTime))
	token.SetSubject("temp_token")

	signed := token.V4Sign(SignKey, nil)
	return signed
}

func ParseToken(token, tokenType string) (bool, *paseto.Token) {
	parser := paseto.NewParser()

	parser.AddRule(paseto.IssuedBy("NYX-BACKEND"))
	parser.AddRule(paseto.Subject(tokenType))
	parser.AddRule(paseto.ValidAt(time.Now()))
	parser.AddRule(paseto.NotExpired())

	parsedToken, err := parser.ParseV4Public(VerifyKey, token, nil)
	if err != nil {
		return false, nil
	}
	return true, parsedToken
}

func VerifyTokens(c *gin.Context, authToken, refreshToken string) bool {
	ok, parsedAuthToken := ParseToken(authToken, "access_token")
	if !ok {
		return false
	}
	ok, parsedRefToken := ParseToken(refreshToken, "refresh_token")
	if !ok {
		return false
	}

	authData := parsedAuthToken.Claims()
	refData := parsedRefToken.Claims()

	authRole := authData["role"].(string)
	refRole := refData["role"].(string)

	// Verification conditions
	c1 := authData["aud"] != refData["aud"]
	c2 := authData["jti"] != refData["jti"]
	c3 := authRole != refRole

	if c1 || c2 || c3 {
		return false
	}

	// Setting up variables in *gin.Context for passing around in handlers
	c.Set("userId", refData["aud"])
	c.Set("email", refData["jti"])
	c.Set("role", refRole)

	return true
}

func VerifyTempToken(c *gin.Context, tempToken string) bool {
	ok, parsedTempToken := ParseToken(tempToken, "temp_token")
	if !ok {
		return false
	}

	tempData := parsedTempToken.Claims()
	c.Set("email", tempData["jti"])

	return true
}

func VerifyRefreshToken(c *gin.Context, refreshToken string) (*paseto.Token, error) {
	ok, parsedRefToken := ParseToken(refreshToken, "refresh_token")
	if !ok {
		logger.Log.ErrorCtx(c, "[REQ-ERROR]: Failed to parse refresh_token", nil)
		return nil, fmt.Errorf("failed to parse refresh token")
	}

	refreshClaims := parsedRefToken.Claims()
	email := refreshClaims["jti"].(string)
	role := refreshClaims["role"].(string)
	c.Set("role", role)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	q := db.New(conn)

	// Possible scenarios
	// 1. RefreshToken does not exist
	// 2. RefreshToken has become invalid
	// 3. RefreshToken is perfect and it can generate AuthToken
	var token pgtype.Text

	token, err = q.CheckRefreshTokenQuery(ctx, email)
	if err != nil {
		logger.Log.FatalCtx(c, "[AUTH-ERROR] Failed to fetch refresh token from DB", err)
		return nil, err
	}

	if token.String == "" {
		return nil, fmt.Errorf("refresh token does not exist in DB")
	}

	ok, validToken := ParseToken(token.String, "refresh_token")
	if !ok {
		return nil, fmt.Errorf("[AUTH-ERROR]: Failed to parse refresh token")
	}

	return validToken, nil
}
