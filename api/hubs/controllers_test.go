package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	api "github.com/KiranRajeev-KV/nyx-backend/api/hubs"
	"github.com/KiranRajeev-KV/nyx-backend/internal/models"
	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/KiranRajeev-KV/nyx-backend/tests"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	tests.InitTestLogger()
	// Change to project root for RSA keys
	os.Chdir("/home/kr/dev/nyx-backend")
	pkg.InitPaseto()
}

// setupCreateHubTest sets up common test dependencies for create hub tests
func setupCreateHubTest() *gin.Engine {
	router := tests.NewTestRouter()
	router.POST("/hubs", api.CreateHub)
	return router
}

func TestCreateHub_ValidRequest_RequiresDBSetup(t *testing.T) {
	t.Skip("Requires DB setup - use integration tests instead")
}

func TestCreateHub_InvalidJSON_BadRequest(t *testing.T) {
	router := setupCreateHubTest()

	req := httptest.NewRequest(http.MethodPost, "/hubs", bytes.NewBufferString(`{invalid json}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateHub_MissingName_BadRequest(t *testing.T) {
	router := setupCreateHubTest()

	reqBody := models.CreateHubRequest{
		Address: "123 Main Street Suite 100",
		Contact: "contact@hub.com",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/hubs", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateHub_NameTooShort_BadRequest(t *testing.T) {
	router := setupCreateHubTest()

	reqBody := models.CreateHubRequest{
		Name:    "AB",
		Address: "123 Main Street Suite 100",
		Contact: "contact@hub.com",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/hubs", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateHub_NameTooLong_BadRequest(t *testing.T) {
	router := setupCreateHubTest()

	longName := "a"
	for i := 0; i < 110; i++ {
		longName += "a"
	}

	reqBody := models.CreateHubRequest{
		Name:    longName,
		Address: "123 Main Street Suite 100",
		Contact: "contact@hub.com",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/hubs", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateHub_MissingAddress_BadRequest(t *testing.T) {
	router := setupCreateHubTest()

	reqBody := models.CreateHubRequest{
		Name:    "Main Hub",
		Contact: "contact@hub.com",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/hubs", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateHub_AddressTooShort_BadRequest(t *testing.T) {
	router := setupCreateHubTest()

	reqBody := models.CreateHubRequest{
		Name:    "Main Hub",
		Address: "123",
		Contact: "contact@hub.com",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/hubs", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateHub_AddressTooLong_BadRequest(t *testing.T) {
	router := setupCreateHubTest()

	longAddress := "123 Main Street"
	for i := 0; i < 210; i++ {
		longAddress += "a"
	}

	reqBody := models.CreateHubRequest{
		Name:    "Main Hub",
		Address: longAddress,
		Contact: "contact@hub.com",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/hubs", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateHub_MissingContact_BadRequest(t *testing.T) {
	router := setupCreateHubTest()

	reqBody := models.CreateHubRequest{
		Name:    "Main Hub",
		Address: "123 Main Street Suite 100",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/hubs", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateHub_ContactTooShort_BadRequest(t *testing.T) {
	router := setupCreateHubTest()

	reqBody := models.CreateHubRequest{
		Name:    "Main Hub",
		Address: "123 Main Street Suite 100",
		Contact: "abc",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/hubs", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateHub_ContactTooLong_BadRequest(t *testing.T) {
	router := setupCreateHubTest()

	longContact := "a"
	for i := 0; i < 60; i++ {
		longContact += "a"
	}

	reqBody := models.CreateHubRequest{
		Name:    "Main Hub",
		Address: "123 Main Street Suite 100",
		Contact: longContact,
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/hubs", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateHub_LongitudeTooLong_BadRequest(t *testing.T) {
	router := setupCreateHubTest()

	longLongitude := "a"
	for i := 0; i < 60; i++ {
		longLongitude += "a"
	}

	reqBody := models.CreateHubRequest{
		Name:      "Main Hub",
		Address:   "123 Main Street Suite 100",
		Contact:   "contact@hub.com",
		Longitude: longLongitude,
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/hubs", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateHub_LatitudeTooLong_BadRequest(t *testing.T) {
	router := setupCreateHubTest()

	longLatitude := "a"
	for i := 0; i < 60; i++ {
		longLatitude += "a"
	}

	reqBody := models.CreateHubRequest{
		Name:     "Main Hub",
		Address:  "123 Main Street Suite 100",
		Contact:  "contact@hub.com",
		Latitude: longLatitude,
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/hubs", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

// FetchHubs Tests

func setupFetchHubsTest() *gin.Engine {
	router := tests.NewTestRouter()
	router.GET("/hubs", api.FetchHubs)
	return router
}

func TestFetchHubs_ValidRequest_RequiresDBSetup(t *testing.T) {
	t.Skip("Requires DB setup - use integration tests instead")
}

// FetchHubById Tests

func setupFetchHubByIdTest() *gin.Engine {
	router := tests.NewTestRouter()
	router.GET("/hubs/:id", api.FetchHubById)
	return router
}

func TestFetchHubById_ValidRequest_RequiresDBSetup(t *testing.T) {
	t.Skip("Requires DB setup - use integration tests instead")
}

func TestFetchHubById_InvalidHubID_BadRequest(t *testing.T) {
	router := setupFetchHubByIdTest()

	tc := tests.ExecuteRequest(router, "GET", "/hubs/not-a-uuid")
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestFetchHubById_MissingHubID_BadRequest(t *testing.T) {
	router := setupFetchHubByIdTest()

	tc := tests.ExecuteRequest(router, "GET", "/hubs/")
	assert.Equal(t, http.StatusNotFound, tc.GetResponseStatus())
}

// UpdateHub Tests

func setupUpdateHubTest() *gin.Engine {
	router := tests.NewTestRouter()
	router.PATCH("/hubs/:id", api.UpdateHub)
	return router
}

func TestUpdateHub_ValidRequest_RequiresDBSetup(t *testing.T) {
	t.Skip("Requires DB setup - use integration tests instead")
}

func TestUpdateHub_InvalidHubID_BadRequest(t *testing.T) {
	router := setupUpdateHubTest()

	reqBody := models.UpdateHubRequest{
		Name: strPtr("Updated Hub"),
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "PATCH", "/hubs/not-a-uuid", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestUpdateHub_MissingHubID_BadRequest(t *testing.T) {
	router := setupUpdateHubTest()

	reqBody := models.UpdateHubRequest{
		Name: strPtr("Updated Hub"),
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "PATCH", "/hubs//", reqBody)
	// Will be 307 redirect due to trailing slash
	assert.Equal(t, http.StatusTemporaryRedirect, tc.GetResponseStatus())
}

func TestUpdateHub_InvalidJSON_BadRequest(t *testing.T) {
	router := setupUpdateHubTest()

	req := httptest.NewRequest(http.MethodPatch, "/hubs/550e8400-e29b-41d4-a716-446655440000", bytes.NewBufferString(`{invalid json}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateHub_NameTooShort_BadRequest(t *testing.T) {
	router := setupUpdateHubTest()

	reqBody := models.UpdateHubRequest{
		Name: strPtr("AB"),
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "PATCH", "/hubs/550e8400-e29b-41d4-a716-446655440000", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestUpdateHub_NameTooLong_BadRequest(t *testing.T) {
	router := setupUpdateHubTest()

	longName := "a"
	for i := 0; i < 110; i++ {
		longName += "a"
	}

	reqBody := models.UpdateHubRequest{
		Name: &longName,
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "PATCH", "/hubs/550e8400-e29b-41d4-a716-446655440000", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestUpdateHub_AddressTooShort_BadRequest(t *testing.T) {
	router := setupUpdateHubTest()

	reqBody := models.UpdateHubRequest{
		Address: strPtr("123"),
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "PATCH", "/hubs/550e8400-e29b-41d4-a716-446655440000", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestUpdateHub_AddressTooLong_BadRequest(t *testing.T) {
	router := setupUpdateHubTest()

	longAddress := "123 Main Street"
	for i := 0; i < 210; i++ {
		longAddress += "a"
	}

	reqBody := models.UpdateHubRequest{
		Address: &longAddress,
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "PATCH", "/hubs/550e8400-e29b-41d4-a716-446655440000", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestUpdateHub_ContactTooShort_BadRequest(t *testing.T) {
	router := setupUpdateHubTest()

	reqBody := models.UpdateHubRequest{
		Contact: strPtr("abc"),
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "PATCH", "/hubs/550e8400-e29b-41d4-a716-446655440000", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestUpdateHub_ContactTooLong_BadRequest(t *testing.T) {
	router := setupUpdateHubTest()

	longContact := "a"
	for i := 0; i < 60; i++ {
		longContact += "a"
	}

	reqBody := models.UpdateHubRequest{
		Contact: &longContact,
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "PATCH", "/hubs/550e8400-e29b-41d4-a716-446655440000", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestUpdateHub_LongitudeTooLong_BadRequest(t *testing.T) {
	router := setupUpdateHubTest()

	longLongitude := "a"
	for i := 0; i < 60; i++ {
		longLongitude += "a"
	}

	reqBody := models.UpdateHubRequest{
		Longitude: &longLongitude,
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "PATCH", "/hubs/550e8400-e29b-41d4-a716-446655440000", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestUpdateHub_LatitudeTooLong_BadRequest(t *testing.T) {
	router := setupUpdateHubTest()

	longLatitude := "a"
	for i := 0; i < 60; i++ {
		longLatitude += "a"
	}

	reqBody := models.UpdateHubRequest{
		Latitude: &longLatitude,
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "PATCH", "/hubs/550e8400-e29b-41d4-a716-446655440000", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestUpdateHub_AllFieldsValid_RequiresDBSetup(t *testing.T) {
	t.Skip("Requires DB setup - use integration tests instead")
}

// DeleteHub Tests

func setupDeleteHubTest() *gin.Engine {
	router := tests.NewTestRouter()
	router.DELETE("/hubs/:id", api.DeleteHub)
	return router
}

func TestDeleteHub_ValidRequest_RequiresDBSetup(t *testing.T) {
	t.Skip("Requires DB setup - use integration tests instead")
}

func TestDeleteHub_InvalidHubID_BadRequest(t *testing.T) {
	router := setupDeleteHubTest()

	tc := tests.ExecuteRequest(router, "DELETE", "/hubs/not-a-uuid")
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestDeleteHub_MissingHubID_BadRequest(t *testing.T) {
	router := setupDeleteHubTest()

	tc := tests.ExecuteRequest(router, "DELETE", "/hubs/")
	assert.Equal(t, http.StatusNotFound, tc.GetResponseStatus())
}

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}
