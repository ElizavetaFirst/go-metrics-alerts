package uploader

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
	"github.com/pkg/errors"
)

type ClientWithMiddleware struct {
	HTTPClient *http.Client
}

func (c *ClientWithMiddleware) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set(contentTypeStr, "application/json")
	req.Header.Set("Accept-Encoding", constants.Gzip)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "can't do client response")
	}

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case constants.Gzip:
		_, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "can't create gzip.NewReader")
		}
	default:
		reader = resp.Body
		fmt.Println(reader)
	}

	return resp, nil
}
