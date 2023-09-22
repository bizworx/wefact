package wefact

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"golang.org/x/net/proxy"
)

const defaultEndpoint string = "https://api.mijnwefact.nl/v2/"

// Client
type Client struct {
	key      string
	endpoint string
	client   *http.Client
}

type ProxyConfig struct {
	Host string
}

// New returns a new WeFact API http client.
// If the url is empty set the url to the wefact default endpoint url
func New(key string, proxyConfig *ProxyConfig) *Client {
	// dialer := &net.Dialer{
	// 	Timeout:   30 * time.Second,
	// 	KeepAlive: 30 * time.Second,
	// }
	// var dialContext = Dia;
	var transport = http.DefaultClient.Transport
	if proxyConfig != nil {
		dialer, err := proxy.SOCKS5("tcp", proxyConfig.Host, nil, proxy.Direct)
		if err != nil {
			log.Fatal(err)
		}

		dialContext := func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialer.Dial(network, address)
		}

		transport = &http.Transport{
			DialContext:       dialContext,
			DisableKeepAlives: true,
		}
	}

	return &Client{
		key:      key,
		endpoint: defaultEndpoint,
		client:   &http.Client{Transport: transport},
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
	form.Add("api_key", c.key)
	form.Add("controller", controller)
	form.Add("action", action)

	req, err := http.NewRequest(http.MethodPost, c.endpoint, bytes.NewBufferString(form.Encode()))
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
