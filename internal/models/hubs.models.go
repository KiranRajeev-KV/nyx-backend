package models

import v "github.com/go-ozzo/ozzo-validation/v4"

type CreateHubRequest struct {
	Name      string `json:"name"`
	Address   string `json:"address"`
	Contact   string `json:"contact"`
	Longitude string `json:"longitude"`
	Latitude  string `json:"latitude"`
}

func (r CreateHubRequest) Validate() (errorMsg string, err error) {
	err = v.ValidateStruct(&r,
		v.Field(&r.Name, v.Required, v.Length(3, 100)),
		v.Field(&r.Address, v.Required, v.Length(5, 200)),
		v.Field(&r.Contact, v.Required, v.Length(5, 50)),
		v.Field(&r.Longitude, v.Length(1, 50)),
		v.Field(&r.Latitude, v.Length(1, 50)),
	)
	return "Invalid request format for creating a hub", err
}

type UpdateHubRequest struct {
	Name      *string `json:"name"`
	Address   *string `json:"address"`
	Contact   *string `json:"contact"`
	Longitude *string `json:"longitude"`
	Latitude  *string `json:"latitude"`
}

func (r UpdateHubRequest) Validate() (errorMsg string, err error) {
	err = v.ValidateStruct(&r,
		v.Field(&r.Name, v.When(r.Name != nil, v.Required, v.Length(3, 100))),
		v.Field(&r.Address, v.When(r.Address != nil, v.Required, v.Length(5, 200))),
		v.Field(&r.Contact, v.When(r.Contact != nil, v.Required, v.Length(5, 50))),
		v.Field(&r.Longitude, v.When(r.Longitude != nil, v.Length(1, 50))),
		v.Field(&r.Latitude, v.When(r.Latitude != nil, v.Length(1, 50))),
	)
	return "Invalid request format for updating a hub", err
}
