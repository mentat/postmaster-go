package postmaster

import (
	"errors"
	"fmt"
	"strconv"
)

// Shipment is a base object used in Shipment API requests.
type Shipment struct {
	p            *Postmaster `dontMap:"true"`
	Id           int         `dontMap:"true"`
	To           Address
	From         Address
	Package      Package
	Carrier      string
	Service      string
	Status       string   `dontMap:"true"`
	Tracking     []string `dontMap:"true"`
	PackageCount int      `json:"package_count"`
	CreatedAt    int      `json:"created_at"`
	Cost         int      `dontMap:"true"`
	Prepaid      bool     `dontMap:"true"`
}

// ShipmentList is returned when asking for list of shipments.
type ShipmentList struct {
	Results        []Shipment
	Cursor         string
	PreviousCursor string `json:"previous_cursor"`
}

// Package (not to be confused with packages in fitting API, which are called "Boxes")
// is being used in Shipment request.
type Package struct {
	Id             int `dontMap:"true"`
	Name           string
	Width          float32
	Height         float32
	Length         float32
	Weight         float32
	Customs        Custom
	DimensionUnits string `dontMap:"true" json:"dimension_units"`
	WeightUnits    string `dontMap:"true" json:"weight_units"`
	Type           string `dontMap:"true"`
	LabelUrl       string `dontMap:"true" json:"label_url"`
}

// CustomContent is being used as a single item in Custom object.
type CustomContent struct {
	Description     string
	Quantity        string
	Value           string
	Weight          float32
	WeightUnits     string `json:"weight_units"`
	HSTariffNumber  string `json:"hs_tariff_number"`
	CountryOfOrigin string `json:"country_of_origin"`
}

// Custom is being used per Package. It is necessary only in international
// packages.
type Custom struct {
	Type          string
	Comments      string
	InvoiceNumber string `json:"invoice_number"`
	Contents      CustomContent
}

// Shipment creates a brand new Shipment structure. Don't use new(postmaster.Shipment),
// use this function instead.
func (p *Postmaster) Shipment() (s *Shipment) {
	s = new(Shipment)
	s.p = p
	s.Id = -1 // default for "null" Shipment
	return
}

// Create creates new Shipment in API.
// You musn't invoke this function from an existing Shipment (i.e. shipment.Id > -1).
func (s *Shipment) Create() (*Shipment, error) {
	if s.Id != -1 {
		return nil, errors.New("You can't create an existing shipment.")
	}
	params := mapStruct(s)
	_, err := post(s.p, "v1", "shipments", params, s)
	return s, err
}

// Get fetches single Shipment from API, and replaces existing Shipment structure.
// You musn't invoke this function from an "empty" Shipment (i.e. shipment.Id == -1).
func (s *Shipment) Get() (*Shipment, error) {
	if s.Id == -1 {
		return nil, errors.New("You must provide a shipment ID.")
	}
	endpoint := fmt.Sprintf("shipments/%d", s.Id)
	_, err := get(s.p, "v1", endpoint, nil, s)
	return s, err
}

// Void sets Shipment's status to "voided".
// You musn't invoke this function from an "empty" Shipment (i.e. shipment.Id == -1).
func (s *Shipment) Void() (bool, error) {
	if s.Id == -1 {
		return false, errors.New("You must provide a shipment ID.")
	}
	endpoint := fmt.Sprintf("shipments/%d/void", s.Id)
	var res map[string]string
	_, err := del(s.p, "v1", endpoint, nil, &res)
	if res["message"] == "OK" {
		s.Status = "Voided"
	}
	return res["message"] == "OK", err
}

// Track returns TrackingResponse for Shipment.
// You musn't invoke this function from an "empty" Shipment (i.e. shipment.Id == -1).
// In order to track shipment just by its tracking number, use Postmaster.TrackRef()
// function.
func (s *Shipment) Track() (*TrackingResponse, error) {
	if s.Id == -1 {
		return nil, errors.New("You must provide a shipment ID.")
	}
	endpoint := fmt.Sprintf("shipments/%d/track", s.Id)
	res := TrackingResponse{}
	_, err := get(s.p, "v1", endpoint, nil, &res)
	return &res, err
}

// ListShipments returns a list of shipments, with limit, status and cursor (e.g. for pagination).
func (p *Postmaster) ListShipments(limit int, cursor string, status string) (*ShipmentList, error) {
	params := make(map[string]string)
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	if cursor != "" {
		params["cursor"] = cursor
	}
	if status != "" {
		params["status"] = status
	}
	res := new(ShipmentList)
	_, err := get(p, "v1", "shipments", params, &res)
	// Set Postmaster "base" object for each shipment, so we can use API with them
	for k, _ := range res.Results {
		res.Results[k].p = p
	}
	return res, err
}

// FindShipments returns a list of shipments matching given search query, with limit,
// status and cursor (e.g. for pagination).
func (p *Postmaster) FindShipments(q string, limit int, cursor string) (*ShipmentList, error) {
	params := make(map[string]string)
	if q == "" {
		return nil, errors.New("You must provide search query.")
	}
	params["q"] = q
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	if cursor != "" {
		params["cursor"] = cursor
	}
	res := new(ShipmentList)
	_, err := get(p, "v1", "shipments/search", params, &res)
	// Set Postmaster "base" object for each shipment, so we can use API with them
	for k, _ := range res.Results {
		res.Results[k].p = p
	}
	return res, err
}
