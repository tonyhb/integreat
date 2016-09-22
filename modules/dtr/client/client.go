package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

type Opts struct {
	Logger *logrus.Logger
	Host   string
	User   string
	Pass   string
}

func New(opts Opts) Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	c := http.Client{
		Transport: transport,
	}

	return Client{
		client: c,
		logger: opts.Logger,
		host:   opts.Host,
		user:   opts.User,
		pass:   opts.Pass,
	}
}

// Client ...
type Client struct {
	client http.Client
	logger *logrus.Logger
	host   string
	user   string
	pass   string
}

func (c Client) Do(method, path string, data map[string]interface{}) (map[string]interface{}, error) {
	req, err := c.request(method, path, data)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	defer resp.Body.Close()

	byt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read failed")
	}

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("invalid status: %d (body %s)", resp.StatusCode, byt)
	}

	result := make(map[string]interface{})
	err = json.Unmarshal(byt, &result)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("error unmarshaling request (body %s)", byt))
	}

	return result, nil
}

func (c Client) request(method, path string, data map[string]interface{}) (*http.Request, error) {
	url := fmt.Sprintf("https://%s%s", c.host, path)
	byt, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling request data")
	}
	buf := bytes.NewBuffer(byt)
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, errors.Wrap(err, "error creating request type")
	}
	req.SetBasicAuth(c.user, c.pass)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("accept", "application/json")
	return req, nil
}
