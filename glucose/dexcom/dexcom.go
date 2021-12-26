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

	"github.com/algao1/ichor/store"
	"go.uber.org/zap"
)

const (
	appID = "d89443d2-327c-4a6f-89e5-496bbb0317db"

	baseUrl          = "https://shareous1.dexcom.com/ShareWebServices/Services"
	loginEndpoint    = "General/LoginPublisherAccountByName"
	authEndpoint     = "General/AuthenticatePublisherAccount"
	readingsEndpoint = "Publisher/ReadPublisherLatestGlucoseValues"
)

type Client struct {
	client      *http.Client
	logger      *zap.Logger
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
	Trend store.Trend
}

type Option func(*Client)

func WithLocation(loc *time.Location) Option {
	return func(c *Client) {
		c.loc = loc
	}
}

func New(accountName, password string, logger *zap.Logger, options ...Option) *Client {
	loc, _ := time.LoadLocation("America/Toronto")

	c := &Client{
		client:      &http.Client{},
		logger:      logger,
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

	c.logger.Debug("making login request for sessionID",
		zap.ByteString("request", b),
	)

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

	c.logger.Debug("successfully obtained sessionID",
		zap.String("sessionID", c.sessionID),
	)

	return nil
}

func Transform(r *Reading, loc *time.Location) (*TransformedReading, error) {
	parsedTime := strings.Trim(r.WT[4:], "()")
	unix, err := strconv.Atoi(parsedTime)
	if err != nil {
		return nil, err
	}

	var trend store.Trend
	switch r.Trend {
	case "DoubleUp":
		trend = store.DoubleUp
	case "SingleUp":
		trend = store.SingleUp
	case "FortyFiveUp":
		trend = store.HalfUp
	case "Flat":
		trend = store.Flat
	case "FortyFiveDown":
		trend = store.HalfDown
	case "SingleDown":
		trend = store.SingleDown
	case "DoubleDown":
		trend = store.DoubleDown
	default:
		trend = store.Missing
	}

	return &TransformedReading{
		Time:  time.Unix(int64(unix/1000), 0).In(loc),
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

		c.logger.Debug("making fetch request",
			zap.String("sessionID", c.sessionID),
			zap.Int("minutes", minutes),
			zap.Int("maximum count", maxCount),
		)

		resp, err := http.Get(baseUrl + "/" + readingsEndpoint + "?" + params.Encode())
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		err = json.NewDecoder(resp.Body).Decode(&readings)
		if err == nil {
			break
		}

		c.logger.Debug("failed to decode readings response, restarting session")

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

	c.logger.Debug("received readings from share API",
		zap.Int("count", len(res)),
	)

	return res, nil
}
