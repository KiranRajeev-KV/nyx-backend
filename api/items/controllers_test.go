package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	api "github.com/KiranRajeev-KV/nyx-backend/api/items"
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

// setupFetchItemsTest sets up common test dependencies for fetch items tests
func setupFetchItemsTest() *gin.Engine {
	router := tests.NewTestRouter()
	router.GET("/items", api.FetchItems)
	return router
}

func TestFetchItems_ValidRequest_RequiresDBSetup(t *testing.T) {
	t.Skip("Requires DB setup - use integration tests instead")
}

func TestFetchItems_InvalidTypeParam_BadRequest(t *testing.T) {
	t.Skip("Requires DB setup - validation happens in handler with DB access - use integration tests instead")
}

func TestFetchItems_TypeParamLOST_RequiresDBSetup(t *testing.T) {
	t.Skip("Requires DB setup - use integration tests instead")
}

func TestFetchItems_TypeParamFOUND_RequiresDBSetup(t *testing.T) {
	t.Skip("Requires DB setup - use integration tests instead")
}

// CreateItem Tests

func setupCreateItemTest() *gin.Engine {
	router := tests.NewTestRouterWithAuthMiddleware()
	router.POST("/items", api.CreateItem)
	return router
}

func TestCreateItem_ValidRequest_RequiresDBSetup(t *testing.T) {
	t.Skip("Requires DB setup - use integration tests instead")
}

