package models_test

import (
	"testing"

	"github.com/KiranRajeev-KV/nyx-backend/internal/models"
	"github.com/stretchr/testify/assert"
)

// ==================== CreateItemRequest Tests ====================

func TestCreateItemRequest_ValidLostItem_NoError(t *testing.T) {
	req := models.CreateItemRequest{
		IsAnonymous: false,
		Name:        "Blue Backpack",
		Description: "A blue backpack with laptop compartment",
		Type:        "LOST",
		Location:    "Library second floor",
	}
	_, err := req.Validate()
	assert.NoError(t, err)
}

func TestCreateItemRequest_ValidFoundItem_NoError(t *testing.T) {
	hubId := "550e8400-e29b-41d4-a716-446655440000"
	req := models.CreateItemRequest{
		IsAnonymous: true,
		HubId:       &hubId,
		Name:        "Black Wallet",
		Description: "Leather wallet found near cafeteria",
		Type:        "FOUND",
		Location:    "Cafeteria entrance",
	}
	_, err := req.Validate()
	assert.NoError(t, err)
}

func TestCreateItemRequest_MissingName_ReturnsError(t *testing.T) {
	req := models.CreateItemRequest{
		Description: "A blue backpack with laptop compartment",
		Type:        "LOST",
	}
	msg, err := req.Validate()
	assert.Error(t, err)
	assert.Equal(t, "Invalid request format for creating an item", msg)
}

