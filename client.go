package main

import (
	"fmt"
	"io"
	"net/http"
)

// REST API unofficial reference: https://xled-docs.readthedocs.io/en/latest/rest_api.html

type Client struct {
	baseURL string
	client  *http.Client
	token   string
}

func MakeClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *Client) SetAuthToken(token string) {
	c.token = token
}

func (c *Client) do(method string, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.baseURL+"/"+url, body)
	if err != nil {
		return nil, err
	}

	if c.token != "" {
		req.Header.Set("X-Auth-Token", c.token)
	}

	return c.client.Do(req)
}

func (c *Client) Get(url string) (map[string]interface{}, error) {
	resp, err := c.do(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received status code %d", resp.StatusCode)
	}

	return parseXledResponse(resp)
}

func (c *Client) Post(url string, body interface{}) (map[string]interface{}, error) {
	reqBody, err := toJson(body)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(http.MethodPost, url, reqBody)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received status code %d", resp.StatusCode)
	}

	return parseXledResponse(resp)
}

func parseXledResponse(resp *http.Response) (map[string]interface{}, error) {
	respBody := make(map[string]interface{})
	err := fromJson(resp.Body, &respBody)
	if err != nil {
		return nil, err
	}

	if responseCode, ok := respBody["code"]; !ok {
		return nil, fmt.Errorf("No response code in response")
	} else if responseCode != float64(1000) {
		return nil, fmt.Errorf("Incorrect response code %d", responseCode)
	}

	return respBody, nil
}