func TestCreateItem_InvalidJSON_BadRequest(t *testing.T) {
	router := setupCreateItemTest()

	req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewBufferString(`{invalid json}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateItem_MissingName_BadRequest(t *testing.T) {
	router := setupCreateItemTest()

	reqBody := models.CreateItemRequest{
		IsAnonymous: false,
		Description: "This is a detailed description of the item",
		Type:        "LOST",
		Location:    "Central Park",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/items", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateItem_NameTooShort_BadRequest(t *testing.T) {
	router := setupCreateItemTest()

	reqBody := models.CreateItemRequest{
		IsAnonymous: false,
		Name:        "AB",
		Description: "This is a detailed description of the item",
		Type:        "LOST",
		Location:    "Central Park",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/items", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateItem_NameTooLong_BadRequest(t *testing.T) {
	router := setupCreateItemTest()

	longName := "a"
	for i := 0; i < 110; i++ {
		longName += "a"
	}

	reqBody := models.CreateItemRequest{
		IsAnonymous: false,
		Name:        longName,
		Description: "This is a detailed description of the item",
		Type:        "LOST",
		Location:    "Central Park",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/items", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateItem_MissingDescription_BadRequest(t *testing.T) {
	router := setupCreateItemTest()

	reqBody := models.CreateItemRequest{
		IsAnonymous: false,
		Name:        "Lost Item",
		Type:        "LOST",
		Location:    "Central Park",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/items", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateItem_DescriptionTooShort_BadRequest(t *testing.T) {
	router := setupCreateItemTest()

	reqBody := models.CreateItemRequest{
		IsAnonymous: false,
		Name:        "Lost Item",
		Description: "short",
		Type:        "LOST",
		Location:    "Central Park",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/items", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateItem_DescriptionTooLong_BadRequest(t *testing.T) {
	router := setupCreateItemTest()

	longDesc := "a"
	for i := 0; i < 510; i++ {
		longDesc += "a"
	}

	reqBody := models.CreateItemRequest{
		IsAnonymous: false,
		Name:        "Lost Item",
		Description: longDesc,
		Type:        "LOST",
		Location:    "Central Park",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/items", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateItem_MissingType_BadRequest(t *testing.T) {
	router := setupCreateItemTest()

	reqBody := models.CreateItemRequest{
		IsAnonymous: false,
		Name:        "Lost Item",
		Description: "This is a detailed description of the item",
		Location:    "Central Park",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/items", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateItem_InvalidType_BadRequest(t *testing.T) {
	router := setupCreateItemTest()

	reqBody := models.CreateItemRequest{
		IsAnonymous: false,
		Name:        "Lost Item",
		Description: "This is a detailed description of the item",
		Type:        "INVALID",
		Location:    "Central Park",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/items", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateItem_LOSTAnonymous_BadRequest(t *testing.T) {
	router := setupCreateItemTest()

	reqBody := models.CreateItemRequest{
		IsAnonymous: true,
		Name:        "Lost Item",
		Description: "This is a detailed description of the item",
		Type:        "LOST",
		Location:    "Central Park",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/items", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateItem_FOUNDWithoutHubId_BadRequest(t *testing.T) {
	router := setupCreateItemTest()

	reqBody := models.CreateItemRequest{
		IsAnonymous: false,
		Name:        "Found Item",
		Description: "This is a detailed description of the item",
		Type:        "FOUND",
		Location:    "Central Park",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/items", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateItem_FOUNDWithInvalidHubId_BadRequest(t *testing.T) {
	router := setupCreateItemTest()

	hubId := "not-a-uuid"
	reqBody := models.CreateItemRequest{
		IsAnonymous: false,
		Name:        "Found Item",
		Description: "This is a detailed description of the item",
		Type:        "FOUND",
		Location:    "Central Park",
		HubId:       &hubId,
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/items", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateItem_LocationTooShort_BadRequest(t *testing.T) {
	router := setupCreateItemTest()

	reqBody := models.CreateItemRequest{
		IsAnonymous: false,
		Name:        "Lost Item",
		Description: "This is a detailed description of the item",
		Type:        "LOST",
		Location:    "Park",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/items", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateItem_LocationTooLong_BadRequest(t *testing.T) {
	router := setupCreateItemTest()

	longLocation := "a"
	for i := 0; i < 210; i++ {
		longLocation += "a"
	}

	reqBody := models.CreateItemRequest{
		IsAnonymous: false,
		Name:        "Lost Item",
		Description: "This is a detailed description of the item",
		Type:        "LOST",
		Location:    longLocation,
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/items", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestCreateItem_MissingAuthContext_Fatal(t *testing.T) {
	router := setupCreateItemTest()

	hubId := "550e8400-e29b-41d4-a716-446655440000"
	reqBody := models.CreateItemRequest{
		IsAnonymous: false,
		Name:        "Found Item",
		Description: "This is a detailed description of the item",
		Type:        "FOUND",
		Location:    "Central Park",
		HubId:       &hubId,
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/items", reqBody)
	// GrabUserId logs fatal and returns 500
	assert.Equal(t, http.StatusInternalServerError, tc.GetResponseStatus())
}

// FetchItemById Tests

func setupFetchItemByIdTest() *gin.Engine {
	router := tests.NewTestRouter()
	router.GET("/items/:id", api.FetchItemById)
	return router
}

func TestFetchItemById_ValidRequest_RequiresDBSetup(t *testing.T) {
	t.Skip("Requires DB setup - use integration tests instead")
}

func TestFetchItemById_InvalidItemID_BadRequest(t *testing.T) {
	router := setupFetchItemByIdTest()

	tc := tests.ExecuteRequest(router, "GET", "/items/not-a-uuid")
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestFetchItemById_MissingItemID_NotFound(t *testing.T) {
	router := setupFetchItemByIdTest()

	tc := tests.ExecuteRequest(router, "GET", "/items/")
	assert.Equal(t, http.StatusNotFound, tc.GetResponseStatus())
}

// FetchAllItemsByUserId Tests

func setupFetchAllItemsByUserIdTest() *gin.Engine {
	router := tests.NewTestRouterWithAuthMiddleware()
	router.GET("/items/user/me", api.FetchAllItemsByUserId)
	return router
}

func TestFetchAllItemsByUserId_ValidRequest_RequiresDBSetup(t *testing.T) {
	t.Skip("Requires DB setup - use integration tests instead")
}

func TestFetchAllItemsByUserId_MissingAuthContext_Fatal(t *testing.T) {
	router := setupFetchAllItemsByUserIdTest()

	tc := tests.ExecuteAuthenticatedRequest(router, "GET", "/items/user/me", "")
	// GrabUserId logs fatal and returns 500
	assert.Equal(t, http.StatusInternalServerError, tc.GetResponseStatus())
}

// UpdateItemById Tests

func setupUpdateItemByIdTest() *gin.Engine {
	router := tests.NewTestRouterWithAuthMiddleware()
	router.PATCH("/items/:id", api.UpdateItemById)
	return router
}

func TestUpdateItemById_ValidRequest_RequiresDBSetup(t *testing.T) {
	t.Skip("Requires DB setup - use integration tests instead")
}

func TestUpdateItemById_InvalidItemID_BadRequest(t *testing.T) {
	router := setupUpdateItemByIdTest()

	reqBody := models.UpdateItemRequest{
		Name: strPtr("Updated Item"),
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "PATCH", "/items/not-a-uuid", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestUpdateItemById_InvalidJSON_BadRequest(t *testing.T) {
	t.Skip("Requires auth middleware - JSON validation happens after auth - use integration tests instead")
}

func TestUpdateItemById_NameTooShort_BadRequest(t *testing.T) {
	t.Skip("Requires auth middleware - field validation happens after auth - use integration tests instead")
}

func TestUpdateItemById_NameTooLong_BadRequest(t *testing.T) {
	t.Skip("Requires auth middleware - field validation happens after auth - use integration tests instead")
}

func TestUpdateItemById_DescriptionTooShort_BadRequest(t *testing.T) {
	t.Skip("Requires auth middleware - field validation happens after auth - use integration tests instead")
}

func TestUpdateItemById_DescriptionTooLong_BadRequest(t *testing.T) {
	t.Skip("Requires auth middleware - field validation happens after auth - use integration tests instead")
}

func TestUpdateItemById_LocationTooShort_BadRequest(t *testing.T) {
	t.Skip("Requires auth middleware - field validation happens after auth - use integration tests instead")
}

func TestUpdateItemById_LocationTooLong_BadRequest(t *testing.T) {
	t.Skip("Requires auth middleware - field validation happens after auth - use integration tests instead")
}

func TestUpdateItemById_InvalidHubId_BadRequest(t *testing.T) {
	t.Skip("Requires auth middleware - field validation happens after auth - use integration tests instead")
}

func TestUpdateItemById_MissingAuthContext_Fatal(t *testing.T) {
	router := setupUpdateItemByIdTest()

	reqBody := models.UpdateItemRequest{
		Name: strPtr("Updated Item"),
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "PATCH", "/items/550e8400-e29b-41d4-a716-446655440000", reqBody)
	// GrabUserId logs fatal and returns 500
	assert.Equal(t, http.StatusInternalServerError, tc.GetResponseStatus())
}

// DeleteItemById Tests

func setupDeleteItemByIdTest() *gin.Engine {
	router := tests.NewTestRouterWithAuthMiddleware()
	router.DELETE("/items/:id", api.DeleteItemById)
	return router
}

func TestDeleteItemById_ValidRequest_RequiresDBSetup(t *testing.T) {
	t.Skip("Requires DB setup - use integration tests instead")
}

func TestDeleteItemById_InvalidItemID_BadRequest(t *testing.T) {
	router := setupDeleteItemByIdTest()

	tc := tests.ExecuteRequest(router, "DELETE", "/items/not-a-uuid")
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestDeleteItemById_MissingItemID_BadRequest(t *testing.T) {
	router := setupDeleteItemByIdTest()

	tc := tests.ExecuteRequest(router, "DELETE", "/items/")
	assert.Equal(t, http.StatusNotFound, tc.GetResponseStatus())
}

func TestDeleteItemById_MissingAuthContext_Fatal(t *testing.T) {
	router := setupDeleteItemByIdTest()

	tc := tests.ExecuteAuthenticatedRequest(router, "DELETE", "/items/550e8400-e29b-41d4-a716-446655440000", "")
	// GrabUserId logs fatal and returns 500
	assert.Equal(t, http.StatusInternalServerError, tc.GetResponseStatus())
}

// UpdateItemStatus Tests

func setupUpdateItemStatusTest() *gin.Engine {
	router := tests.NewTestRouter()
	router.PATCH("/items/:id/status", api.UpdateItemStatus)
	return router
}

func TestUpdateItemStatus_ValidRequest_RequiresDBSetup(t *testing.T) {
	t.Skip("Requires DB setup - use integration tests instead")
}

func TestUpdateItemStatus_InvalidItemID_BadRequest(t *testing.T) {
	router := setupUpdateItemStatusTest()

	reqBody := models.UpdateItemStatusRequest{
		Status: "OPEN",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "PATCH", "/items/not-a-uuid/status", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestUpdateItemStatus_MissingStatus_BadRequest(t *testing.T) {
	router := setupUpdateItemStatusTest()

	reqBody := models.UpdateItemStatusRequest{}

	tc := tests.ExecuteRequestWithJSONBody(router, "PATCH", "/items/550e8400-e29b-41d4-a716-446655440000/status", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestUpdateItemStatus_InvalidStatus_BadRequest(t *testing.T) {
	router := setupUpdateItemStatusTest()

	reqBody := models.UpdateItemStatusRequest{
		Status: "INVALID",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "PATCH", "/items/550e8400-e29b-41d4-a716-446655440000/status", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestUpdateItemStatus_StatusOPEN_ValidRequest(t *testing.T) {
	t.Skip("Requires DB setup - validation passes but handler accesses DB - use integration tests instead")
}

func TestUpdateItemStatus_StatusPENDING_CLAIM_ValidRequest(t *testing.T) {
	t.Skip("Requires DB setup - validation passes but handler accesses DB - use integration tests instead")
}

func TestUpdateItemStatus_StatusARCHIVED_ValidRequest(t *testing.T) {
	t.Skip("Requires DB setup - validation passes but handler accesses DB - use integration tests instead")
}

func TestUpdateItemStatus_StatusRESOLVED_ValidRequest(t *testing.T) {
	t.Skip("Requires DB setup - validation passes but handler accesses DB - use integration tests instead")
}

func TestUpdateItemStatus_InvalidJSON_BadRequest(t *testing.T) {
	router := setupUpdateItemStatusTest()

	req := httptest.NewRequest(http.MethodPatch, "/items/550e8400-e29b-41d4-a716-446655440000/status", bytes.NewBufferString(`{invalid json}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}
