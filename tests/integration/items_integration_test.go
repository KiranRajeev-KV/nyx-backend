package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/KiranRajeev-KV/nyx-backend/internal/db/gen"
	"github.com/KiranRajeev-KV/nyx-backend/internal/models"
	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create and authenticate a test user
func createAndAuthUser(t *testing.T, email string) []*http.Cookie {
	// Bypass OTP flow by inserting directly into DB

	// Ensure user exists
	_, err := testDBPool.Exec(context.Background(), `
		INSERT INTO users (name, email, password, is_verified) 
		VALUES ($1, $2, 'hashed_password_mock', true)
		ON CONFLICT (email) DO NOTHING
	`, "Test User", email)
	require.NoError(t, err)

	// Because LoginUser requires matched hashed password, and we inserted fake hash,
	// let's create a generic user test helper instead that just mocks the login cookie directly
	// for items tests so we don't depend on auth's bcrypt cost.

	// Fetch actual user ID created or existing
	var userID string
	err = testDBPool.QueryRow(context.Background(), "SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	require.NoError(t, err)

	// Since we mock JWTs in gin context middleware for auth in tests commonly or generate actual PASETO
	// We should just use the actual package to generate valid token
	return generateAssumedPasetoCookies(t, userID, email, "USER")
}

func createAdmin(t *testing.T, email string) []*http.Cookie {
	_, err := testDBPool.Exec(context.Background(), `
		INSERT INTO users (name, email, password, is_verified, role) 
		VALUES ($1, $2, 'hashed_password_mock', true, 'ADMIN')
		ON CONFLICT (email) DO UPDATE SET role = 'ADMIN'
	`, "Admin User", email)
	require.NoError(t, err)

	var userID string
	err = testDBPool.QueryRow(context.Background(), "SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	require.NoError(t, err)

	return generateAssumedPasetoCookies(t, userID, email, "ADMIN")
}

func generateAssumedPasetoCookies(t *testing.T, id string, email string, role string) []*http.Cookie {
	accToken, _ := pkg.CreateAuthToken(id, email, db.UserRole(role))
	refToken, _ := pkg.CreateRefreshToken(id, email, db.UserRole(role))

	return []*http.Cookie{
		{
			Name:  "access_token",
			Value: accToken,
		},
		{
			Name:  "refresh_token",
			Value: refToken,
		},
	}
}

func addCookies(req *http.Request, cookies []*http.Cookie) {
	for _, c := range cookies {
		req.AddCookie(c)
	}
}

func TestItemsFlow_Integration(t *testing.T) {
	cleanDB(t)

	accessCookie := createAndAuthUser(t, "item_user@test.com")

	var createdLostItemID string
	var createdHubID string

	// Create a hub to use for foundational testing
	t.Run("Setup: Create Hub", func(t *testing.T) {
		adminCookie := createAdmin(t, "admin@test.com")

		hubReq := models.CreateHubRequest{
			Name:      "Test Campus Hub",
			Address:   "123 Test Ave",
			Contact:   "1234567890",
			Longitude: "12.345",
			Latitude:  "67.890",
		}

		body, _ := json.Marshal(hubReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/hubs/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		addCookies(req, adminCookie)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		// Assuming we can grab ID from DB, let's just fetch it
		err := testDBPool.QueryRow(context.Background(), "SELECT id FROM hubs LIMIT 1").Scan(&createdHubID)
		require.NoError(t, err)
	})

	t.Run("INT-ITEM-001: Create LOST item", func(t *testing.T) {
		itemReq := models.CreateItemRequest{
			Name:        "Lost Keys",
			Description: "Set of 3 keys with a blue keychain",
			Type:        "LOST",
			Location:    "Library",
		}

		body, _ := json.Marshal(itemReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/items/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		addCookies(req, accessCookie)

		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		createdLostItemID = response["data"].(map[string]interface{})["id"].(string)

		// Verify fetch by ID
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/api/v1/items/"+createdLostItemID, nil)
		addCookies(req2, accessCookie)
		testRouter.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)
		var getResponse map[string]interface{}
		json.Unmarshal(w2.Body.Bytes(), &getResponse)

		itemData := getResponse["data"].(map[string]interface{})
		assert.Equal(t, "Lost Keys", itemData["name"])
		assert.Equal(t, "LOST", itemData["type"])
	})

	t.Run("INT-ITEM-002: Create FOUND item with constraints", func(t *testing.T) {
		// First try without hub
		itemReq := models.CreateItemRequest{
			Name:        "Found Wallet",
			Description: "Black leather wallet",
			Type:        "FOUND",
			Location:    "Cafeteria",
		}

		body, _ := json.Marshal(itemReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/items/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		addCookies(req, accessCookie)

		testRouter.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code, "Should require hub implicitly or explicitly for FOUND")

		// Now with hub
		itemReq.HubId = &createdHubID
		body2, _ := json.Marshal(itemReq)
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("POST", "/api/v1/items/", bytes.NewBuffer(body2))
		req2.Header.Set("Content-Type", "application/json")
		addCookies(req2, accessCookie)

		testRouter.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusCreated, w2.Code, "Should succeed with hub attached")
	})

	t.Run("INT-DISC-001 & 002: Filter items by type", func(t *testing.T) {
		// Filter LOST
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/items/?type=LOST", nil)
		addCookies(req, accessCookie)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var lostItems []map[string]interface{}
		// Depends on actual return structure. If wrapped in "data":
		var wrap map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &wrap)

		if wrap["data"] != nil {
			bytes, _ := json.Marshal(wrap["data"])
			json.Unmarshal(bytes, &lostItems)
		} else {
			json.Unmarshal(w.Body.Bytes(), &lostItems)
		}

		assert.True(t, len(lostItems) >= 1)
		for _, item := range lostItems {
			assert.Equal(t, "LOST", item["type"])
		}

		// Filter FOUND
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/api/v1/items/?type=FOUND", nil)
		addCookies(req2, accessCookie)
		testRouter.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)

		var foundItems []map[string]interface{}
		if wrap["data"] != nil {
			var wrap2 map[string]interface{}
			json.Unmarshal(w2.Body.Bytes(), &wrap2)
			bytes2, _ := json.Marshal(wrap2["data"])
			json.Unmarshal(bytes2, &foundItems)
		} else {
			json.Unmarshal(w2.Body.Bytes(), &foundItems)
		}

		assert.True(t, len(foundItems) >= 1)
		for _, item := range foundItems {
			assert.Equal(t, "FOUND", item["type"])
		}
	})

	t.Run("INT-DISC-003: Invalid Filter", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/items/?type=INVALID_TYPE", nil)
		addCookies(req, accessCookie)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("INT-ITEM-004: Unauthorized Update", func(t *testing.T) {
		otherUserCookie := createAndAuthUser(t, "other_user@test.com")

		nameStr := "Should not update"
		updateReq := models.UpdateItemRequest{
			Name: &nameStr,
		}

		body, _ := json.Marshal(updateReq)
		w := httptest.NewRecorder()
		// Try to update the first created item (createdLostItemID) which belongs to first user
		req, _ := http.NewRequest("PATCH", "/api/v1/items/"+createdLostItemID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		addCookies(req, otherUserCookie)

		testRouter.ServeHTTP(w, req)

		// Could be 403 or 404 depending on how the query is structured
		assert.Contains(t, []int{http.StatusForbidden, http.StatusNotFound}, w.Code)
	})

	t.Run("INT-ITEM-003: Owner Update and Delete", func(t *testing.T) {
		// Update
		descStr := "Updated description for testing"
		updateReq := models.UpdateItemRequest{
			Description: &descStr,
		}
		body, _ := json.Marshal(updateReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/api/v1/items/"+createdLostItemID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		addCookies(req, accessCookie)

		testRouter.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Delete
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("DELETE", "/api/v1/items/"+createdLostItemID, nil)
		addCookies(req2, accessCookie)

		testRouter.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)

		// Verify Status is DELETED using DB query
		var status string
		err := testDBPool.QueryRow(context.Background(), "SELECT status FROM items WHERE id = $1", createdLostItemID).Scan(&status)
		require.NoError(t, err)
		assert.Equal(t, "DELETED", status)
	})
}
