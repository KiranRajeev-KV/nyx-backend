package models_test

import (
	"testing"

	"github.com/KiranRajeev-KV/nyx-backend/internal/models"
	"github.com/stretchr/testify/assert"
)

// ==================== CreateClaimRequest Tests ====================

func TestCreateClaimRequest_Valid_NoError(t *testing.T) {
	req := models.CreateClaimRequest{
		FoundItemID: "550e8400-e29b-41d4-a716-446655440000",
		ProofText:   "I lost this item at the library on Monday afternoon.",
	}
	_, err := req.Validate()
	assert.NoError(t, err)
}

func TestCreateClaimRequest_ValidWithImageUrl_NoError(t *testing.T) {
	imageUrl := "https://example.com/proof.jpg"
	req := models.CreateClaimRequest{
		FoundItemID:   "550e8400-e29b-41d4-a716-446655440000",
		ProofText:     "I lost this item at the library on Monday afternoon.",
		ProofImageUrl: &imageUrl,
	}
	_, err := req.Validate()
	assert.NoError(t, err)
}

func TestCreateClaimRequest_MissingFoundItemID_ReturnsError(t *testing.T) {
	req := models.CreateClaimRequest{
		ProofText: "I lost this item at the library on Monday afternoon.",
	}
	msg, err := req.Validate()
	assert.Error(t, err)
	assert.Equal(t, "Invalid request format for creating a claim", msg)
}

func TestCreateClaimRequest_InvalidFoundItemID_ReturnsError(t *testing.T) {
	req := models.CreateClaimRequest{
		FoundItemID: "not-a-uuid",
		ProofText:   "I lost this item at the library on Monday afternoon.",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateClaimRequest_MissingProofText_ReturnsError(t *testing.T) {
	req := models.CreateClaimRequest{
		FoundItemID: "550e8400-e29b-41d4-a716-446655440000",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateClaimRequest_ProofTextTooShort_ReturnsError(t *testing.T) {
	req := models.CreateClaimRequest{
		FoundItemID: "550e8400-e29b-41d4-a716-446655440000",
		ProofText:   "Too short",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateClaimRequest_ProofTextTooLong_ReturnsError(t *testing.T) {
	longText := ""
	for i := 0; i < 1001; i++ {
		longText += "a"
	}
	req := models.CreateClaimRequest{
		FoundItemID: "550e8400-e29b-41d4-a716-446655440000",
		ProofText:   longText,
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestCreateClaimRequest_ImageUrlTooLong_ReturnsError(t *testing.T) {
	longUrl := ""
	for i := 0; i < 501; i++ {
		longUrl += "a"
	}
	req := models.CreateClaimRequest{
		FoundItemID:   "550e8400-e29b-41d4-a716-446655440000",
		ProofText:     "I lost this item at the library on Monday afternoon.",
		ProofImageUrl: &longUrl,
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

// ==================== ProcessClaimRequest Tests ====================

func TestProcessClaimRequest_ValidApproved_NoError(t *testing.T) {
	req := models.ProcessClaimRequest{
		Status:     "APPROVED",
		AdminNotes: "Verified with student ID",
	}
	_, err := req.Validate()
	assert.NoError(t, err)
}

func TestProcessClaimRequest_ValidRejected_NoError(t *testing.T) {
	req := models.ProcessClaimRequest{
		Status:     "REJECTED",
		AdminNotes: "Description does not match the item",
	}
	_, err := req.Validate()
	assert.NoError(t, err)
}

func TestProcessClaimRequest_MissingStatus_ReturnsError(t *testing.T) {
	req := models.ProcessClaimRequest{
		AdminNotes: "Some notes here",
	}
	msg, err := req.Validate()
	assert.Error(t, err)
	assert.Equal(t, "Invalid request format for processing a claim", msg)
}

func TestProcessClaimRequest_InvalidStatus_ReturnsError(t *testing.T) {
	req := models.ProcessClaimRequest{
		Status:     "PENDING",
		AdminNotes: "Some notes here",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestProcessClaimRequest_MissingAdminNotes_ReturnsError(t *testing.T) {
	req := models.ProcessClaimRequest{
		Status: "APPROVED",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestProcessClaimRequest_AdminNotesTooShort_ReturnsError(t *testing.T) {
	req := models.ProcessClaimRequest{
		Status:     "APPROVED",
		AdminNotes: "Hi",
	}
	_, err := req.Validate()
	assert.Error(t, err)
}

func TestProcessClaimRequest_AdminNotesTooLong_ReturnsError(t *testing.T) {
	longNotes := ""
	for i := 0; i < 501; i++ {
		longNotes += "a"
	}
	req := models.ProcessClaimRequest{
		Status:     "APPROVED",
		AdminNotes: longNotes,
	}
	_, err := req.Validate()
	assert.Error(t, err)
}
