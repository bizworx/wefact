package wefact

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

const defaultEndpoint string = "https://api.mijnwefact.nl/v2/"

type Config struct {
	Key string // WeFact API key
	Url string // WeFact endpoint default: https://api.mijnwefact.nl/v2/
}

// Client
type Client struct {
	config *Config
	client *http.Client
}

// New returns a new WeFact API http client.
// If the url is empty set the url to the wefact default endpoint url
func New(config *Config) *Client {

	if config.Url == "" {
		config.Url = defaultEndpoint
	}
	return &Client{
		config: config,
		client: http.DefaultClient,
	}
}

type requestError struct {
	Controller string    `json:"controller"`
	Action     string    `json:"action"`
	Status     string    `json:"status"`
	Date       time.Time `json:"date"`
	Err        error     `json:"errors"`
}

func newRequestError(err error) requestError {
	return requestError{Controller: "invalid", Action: "invalid", Status: "error", Date: time.Now(), Err: err}
}

func (e requestError) Error() string {
	enc, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	return string(enc)
}

// Request execute an API call to the wefact endpoint
func (c *Client) Request(controller, action string, form url.Values, results interface{}) error {
	if form == nil {
		form = url.Values{}
	}
	form.Add("api_key", c.config.Key)
	form.Add("controller", controller)
	form.Add("action", action)

	req, err := http.NewRequest(http.MethodPost, c.config.Url, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return errors.Wrap(err, "http.NewRequest")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "http.Do")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return newRequestError(errors.New(http.StatusText(http.StatusUnauthorized)))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return newRequestError(err)
	}

	if err := json.Unmarshal(body, results); err != nil {
		return newRequestError(err)
	}

	return nil
}
