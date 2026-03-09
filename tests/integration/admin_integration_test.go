package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KiranRajeev-KV/nyx-backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminFlow_Integration(t *testing.T) {
	cleanDB(t)

	adminCookie := createAdmin(t, "admin_test@test.com")
	userCookie := createAndAuthUser(t, "normal_user@test.com")
	reporterCookie := createAndAuthUser(t, "reporter@test.com")

	// Pre-create a Hub to attach to FOUND items
	var defaultHubID string
	err := testDBPool.QueryRow(context.Background(), "INSERT INTO hubs (name, address, contact, longitude, latitude) VALUES ('Admin Test Hub', '123 Admin Way', '555-0000', '12.34', '56.78') RETURNING id").Scan(&defaultHubID)
	require.NoError(t, err)

	// Pre-create an item to claim
	var itemID string
	t.Run("Setup: Create Claimable Item", func(t *testing.T) {
		itemReq := models.CreateItemRequest{
			Name:        "Claimable Watch",
			Description: "Silver watch",
			Type:        "FOUND",
			Location:    "University Gym",
			HubId:       &defaultHubID,
		}

		body, _ := json.Marshal(itemReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/items/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		addCookies(req, reporterCookie)
		testRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		itemID = response["data"].(map[string]interface{})["id"].(string)
	})

	var claimID string
	t.Run("Setup: User Creates Claim", func(t *testing.T) {
		claimReq := models.CreateClaimRequest{
			ItemID:    itemID,
			ProofText: "I have the matching receipt for this watch from 2024.",
		}

		body, _ := json.Marshal(claimReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/claims/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		addCookies(req, userCookie)
		testRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		claimID = response["data"].(map[string]interface{})["id"].(string)
	})

	t.Run("INT-ADMIN-004: Admin Boundary (Forbidden for USER)", func(t *testing.T) {
		// Attempt to process claim as normal user
		processReq := models.ProcessClaimRequest{
			Status: "APPROVED",
		}
		body, _ := json.Marshal(processReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/api/v1/claims/"+claimID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		addCookies(req, userCookie)

		testRouter.ServeHTTP(w, req)
		// Custom roles middleware might return 401 or 403
		assert.Contains(t, []int{http.StatusForbidden, http.StatusUnauthorized}, w.Code)
	})

	t.Run("INT-ADMIN-001: Approve Claim", func(t *testing.T) {
		processReq := models.ProcessClaimRequest{
			Status:     "APPROVED",
			AdminNotes: "Receipt verified",
		}
		body, _ := json.Marshal(processReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/api/v1/claims/"+claimID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		addCookies(req, adminCookie)

		testRouter.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify item status was set to RESOLVED
		var itemStatus string
		err := testDBPool.QueryRow(context.Background(), "SELECT status FROM items WHERE id = $1", itemID).Scan(&itemStatus)
		require.NoError(t, err)
		assert.Equal(t, "RESOLVED", itemStatus)
	})

	// Setup another claim to test rejection since previous was resolved
	var claim2ID string
	var item2ID string
	t.Run("Setup: Secondary Claim for Rejection", func(t *testing.T) {
		itemReq := models.CreateItemRequest{Name: "Item2", Type: "FOUND", Description: "Valid test description", Location: "Valid location length", HubId: &defaultHubID}
		body, _ := json.Marshal(itemReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/items/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		addCookies(req, reporterCookie)
		testRouter.ServeHTTP(w, req)
		var res1 map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &res1)
		item2ID = res1["data"].(map[string]interface{})["id"].(string)

		claimReq := models.CreateClaimRequest{ItemID: item2ID, ProofText: "Fake proof"}
		body2, _ := json.Marshal(claimReq)
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("POST", "/api/v1/claims/", bytes.NewBuffer(body2))
		req2.Header.Set("Content-Type", "application/json")
		addCookies(req2, userCookie)
		testRouter.ServeHTTP(w2, req2)
		var res2 map[string]interface{}
		json.Unmarshal(w2.Body.Bytes(), &res2)
		claim2ID = res2["data"].(map[string]interface{})["id"].(string)
	})

	t.Run("INT-ADMIN-002: Reject Claim", func(t *testing.T) {
		processReq := models.ProcessClaimRequest{
			Status:     "REJECTED",
			AdminNotes: "Insufficient proof",
		}
		body, _ := json.Marshal(processReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/api/v1/claims/"+claim2ID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		addCookies(req, adminCookie)

		testRouter.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify item status falls back to OPEN if no other pending claims (logic defined in requirements)
		var itemStatus string
		err := testDBPool.QueryRow(context.Background(), "SELECT status FROM items WHERE id = $1", item2ID).Scan(&itemStatus)
		require.NoError(t, err)
		assert.Equal(t, "OPEN", itemStatus)
	})

	t.Run("INT-ADMIN-003: Hub Moderation constraints", func(t *testing.T) {
		// 1. Admin creates Hub
		hubReq := models.CreateHubRequest{Name: "Constraint Hub", Address: "12345 Valid Long Address Lane", Contact: "123-456-7890", Longitude: "1.0", Latitude: "1.0"}
		hBody, _ := json.Marshal(hubReq)
		hw := httptest.NewRecorder()
		hreq, _ := http.NewRequest("POST", "/api/v1/hubs/", bytes.NewBuffer(hBody))
		hreq.Header.Set("Content-Type", "application/json")
		addCookies(hreq, adminCookie)
		testRouter.ServeHTTP(hw, hreq)

		var hubID string
		testDBPool.QueryRow(context.Background(), "SELECT id FROM hubs WHERE name = 'Constraint Hub'").Scan(&hubID)

		// 2. User creates FOUND item referencing Hub
		itemReq := models.CreateItemRequest{Name: "Found at Hub", Type: "FOUND", Location: "Hub campus area", Description: "Found this right outside the hub area", HubId: &hubID}
		iBody, _ := json.Marshal(itemReq)
		iw := httptest.NewRecorder()
		ireq, _ := http.NewRequest("POST", "/api/v1/items/", bytes.NewBuffer(iBody))
		ireq.Header.Set("Content-Type", "application/json")
		addCookies(ireq, userCookie)
		testRouter.ServeHTTP(iw, ireq)
		require.Equal(t, http.StatusCreated, iw.Code)

		// 3. Admin attempts to delete hub
		dw := httptest.NewRecorder()
		dreq, _ := http.NewRequest("DELETE", "/api/v1/hubs/"+hubID, nil)
		addCookies(dreq, adminCookie)
		testRouter.ServeHTTP(dw, dreq)

		// Due to RESTRICT foreign key, this returns an error constraint violation
		// Should be caught by the backend and returned as 409 Conflict (or 500 depending on generic error handling)
		assert.Contains(t, []int{http.StatusConflict, http.StatusInternalServerError, http.StatusBadRequest}, dw.Code)
	})
}
