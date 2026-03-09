package models

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type CreateClaimRequest struct {
	FoundItemID   string  `json:"found_item_id"`
	LostItemID    string  `json:"lost_item_id"`
	ProofText     string  `json:"proof_text"`
	ProofImageUrl *string `json:"proof_image_url"`
}

func (r CreateClaimRequest) Validate() (errorMsg string, err error) {
	err = v.ValidateStruct(&r,
		v.Field(&r.FoundItemID, v.Required.Error("Found item ID is required"), is.UUID.Error("Found item ID must be a valid UUID")),
		v.Field(&r.LostItemID, v.Required.Error("Lost item ID is required"), is.UUID.Error("Lost item ID must be a valid UUID")),
		v.Field(&r.ProofText, v.Required.Error("Proof text is required"), v.Length(10, 1000).Error("Proof text must be between 10 and 1000 characters")),
		v.Field(&r.ProofImageUrl, v.Length(0, 500).Error("Proof image URL must be less than 500 characters")),
	)
	return "Invalid request format for creating a claim", err
}

type ProcessClaimRequest struct {
	Status     string `json:"status"`
	AdminNotes string `json:"admin_notes"`
}

func (r ProcessClaimRequest) Validate() (errorMsg string, err error) {
	err = v.ValidateStruct(&r,
		v.Field(&r.Status, v.Required.Error("Status is required"), v.In("APPROVED", "REJECTED").Error("Status must be either APPROVED or REJECTED")),
		v.Field(&r.AdminNotes, v.Required.Error("Admin notes are required"), v.Length(5, 500).Error("Admin notes must be between 5 and 500 characters")),
	)
	return "Invalid request format for processing a claim", err
}

type UploadClaimProofImageRequest struct {
	ContentType string `json:"content_type"`
}

func (r UploadClaimProofImageRequest) Validate() (errorMsg string, err error) {
	err = v.ValidateStruct(&r,
		v.Field(&r.ContentType, v.Required.Error("Content type is required"), v.In("image/jpeg", "image/png", "image/webp").Error("Only JPEG, PNG, and WebP images are supported")),
	)
	return "Invalid request format for uploading claim proof image", err
}

type UploadClaimProofImageResponse struct {
	PresignedUrl string `json:"presigned_url"`
	ObjectKey    string `json:"object_key"`
}
