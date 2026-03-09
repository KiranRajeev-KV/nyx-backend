package models

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type CreateItemRequest struct {
	IsAnonymous bool    `json:"is_anonymous"`
	HubId       *string `json:"hub_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
	Location    string  `json:"location_description"`
	TimeAt      string  `json:"time_at"`
	Latitude    string  `json:"latitude"`
	Longitude   string  `json:"longitude"`
}

func (r CreateItemRequest) Validate() (errorMsg string, err error) {
	err = v.ValidateStruct(&r,
		v.Field(&r.IsAnonymous, v.When(r.Type == "LOST", v.In(false).Error("LOST items cannot be requested anonymously"))),
		v.Field(&r.HubId, v.When(r.Type == "FOUND", v.Required.Error("Hub ID is required for FOUND items"), is.UUID.Error("Hub ID must be a valid UUID"))),
		v.Field(&r.Name, v.Required.Error("Name is required"), v.Length(3, 100).Error("Name must be between 3 and 100 characters")),
		v.Field(&r.Description, v.Required.Error("Description is required"), v.Length(10, 500).Error("Description must be between 10 and 500 characters")),
		v.Field(&r.Type, v.Required.Error("Type is required"), v.In("LOST", "FOUND").Error("Type must be either LOST or FOUND")),
		v.Field(&r.Location, v.Length(5, 200).Error("Last known location/Found location must be between 5 and 200 characters")),
		v.Field(&r.TimeAt),
		v.Field(&r.Latitude),
		v.Field(&r.Longitude),
	)
	return "Invalid request format for creating an item", err
}

type UpdateItemRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Location    *string `json:"location_description"`
	TimeAt      *string `json:"time_at"`
	Latitude    *string `json:"latitude"`
	Longitude   *string `json:"longitude"`
	HubId       *string `json:"hub_id"`
}

func (r UpdateItemRequest) Validate() (errorMsg string, err error) {
	err = v.ValidateStruct(&r,
		v.Field(&r.Name, v.When(r.Name != nil, v.Length(3, 100).Error("Name must be between 3 and 100 characters"))),
		v.Field(&r.Description, v.When(r.Description != nil, v.Length(10, 500).Error("Description must be between 10 and 500 characters"))),
		v.Field(&r.Location, v.When(r.Location != nil, v.Length(5, 200).Error("Location must be between 5 and 200 characters"))),
		v.Field(&r.HubId, v.When(r.HubId != nil, is.UUID.Error("Hub ID must be a valid UUID"))),
		v.Field(&r.TimeAt),
		v.Field(&r.Latitude),
		v.Field(&r.Longitude),
	)
	return "Invalid request format for updating an item", err
}

type UpdateItemStatusRequest struct {
	Status string `json:"status"`
}

func (r UpdateItemStatusRequest) Validate() (errorMsg string, err error) {
	err = v.ValidateStruct(&r,
		v.Field(&r.Status, v.Required.Error("Status is required"), v.In("OPEN", "PENDING_CLAIM", "ARCHIVED", "RESOLVED").Error("Status must be one of OPEN, PENDING_CLAIM, ARCHIVED, or RESOLVED")),
	)
	return "Invalid request format for updating item status", err
}

type UploadItemImageRequest struct {
	ContentType string `json:"content_type"`
}

func (r UploadItemImageRequest) Validate() (errorMsg string, err error) {
	err = v.ValidateStruct(&r,
		v.Field(&r.ContentType, v.Required.Error("Content-Type is required"), v.In("image/jpeg", "image/png", "image/webp").Error("Only JPEG, PNG, and WebP images are supported")),
	)
	return "Invalid request format for uploading item image", err
}

// UploadItemImageResponse is not validated since it is an outgoing response
type UploadItemImageResponse struct {
	PresignedUrl string `json:"presigned_url"`
	ObjectKey    string `json:"object_key"`
}

type SearchItemsRequest struct {
	Query string `json:"q"`
	Type  string `json:"type"`
}

func (r SearchItemsRequest) Validate() (errorMsg string, err error) {
	err = v.ValidateStruct(&r,
		v.Field(&r.Query, v.Required.Error("Search query is required"), v.Length(1, 200).Error("Search query must be between 1 and 200 characters")),
		v.Field(&r.Type, v.When(r.Type != "", v.In("LOST", "FOUND").Error("Type must be either LOST or FOUND"))),
	)
	return "Invalid search request", err
}
