package dexcom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	appID = "d89443d2-327c-4a6f-89e5-496bbb0317db"

	baseUrl          = "https://shareous1.dexcom.com/ShareWebServices/Services"
	loginEndpoint    = "General/LoginPublisherAccountByName"
	authEndpoint     = "General/AuthenticatePublisherAccount"
	readingsEndpoint = "Publisher/ReadPublisherLatestGlucoseValues"
)

type Trend int

const (
	DoubleUp Trend = iota
	SingleUp
	HalfUp
	Flat
	HalfDown
	SingleDown
	DoubleDown
	Missing
)

type Client struct {
	client      *http.Client
	accountName string
	password    string
	sessionID   string
	loc         *time.Location
}

type LoginRequest struct {
	AccountName   string `json:"accountName"`
	Password      string `json:"password"`
	ApplicationID string `json:"applicationId"`
}

type Reading struct {
	WT    string  `json:"WT"` // No clue what this is, web time??
	ST    string  `json:"ST"` // System time.
	DT    string  `json:"DT"` // Display time.
	Value float64 `json:"Value"`
	Trend string  `json:"Trend"`
}

type TransformedReading struct {
	Time  time.Time
	Mmol  float64
	Trend Trend
}

type Option func(*Client)

func WithLocation(loc *time.Location) Option {
	return func(c *Client) {
		c.loc = loc
	}
}

func New(accountName, password string, options ...Option) *Client {
	loc, _ := time.LoadLocation("America/Toronto")

	c := &Client{
		client:      &http.Client{},
		accountName: accountName,
		password:    password,
		loc:         loc,
	}

	for _, option := range options {
		option(c)
	}

	return c
}

func (c *Client) CreateSession() error {
	lreq := &LoginRequest{
		AccountName:   c.accountName,
		Password:      c.password,
		ApplicationID: appID,
	}

	b, err := json.Marshal(lreq)
	if err != nil {
		return err
	}

	resp, err := c.client.Post(baseUrl+"/"+loginEndpoint, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	c.sessionID = strings.Trim(string(body), "\"")

	return nil
}

func Transform(r *Reading, loc *time.Location) (*TransformedReading, error) {
	parsedTime := strings.Trim(r.WT[4:], "()")
	unix, err := strconv.Atoi(parsedTime)
	if err != nil {
		return nil, err
	}

	var trend Trend
	switch r.Trend {
	case "DoubleUp":
		trend = DoubleUp
	case "SingleUp":
		trend = SingleUp
	case "FortyFiveUp":
		trend = HalfUp
	case "Flat":
		trend = Flat
	case "FortyFiveDown":
		trend = HalfDown
	case "SingleDown":
		trend = SingleDown
	case "DoubleDown":
		trend = DoubleDown
	default:
		trend = Missing
	}

	return &TransformedReading{
		Time:  time.Unix(int64(unix/100), 0).In(loc),
		Mmol:  r.Value / 18,
		Trend: trend,
	}, nil
}

func (c *Client) GetReadings(minutes, maxCount int) ([]*TransformedReading, error) {
	if minutes > 1440 || maxCount > 288 {
		return nil, fmt.Errorf("window too large: minutes %d, maxCount %d", minutes, maxCount)
	}

	var readings []*Reading
	for i := 0; i < 2; i++ {
		params := url.Values{
			"sessionId": {c.sessionID},
			"minutes":   {strconv.Itoa(minutes)},
			"maxCount":  {strconv.Itoa(maxCount)},
		}

		resp, err := http.Get(baseUrl + "/" + readingsEndpoint + "?" + params.Encode())
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		err = json.NewDecoder(resp.Body).Decode(&readings)
		if err == nil {
			break
		}

		err = c.CreateSession()
		if err != nil {
			return nil, err
		}
	}

	res := make([]*TransformedReading, len(readings))
	for i, r := range readings {
		tr, err := Transform(r, c.loc)
		if err != nil {
			return nil, err
		}
		res[i] = tr
	}

	return res, nil
}
