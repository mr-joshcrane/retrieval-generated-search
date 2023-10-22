package pinecone

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
)

type Client struct {
	ApiKey string
	url    string
	client *http.Client
}

func NewClient(apiKey string, url string) (*Client, error) {
	return &Client{
		ApiKey: apiKey,
		url:    url,
		client: http.DefaultClient,
	}, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Api-Key", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return c.client.Do(req)
}

type UpsertRequest struct {
	Vectors []struct {
		ID     string    `json:"id"`
		Values []float64 `json:"values"`
	} `json:"vectors"`
}

type UpsertResponse struct {
}

func (c *Client) Upsert(uuid string, vector []float64) (*http.Response, error) {
	endpoint := "/vectors/upsert"
	data, err := json.Marshal(UpsertRequest{
		Vectors: []struct {
			ID     string    `json:"id"`
			Values []float64 `json:"values"`
		}{
			{
				ID:     uuid,
				Values: vector,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", c.url+endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

type QueryRequest struct {
	IncludeValues   string    `json:"includeValues"`
	IncludeMetadata string    `json:"includeMetadata"`
	TopK            int       `json:"topK"`
	Vector          []float64 `json:"vector"`
}

type QueryResponse struct {
	Matches []struct {
		ID string `json:"id"`
	}
}

func (c *Client) Query(vector []float64) (*http.Response, error) {
	endpoint := "/query"
	body := QueryRequest{
		IncludeValues:   "true",
		IncludeMetadata: "true",
		TopK:            1,
		Vector:          vector,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.url+endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	data, err = httputil.DumpRequest(req, true)
	if err != nil {
		return nil, err
	}
	os.WriteFile("pinecone_request.txt", data, 0644)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("pinecone query failed %s", resp.Status)
	}
	return resp, nil
}