func TestCreateItemRequest_NameTooShort_ReturnsError(t *testing.T) {
	req := models.CreateItemRequest{
		Name:        "AB",
		Description: "A blue backpack with laptop compartment",
		Type:        "LOST",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateItemRequest_NameTooLong_ReturnsError(t *testing.T) {
	longName := ""
	for i := 0; i < 101; i++ {
		longName += "a"
	}
	req := models.CreateItemRequest{
		Name:        longName,
		Description: "A blue backpack with laptop compartment",
		Type:        "LOST",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateItemRequest_MissingDescription_ReturnsError(t *testing.T) {
	req := models.CreateItemRequest{
		Name: "Blue Backpack",
		Type: "LOST",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateItemRequest_DescriptionTooShort_ReturnsError(t *testing.T) {
	req := models.CreateItemRequest{
		Name:        "Blue Backpack",
		Description: "Short",
		Type:        "LOST",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateItemRequest_DescriptionTooLong_ReturnsError(t *testing.T) {
	longDesc := ""
	for i := 0; i < 501; i++ {
		longDesc += "a"
	}
	req := models.CreateItemRequest{
		Name:        "Blue Backpack",
		Description: longDesc,
		Type:        "LOST",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateItemRequest_MissingType_ReturnsError(t *testing.T) {
	req := models.CreateItemRequest{
		Name:        "Blue Backpack",
		Description: "A blue backpack with laptop compartment",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateItemRequest_InvalidType_ReturnsError(t *testing.T) {
	req := models.CreateItemRequest{
		Name:        "Blue Backpack",
		Description: "A blue backpack with laptop compartment",
		Type:        "STOLEN",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateItemRequest_LostItemAnonymous_ReturnsError(t *testing.T) {
	req := models.CreateItemRequest{
		IsAnonymous: true,
		Name:        "Blue Backpack",
		Description: "A blue backpack with laptop compartment",
		Type:        "LOST",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateItemRequest_FoundItemMissingHubId_ReturnsError(t *testing.T) {
	req := models.CreateItemRequest{
		Name:        "Black Wallet",
		Description: "Leather wallet found near cafeteria",
		Type:        "FOUND",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateItemRequest_FoundItemInvalidHubId_ReturnsError(t *testing.T) {
	hubId := "not-a-uuid"
	req := models.CreateItemRequest{
		HubId:       &hubId,
		Name:        "Black Wallet",
		Description: "Leather wallet found near cafeteria",
		Type:        "FOUND",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateItemRequest_LocationTooShort_ReturnsError(t *testing.T) {
	req := models.CreateItemRequest{
		Name:        "Blue Backpack",
		Description: "A blue backpack with laptop compartment",
		Type:        "LOST",
		Location:    "Lib",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateItemRequest_LocationTooLong_ReturnsError(t *testing.T) {
	longLocation := ""
	for i := 0; i < 201; i++ {
		longLocation += "a"
	}
	req := models.CreateItemRequest{
		Name:        "Blue Backpack",
		Description: "A blue backpack with laptop compartment",
		Type:        "LOST",
		Location:    longLocation,
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

// ==================== UpdateItemRequest Tests ====================

func TestUpdateItemRequest_ValidPartialUpdate_NoError(t *testing.T) {
	name := "Updated Backpack"
	req := models.UpdateItemRequest{
		Name: &name,
	}
	_, err := req.Validate()
	assert.NoError(t, err)
}

func TestUpdateItemRequest_ValidFullUpdate_NoError(t *testing.T) {
	name := "Updated Backpack"
	desc := "Updated description for the item"
	location := "New location here"
	hubId := "550e8400-e29b-41d4-a716-446655440000"
	req := models.UpdateItemRequest{
		Name:        &name,
		Description: &desc,
		Location:    &location,
		HubId:       &hubId,
	}
	_, err := req.Validate()
	assert.NoError(t, err)
}

func TestUpdateItemRequest_EmptyRequest_NoError(t *testing.T) {
	req := models.UpdateItemRequest{}
	_, err := req.Validate()
	assert.NoError(t, err)
}

func TestUpdateItemRequest_NameTooShort_ReturnsError(t *testing.T) {
	name := "AB"
	req := models.UpdateItemRequest{
		Name: &name,
	}
	msg, err := req.Validate()
	assert.Error(t, err)
	assert.Equal(t, "Invalid request format for updating an item", msg)
}

func TestUpdateItemRequest_NameTooLong_ReturnsError(t *testing.T) {
	longName := ""
	for i := 0; i < 101; i++ {
		longName += "a"
	}
	req := models.UpdateItemRequest{
		Name: &longName,
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestUpdateItemRequest_DescriptionTooShort_ReturnsError(t *testing.T) {
	desc := "Short"
	req := models.UpdateItemRequest{
		Description: &desc,
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestUpdateItemRequest_DescriptionTooLong_ReturnsError(t *testing.T) {
	longDesc := ""
	for i := 0; i < 501; i++ {
		longDesc += "a"
	}
	req := models.UpdateItemRequest{
		Description: &longDesc,
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestUpdateItemRequest_LocationTooShort_ReturnsError(t *testing.T) {
	location := "Lib"
	req := models.UpdateItemRequest{
		Location: &location,
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestUpdateItemRequest_LocationTooLong_ReturnsError(t *testing.T) {
	longLocation := ""
	for i := 0; i < 201; i++ {
		longLocation += "a"
	}
	req := models.UpdateItemRequest{
		Location: &longLocation,
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestUpdateItemRequest_InvalidHubId_ReturnsError(t *testing.T) {
	hubId := "not-a-uuid"
	req := models.UpdateItemRequest{
		HubId: &hubId,
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

// ==================== UpdateItemStatusRequest Tests ====================

func TestUpdateItemStatusRequest_ValidOpen_NoError(t *testing.T) {
	req := models.UpdateItemStatusRequest{
		Status: "OPEN",
	}
	_, err := req.Validate()
	assert.NoError(t, err)
}

func TestUpdateItemStatusRequest_ValidPendingClaim_NoError(t *testing.T) {
	req := models.UpdateItemStatusRequest{
		Status: "PENDING_CLAIM",
	}
	_, err := req.Validate()
	assert.NoError(t, err)
}

func TestUpdateItemStatusRequest_ValidArchived_NoError(t *testing.T) {
	req := models.UpdateItemStatusRequest{
		Status: "ARCHIVED",
	}
	_, err := req.Validate()
	assert.NoError(t, err)
}

func TestUpdateItemStatusRequest_ValidResolved_NoError(t *testing.T) {
	req := models.UpdateItemStatusRequest{
		Status: "RESOLVED",
	}
	_, err := req.Validate()
	assert.NoError(t, err)
}

func TestUpdateItemStatusRequest_MissingStatus_ReturnsError(t *testing.T) {
	req := models.UpdateItemStatusRequest{}
	msg, err := req.Validate()
	assert.Error(t, err)
	assert.Equal(t, "Invalid request format for updating item status", msg)
}

func TestUpdateItemStatusRequest_InvalidStatus_ReturnsError(t *testing.T) {
	req := models.UpdateItemStatusRequest{
		Status: "CLOSED",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}
