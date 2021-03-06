package gonaturalist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Location struct {
	Longitude float64
	Latitude  float64
}

type Rectangle struct {
	Southwest Location
	Northeast Location
}

type SimpleObservation struct {
	Id                       int64     `json:"id"`
	UserLogin                string    `json:"user_login"`
	PlaceGuess               string    `json:"place_guess"`
	SpeciesGuess             string    `json:"species_guess"`
	Latitude                 float64   `json:"latitude,string"`
	Longitude                float64   `json:"longitude,string"`
	CreatedAt                time.Time `json:"created_at_utc"`
	ObservedOn               string    `json:"observed_on"`
	ObservedOnString         string    `json:"observed_on_string"`
	UpdatedAt                time.Time `json:"updated_at_utc"`
	TaxonId                  int32     `json:"taxon_id"`
	UserId                   int64     `json:"user_id"`
	SiteId                   int64     `json:"site_id"`
	TimeZone                 string    `json:"time_zone"`
	Description              string    `json:"description"`
	Uri                      string    `json:"uri"`
	Uuid                     string    `json:"uuid"`
	TimeObservedAtUtc        time.Time `json:"time_observed_at_utc"`
	PositionalAccuracy       int32     `json:"positional_accuracy"`
	PublicPositionalAccuracy int32     `json:"public_positional_accuracy"`
}

type ObservationsPage struct {
	Paging       *PageHeaders
	Observations []*SimpleObservation
}

type ProjectObservation struct {
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
	ObservationId           int64     `json:"observation_id"`
	Id                      int64     `json:"id"`
	TrackingCode            string    `json:"tracking_code"`
	CuratorIdentificationId int64     `json:"curator_identification_id"`
}

type SimpleUser struct {
	Name  string `json:"name"`
	Id    int64  `json:"id"`
	Login string `json:"login"`
}

type Comment struct {
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Id        int64     `json:"id"`
	ParentId  int64     `json:"parent_id"`
	UserId    int64     `json:"user_id"`
	User      SimpleUser
}

type SimplePhoto struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Id        int64     `json:"id"`
	LargeUrl  string    `json:"large_url"`
	MediumUrl string    `json:"medium_url"`
	SmallUrl  string    `json:"small_url"`
	SquareUrl string    `json:"square_url"`
}

type ObservationPhoto struct {
	Id            int64     `json:"id"`
	PhotoId       int64     `json:"photo_id"`
	ObservationId int64     `json:"observation_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Photo         SimplePhoto
}

type FullObservation struct {
	Id               int64                 `json:"id"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
	Longitude        string                `json:"longitude"`
	Latitude         string                `json:"latitude"`
	ObservedOnString string                `json:"observed_on_string"`
	Photos           []*ObservationPhoto   `json:"observation_photos"`
	Comments         []*Comment            `json:"comments"`
	Projects         []*ProjectObservation `json:"project_observations"`
}

type GetObservationsOpt struct {
	PerPage        *int
	Page           *int
	Rectangle      *Rectangle
	On             *time.Time
	UpdatedSince   *time.Time
	OrderBy        *string
	OrderAscending *bool
	HasGeo         *bool
}

func (o *SimpleObservation) TryParseObservedOn() (time.Time, error) {
	return TryParseObservedOn(o.ObservedOnString)
}

