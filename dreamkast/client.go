package dreamkast

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/hashicorp/go-retryablehttp"
)

type Client interface {
	ListConferences(ctx context.Context) (ListConferencesResp, error)
}

type ClientImpl struct {
	client        *http.Client
	dkEndpointUrl url.URL
}

func NewClient(dkEndpointUrl string) (Client, error) {
	dkUrl, err := url.Parse(dkEndpointUrl)
	if err != nil {
		return nil, err
	}
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	standardClient := retryClient.StandardClient()

	return &ClientImpl{client: standardClient, dkEndpointUrl: *dkUrl}, nil
}

func (c *ClientImpl) ListConferences(ctx context.Context) (ListConferencesResp, error) {
	url := c.dkEndpointUrl
	url.Path = filepath.Join(url.Path, "/api/v1/events")
	req, err := http.NewRequestWithContext(ctx, "GET", "url.String()", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ListConferencesResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}
