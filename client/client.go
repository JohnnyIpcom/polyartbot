package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"

	"github.com/johnnyipcom/polyartbot/config"
	"github.com/johnnyipcom/polyartbot/glue"
	"go.uber.org/zap"
	"golang.org/x/oauth2/clientcredentials"
)

type Client struct {
	log     *zap.Logger
	baseURL *url.URL
	c       *http.Client
}

func New(cfg config.Client, log *zap.Logger) (*Client, error) {
	baseURL, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, err
	}

	scopes := make([]string, len(cfg.OAuth2.Scopes))
	copy(scopes, cfg.OAuth2.Scopes)

	tokenURL, err := baseURL.Parse(cfg.OAuth2.TokenURL)
	if err != nil {
		return nil, err
	}

	cCfg := clientcredentials.Config{
		ClientID:     cfg.OAuth2.ClientID,
		ClientSecret: cfg.OAuth2.ClientSecret,
		Scopes:       scopes,
		TokenURL:     tokenURL.String(),
	}

	return &Client{
		log:     log.Named("client"),
		baseURL: baseURL,
		c:       cCfg.Client(context.Background()),
	}, nil
}

var errorRx = regexp.MustCompile(`{.+"error":(\d+),?}`)

func (c *Client) Raw(ctx context.Context, method string, url string, payload interface{}) ([]byte, error) {
	u, err := c.baseURL.Parse(url)
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

func (c *Client) GetImage(ctx context.Context, fileID string) (string, []byte, error) {
	url := fmt.Sprintf("/cdn/image/%s", fileID)
	data, err := c.Raw(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", nil, err
	}

	type respGet struct {
		RespMessage string `json:"message"`
		RespData    []byte `json:"data"`
	}

	var r respGet
	if err := json.Unmarshal(data, &r); err != nil {
		return "", nil, err
	}

	return r.RespMessage, r.RespData, nil
}

func (c *Client) Health(ctx context.Context) (string, error) {
	data, err := c.Raw(ctx, http.MethodGet, "/health", nil)
	if err != nil {
		return "", err
	}

	type respHealth struct {
		RespMessage string `json:"status"`
	}

	var r respHealth
	if err := json.Unmarshal(data, &r); err != nil {
		return "", err
	}

	return r.RespMessage, nil
}

func extractOk(data []byte) error {
	if !errorRx.Match(data) {
		return nil
	}

	var respErr glue.RespError
	if err := json.Unmarshal(data, &respErr); err != nil {
		return err
	}

	return respErr
}
