package uploader

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
	"github.com/hashicorp/go-retryablehttp"
)

type ClientWithMiddleware struct {
	HTTPClient *retryablehttp.Client
}

func (c *ClientWithMiddleware) Do(req *retryablehttp.Request) (*http.Response, error) {
	req.Header.Set(contentTypeStr, "application/json")
	req.Header.Set("Accept-Encoding", constants.Gzip)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("can't do client response %w", err)
	}

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case constants.Gzip:
		_, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("can't create gzip.NewReader %w", err)
		}
	default:
		reader = resp.Body
		log.Println(reader)
	}

	return resp, nil
}