func (c *Client) GetObservations(opt *GetObservationsOpt) (*ObservationsPage, error) {
	var result []*SimpleObservation

	u := c.buildUrl("/observations.json")
	if opt != nil {
		v := url.Values{}
		if opt.Page != nil {
			v.Set("page", strconv.Itoa(*opt.Page))
		}
		if opt.PerPage != nil {
			v.Set("per_page", strconv.Itoa(*opt.PerPage))
		}
		if opt.Rectangle != nil {
			v.Set("swlng", fmt.Sprintf("%v", opt.Rectangle.Southwest.Longitude))
			v.Set("swlat", fmt.Sprintf("%v", opt.Rectangle.Southwest.Latitude))
			v.Set("nelng", fmt.Sprintf("%v", opt.Rectangle.Northeast.Longitude))
			v.Set("nelat", fmt.Sprintf("%v", opt.Rectangle.Northeast.Latitude))
		}
		if opt.OrderBy != nil {
			v.Set("order_by", *opt.OrderBy)
			if opt.OrderAscending == nil {
				v.Set("order", "desc")
			}
		}
		if opt.OrderAscending != nil {
			if *opt.OrderAscending {
				v.Set("order", "asc")
			} else {
				v.Set("order", "desc")
			}
		}
		if opt.UpdatedSince != nil {
			v.Set("updated_since", opt.UpdatedSince.Format(time.RFC3339))
		}
		if opt.HasGeo != nil {
			v.Set("has[]", "geo")
		}
		if opt.On != nil {
			v.Set("on", opt.On.Format("2006-01-02"))
		}
		if params := v.Encode(); params != "" {
			u += "?" + params
		}
	}
	p, err := c.get(u, &result)
	if err != nil {
		return nil, fmt.Errorf("Error getting observations: %v", err)
	}

	return &ObservationsPage{
		Observations: result,
		Paging:       p,
	}, nil
}

type AddObservationOpt struct {
	SpeciesGuess       string    `json:"species_guess"`
	ObservedOnString   time.Time `json:"observed_on_string,omit_empty"`
	Description        string    `json:"description"`
	Latitude           float64   `json:"latitude"`
	Longitude          float64   `json:"longitude"`
	PositionalAccuracy int32     `json:"positional_accuracy"`
	Tags               string    `json:"tag_list"`
	GeoPrivacy         string    `json:"geoprivacy"`
}

func (c *Client) AddObservation(opt *AddObservationOpt) (*SimpleObservation, error) {
	u := c.buildUrl("/observations.json")

	bodyJson, err := json.Marshal(opt)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", u, bytes.NewReader(bodyJson))
	if err != nil {
		return nil, err
	}
	var result []*SimpleObservation
	err = c.execute(req, &result, http.StatusCreated)
	if err != nil {
		return nil, err
	}

	return result[0], nil
}

func (c *Client) GetObservation(id int64) (*FullObservation, error) {
	var result FullObservation

	u := c.buildUrl("/observations/%d.json", id)
	_, err := c.get(u, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetSimpleObservation(id int64) (*SimpleObservation, error) {
	var result SimpleObservation

	u := c.buildUrl("/observations/%d.json", id)
	_, err := c.get(u, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

type UpdateObservationOpt struct {
	Id                 int64      `json:"-"`
	SpeciesGuess       string     `json:"species_guess,omitempty"`
	ObservedOnString   *time.Time `json:"observed_on_string,omitempty"`
	Description        string     `json:"description,omitempty"`
	Latitude           float64    `json:"latitude,omitempty"`
	Longitude          float64    `json:"longitude,omitempty"`
	PositionalAccuracy int32      `json:"positional_accuracy,omitempty"`
	Tags               string     `json:"tag_list,omitempty"`
	GeoPrivacy         string     `json:"geoprivacy,omitempty"`
}

func (c *Client) UpdateObservation(opt *UpdateObservationOpt) error {
	u := c.buildUrl("/observations/%d.json", opt.Id)

	bodyJson, err := json.Marshal(opt)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", u, bytes.NewReader(bodyJson))
	if err != nil {
		return err
	}
	var p interface{}
	err = c.execute(req, &p, http.StatusCreated)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteObservation(id int64) error {
	u := c.buildUrl("/observations/%d.json", id)

	empty := make([]byte, 0)

	req, err := http.NewRequest("DELETE", u, bytes.NewReader(empty))
	if err != nil {
		return err
	}
	err = c.execute(req, nil, http.StatusCreated)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetObservationsByUsername(username string) (*ObservationsPage, error) {
	var result []*SimpleObservation

	u := c.buildUrl("/observations/%s.json", username)
	p, err := c.get(u, &result)
	if err != nil {
		return nil, err
	}

	return &ObservationsPage{
		Observations: result,
		Paging:       p,
	}, nil
}
