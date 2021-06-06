package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"

	"github.com/johnnyipcom/polyartbot/config"
	"github.com/johnnyipcom/polyartbot/models"
	"go.uber.org/zap"
	"golang.org/x/oauth2/clientcredentials"
)

type Client interface {
	GetImage(fileID string) (string, []byte, error)
	PostImage(filename string, data []byte, from int64, to int64) (string, error)
	DeleteImage(fileID string) error

	Health() (string, error)
}

type client struct {
	log     *zap.Logger
	baseURL *url.URL
	c       *http.Client
}

func New(cfg config.Client, log *zap.Logger) (Client, error) {
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

	var c *http.Client
	if cfg.OAuth2.Enabled {
		cCfg := clientcredentials.Config{
			ClientID:     cfg.OAuth2.ClientID,
			ClientSecret: cfg.OAuth2.ClientSecret,
			Scopes:       scopes,
			TokenURL:     tokenURL.String(),
		}

		c = cCfg.Client(context.Background())
	} else {
		c = http.DefaultClient
	}

	return &client{
		log:     log.Named("client"),
		baseURL: baseURL,
		c:       c,
	}, nil
}

var errorRx = regexp.MustCompile(`{.+"error":"(.+)".+}`)

func (c *client) raw(method string, url string, payload interface{}) ([]byte, error) {
	u, err := c.baseURL.Parse(url)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

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

func (c *client) sendFiles(url string, rawFiles map[string]io.Reader, params map[string]string) ([]byte, error) {
	if len(rawFiles) == 0 {
		return c.raw(http.MethodPost, url, params)
	}

	pipeReader, pipeWriter := io.Pipe()
	writer := multipart.NewWriter(pipeWriter)
	go func() {
		defer pipeWriter.Close()

		for field, file := range rawFiles {
			if err := addFileToWriter(writer, params["file_name"], field, file); err != nil {
				pipeWriter.CloseWithError(err)
				return
			}
		}
		for field, value := range params {
			if err := writer.WriteField(field, value); err != nil {
				pipeWriter.CloseWithError(err)
				return
			}
		}
		if err := writer.Close(); err != nil {
			pipeWriter.CloseWithError(err)
			return
		}
	}()

	u, err := c.baseURL.Parse(url)
	if err != nil {
		return nil, err
	}

	resp, err := c.c.Post(u.String(), writer.FormDataContentType(), pipeReader)
	if err != nil {
		pipeReader.CloseWithError(err)
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

func addFileToWriter(writer *multipart.Writer, filename, field string, r io.Reader) error {
	part, err := writer.CreateFormFile(field, filename)
	if err != nil {
		return err
	}

	_, err = io.Copy(part, r)
	return err
}

func (c *client) GetImage(fileID string) (string, []byte, error) {
	url := fmt.Sprintf("/cdn/image/%s", fileID)
	data, err := c.raw(http.MethodGet, url, nil)
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

func (c *client) PostImage(filename string, data []byte, from int64, to int64) (string, error) {
	url := fmt.Sprintf("/cdn/image?from=%d&to=%d", from, to)

	files := make(map[string]io.Reader)
	files["file"] = bytes.NewReader(data)

	params := make(map[string]string)
	params["file_name"] = filename

	data, err := c.sendFiles(url, files, params)
	if err != nil {
		return "", err
	}

	type respPost struct {
		RespMessage string            `json:"message"`
		RespFiles   []models.RespFile `json:"files"`
	}

	var r respPost
	if err := json.Unmarshal(data, &r); err != nil {
		return "", err
	}

	return r.RespFiles[0].ID(), nil
}

func (c *client) DeleteImage(fileID string) error {
	url := fmt.Sprintf("/cdn/image/%s", fileID)
	_, err := c.raw(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) Health() (string, error) {
	data, err := c.raw(http.MethodGet, "/health", nil)
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

	var respErr models.RespError
	if err := json.Unmarshal(data, &respErr); err != nil {
		return err
	}

	return respErr
}
