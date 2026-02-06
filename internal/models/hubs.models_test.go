package models_test

import (
	"testing"

	"github.com/KiranRajeev-KV/nyx-backend/internal/models"
	"github.com/stretchr/testify/assert"
)

// ==================== CreateHubRequest Tests ====================

func TestCreateHubRequest_Valid_NoError(t *testing.T) {
	req := models.CreateHubRequest{
		Name:    "Main Library Hub",
		Address: "123 University Ave, Building A",
		Contact: "library@university.edu",
	}
	_, err := req.Validate()
	assert.NoError(t, err)
}

func TestCreateHubRequest_ValidWithCoordinates_NoError(t *testing.T) {
	req := models.CreateHubRequest{
		Name:      "Main Library Hub",
		Address:   "123 University Ave, Building A",
		Contact:   "library@university.edu",
		Longitude: "-122.4194",
		Latitude:  "37.7749",
	}
	_, err := req.Validate()
	assert.NoError(t, err)
}

func TestCreateHubRequest_MissingName_ReturnsError(t *testing.T) {
	req := models.CreateHubRequest{
		Address: "123 University Ave, Building A",
		Contact: "library@university.edu",
	}
	msg, err := req.Validate()
	assert.Error(t, err)
	assert.Equal(t, "Invalid request format for creating a hub", msg)
}

func TestCreateHubRequest_NameTooShort_ReturnsError(t *testing.T) {
	req := models.CreateHubRequest{
		Name:    "AB",
		Address: "123 University Ave, Building A",
		Contact: "library@university.edu",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateHubRequest_NameTooLong_ReturnsError(t *testing.T) {
	longName := ""
	for i := 0; i < 101; i++ {
		longName += "a"
	}
	req := models.CreateHubRequest{
		Name:    longName,
		Address: "123 University Ave, Building A",
		Contact: "library@university.edu",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateHubRequest_MissingAddress_ReturnsError(t *testing.T) {
	req := models.CreateHubRequest{
		Name:    "Main Library Hub",
		Contact: "library@university.edu",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateHubRequest_AddressTooShort_ReturnsError(t *testing.T) {
	req := models.CreateHubRequest{
		Name:    "Main Library Hub",
		Address: "123",
		Contact: "library@university.edu",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateHubRequest_AddressTooLong_ReturnsError(t *testing.T) {
	longAddress := ""
	for i := 0; i < 201; i++ {
		longAddress += "a"
	}
	req := models.CreateHubRequest{
		Name:    "Main Library Hub",
		Address: longAddress,
		Contact: "library@university.edu",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateHubRequest_MissingContact_ReturnsError(t *testing.T) {
	req := models.CreateHubRequest{
		Name:    "Main Library Hub",
		Address: "123 University Ave, Building A",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateHubRequest_ContactTooShort_ReturnsError(t *testing.T) {
	req := models.CreateHubRequest{
		Name:    "Main Library Hub",
		Address: "123 University Ave, Building A",
		Contact: "ab",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateHubRequest_ContactTooLong_ReturnsError(t *testing.T) {
	longContact := ""
	for i := 0; i < 51; i++ {
		longContact += "a"
	}
	req := models.CreateHubRequest{
		Name:    "Main Library Hub",
		Address: "123 University Ave, Building A",
		Contact: longContact,
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateHubRequest_LongitudeTooLong_ReturnsError(t *testing.T) {
	longLongitude := ""
	for i := 0; i < 51; i++ {
		longLongitude += "1"
	}
	req := models.CreateHubRequest{
		Name:      "Main Library Hub",
		Address:   "123 University Ave, Building A",
		Contact:   "library@university.edu",
		Longitude: longLongitude,
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateHubRequest_LatitudeTooLong_ReturnsError(t *testing.T) {
	longLatitude := ""
	for i := 0; i < 51; i++ {
		longLatitude += "1"
	}
	req := models.CreateHubRequest{
		Name:     "Main Library Hub",
		Address:  "123 University Ave, Building A",
		Contact:  "library@university.edu",
		Latitude: longLatitude,
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

// ==================== UpdateHubRequest Tests ====================

func TestUpdateHubRequest_ValidPartialUpdate_NoError(t *testing.T) {
	name := "Updated Hub Name"
	req := models.UpdateHubRequest{
		Name: &name,
	}
	_, err := req.Validate()
	assert.NoError(t, err)
}

func TestUpdateHubRequest_ValidFullUpdate_NoError(t *testing.T) {
	name := "Updated Hub Name"
	address := "456 New Address St"
	contact := "newemail@university.edu"
	longitude := "-122.4194"
	latitude := "37.7749"
	req := models.UpdateHubRequest{
		Name:      &name,
		Address:   &address,
		Contact:   &contact,
		Longitude: &longitude,
		Latitude:  &latitude,
	}
	_, err := req.Validate()
	assert.NoError(t, err)
}

func TestUpdateHubRequest_EmptyRequest_NoError(t *testing.T) {
	req := models.UpdateHubRequest{}
	_, err := req.Validate()
	assert.NoError(t, err)
}

func TestUpdateHubRequest_NameTooShort_ReturnsError(t *testing.T) {
	name := "AB"
	req := models.UpdateHubRequest{
		Name: &name,
	}
	msg, err := req.Validate()
	assert.Error(t, err)
	assert.Equal(t, "Invalid request format for updating a hub", msg)
}

func TestUpdateHubRequest_NameTooLong_ReturnsError(t *testing.T) {
	longName := ""
	for i := 0; i < 101; i++ {
		longName += "a"
	}
	req := models.UpdateHubRequest{
		Name: &longName,
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestUpdateHubRequest_AddressTooShort_ReturnsError(t *testing.T) {
	address := "123"
	req := models.UpdateHubRequest{
		Address: &address,
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestUpdateHubRequest_AddressTooLong_ReturnsError(t *testing.T) {
	longAddress := ""
	for i := 0; i < 201; i++ {
		longAddress += "a"
	}
	req := models.UpdateHubRequest{
		Address: &longAddress,
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestUpdateHubRequest_ContactTooShort_ReturnsError(t *testing.T) {
	contact := "ab"
	req := models.UpdateHubRequest{
		Contact: &contact,
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestUpdateHubRequest_ContactTooLong_ReturnsError(t *testing.T) {
	longContact := ""
	for i := 0; i < 51; i++ {
		longContact += "a"
	}
	req := models.UpdateHubRequest{
		Contact: &longContact,
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestUpdateHubRequest_LongitudeTooLong_ReturnsError(t *testing.T) {
	longLongitude := ""
	for i := 0; i < 51; i++ {
		longLongitude += "1"
	}
	req := models.UpdateHubRequest{
		Longitude: &longLongitude,
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestUpdateHubRequest_LatitudeTooLong_ReturnsError(t *testing.T) {
	longLatitude := ""
	for i := 0; i < 51; i++ {
		longLatitude += "1"
	}
	req := models.UpdateHubRequest{
		Latitude: &longLatitude,
	}
	_, err := req.Validate()
	assert.Error(t, err)
}
