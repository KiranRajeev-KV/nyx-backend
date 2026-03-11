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

// REGRESSION: Verify duplicate claims on same items are rejected
func TestRegression_Logic_DuplicateClaimPrevention(t *testing.T) {
	cleanDB(t)

	// Create users
	ownerCookie := createAndAuthUser(t, "owner@nyx.com")
	finderCookie := createAndAuthUser(t, "finder@nyx.com")

	// Hub
	var hubID string
	err := testDBPool.QueryRow(context.Background(), "INSERT INTO hubs (name, address, contact, longitude, latitude) VALUES ('Reg Hub', 'Reg Address', '123', '0.0', '0.0') RETURNING id").Scan(&hubID)
	require.NoError(t, err)

	// FOUND
	foundReq := models.CreateItemRequest{Name: "Found Reg", Type: "FOUND", Description: "Valid Description Here", Location: "Valid Location Here", HubId: &hubID}
	fBody, _ := json.Marshal(foundReq)
	fW := httptest.NewRecorder()
	fR, _ := http.NewRequest("POST", "/api/v1/items/", bytes.NewBuffer(fBody))
	fR.Header.Set("Content-Type", "application/json")
	addCookies(fR, finderCookie)
	testRouter.ServeHTTP(fW, fR)
	var fRes map[string]interface{}
	json.Unmarshal(fW.Body.Bytes(), &fRes)
	foundID := fRes["data"].(map[string]interface{})["id"].(string)

	// LOST
	lostReq := models.CreateItemRequest{Name: "Lost Reg", Type: "LOST", Description: "Valid Description Here", Location: "Valid Location Here"}
	lBody, _ := json.Marshal(lostReq)
	lW := httptest.NewRecorder()
	lR, _ := http.NewRequest("POST", "/api/v1/items/", bytes.NewBuffer(lBody))
	lR.Header.Set("Content-Type", "application/json")
	addCookies(lR, ownerCookie)
	testRouter.ServeHTTP(lW, lR)
	var lRes map[string]interface{}
	json.Unmarshal(lW.Body.Bytes(), &lRes)
	lostID := lRes["data"].(map[string]interface{})["id"].(string)

	// CLAIM 1 (Should Pass)
	claimReq := models.CreateClaimRequest{FoundItemID: foundID, LostItemID: lostID, ProofText: "Original valid proof text"}
	cBody, _ := json.Marshal(claimReq)
	cW1 := httptest.NewRecorder()
	cR1, _ := http.NewRequest("POST", "/api/v1/claims/", bytes.NewBuffer(cBody))
	cR1.Header.Set("Content-Type", "application/json")
	addCookies(cR1, ownerCookie)
	testRouter.ServeHTTP(cW1, cR1)
	assert.Equal(t, http.StatusCreated, cW1.Code)

	// CLAIM 2 (Should Fail as Duplicate Constraint)
	cW2 := httptest.NewRecorder()
	cR2, _ := http.NewRequest("POST", "/api/v1/claims/", bytes.NewBuffer(cBody))
	cR2.Header.Set("Content-Type", "application/json")
	addCookies(cR2, ownerCookie)
	testRouter.ServeHTTP(cW2, cR2)
	assert.Contains(t, []int{http.StatusConflict, http.StatusBadRequest, http.StatusInternalServerError}, cW2.Code)
}

// REGRESSION: Verify Hub Public List
func TestRegression_Logic_HubPublicList(t *testing.T) {
	cleanDB(t)

	// Create 2 hubs quietly
	testDBPool.Exec(context.Background(), "INSERT INTO hubs (name, address, contact, longitude, latitude) VALUES ('Hub1', 'L', '1', '1', '1')")
	testDBPool.Exec(context.Background(), "INSERT INTO hubs (name, address, contact, longitude, latitude) VALUES ('Hub2', 'L', '1', '1', '1')")

	// Verify unauthenticated user can hit list hubs GET route
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/hubs/", nil)
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Should return length >= 2
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	dataList := response["data"].([]interface{})
	assert.True(t, len(dataList) >= 2)
}
