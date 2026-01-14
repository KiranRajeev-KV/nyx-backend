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
