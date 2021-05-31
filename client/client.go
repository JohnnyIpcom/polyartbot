package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"

	"github.com/johnnyipcom/polyartbot/config"
	"go.uber.org/zap"
)

type Client struct {
	log *zap.Logger
	url *url.URL
	c   *http.Client
}

func New(cfg config.Client, log *zap.Logger) (*Client, error) {
	url, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, err
	}

	return &Client{
		log: log.Named("client"),
		url: url,
		c:   http.DefaultClient,
	}, nil
}

var errorRx = regexp.MustCompile(`{.+"error":(\d+),?}`)

func (c *Client) Raw(ctx context.Context, method string, url string, payload interface{}) ([]byte, error) {
	u, err := c.url.Parse(url)
	if err != nil {
		return nil, err
	}

	c.log.Info("Processing url", zap.String("url", u.String()))

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	req = req.WithContext(ctx)

	resp, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	resp.Close = true
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, extractOk(data)
}

func (c *Client) Get(ctx context.Context, fileID string) ([]byte, error) {
	url := fmt.Sprintf("/v1/image/%s", fileID)
	data, err := c.Raw(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	type respGet struct {
		RespMessage string `json:"message"`
		RespData    []byte `json:"data"`
	}

	var r respGet
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}

	return r.RespData, nil
}

func extractOk(data []byte) error {
	if !errorRx.Match(data) {
		return nil
	}

	type respError struct {
		ErrError string `json:"error"`
	}

	var respErr respError
	if err := json.Unmarshal(data, &respErr); err != nil {
		return err
	}

	return errors.New(respErr.ErrError)
}
