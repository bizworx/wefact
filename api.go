package wefact

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/mitchellh/mapstructure"
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

type Response struct {
	Controller     string                 `mapstructure:"controller"`
	Action         string                 `mapstructure:"action"`
	Status         string                 `mapstructure:"status"`
	Date           string                 `mapstructure:"date"`
	TotalResults   int                    `mapstructure:"totalresults"`
	CurrentResults int                    `mapstructure:"currentresults"`
	Offset         int                    `mapstructure:"offset"`
	Result         map[string]interface{} `mapstructure:",remain"`
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
func (c *Client) Request(controller, action string, form url.Values) (*Response, error) {
	if form == nil {
		form = url.Values{}
	}
	form.Add("api_key", c.config.Key)
	form.Add("controller", controller)
	form.Add("action", action)

	req, err := http.NewRequest(http.MethodPost, c.config.Url, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, errors.Wrap(err, "http.NewRequest")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "http.Do")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, newRequestError(errors.New(http.StatusText(http.StatusUnauthorized)))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, newRequestError(err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, newRequestError(err)
	}

	var output = new(Response)
	if err := mapstructure.Decode(response, output); err != nil {
		return nil, errors.Wrap(err, "mapstructure.Decode")
	}

	return output, nil
}
