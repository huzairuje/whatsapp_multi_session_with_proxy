package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type Client interface {
	Get(url string) ([]byte, error)
	Post(url string, body interface{}) ([]byte, int, error)
	Patch(url string, body interface{}) ([]byte, int, error)
	PostContext(ctx context.Context, url string, body interface{}, header map[string]string) ([]byte, error)
	PutContext(ctx context.Context, url string, body interface{}, header map[string]string) ([]byte, error)
	GetContext(ctx context.Context, url string, header map[string]string) ([]byte, error)
	DeleteContext(ctx context.Context, url string, header map[string]string) ([]byte, error)
}

type httpClient struct{}

func NewHttpClient() Client {
	return &httpClient{}
}

func (h *httpClient) Get(url string) ([]byte, error) {
	resp, err := sendRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error on defer close body : %v", err)
			return
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var messageModel struct {
			Message string `json:"message"`
		}
		bt, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(bt, &messageModel); err != nil {
			return nil, err
		}
		return nil, errors.New(messageModel.Message)
	}

	bt, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bt, nil
}

func (h *httpClient) GetContext(ctx context.Context, url string, header map[string]string) ([]byte, error) {
	resp, err := h.buildRequestContext(ctx, http.MethodGet, url, nil, header)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error on defer close body : %v", err)
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var messageModel struct {
			Message string `json:"message"`
		}
		bt, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(bt, &messageModel); err != nil {
			return nil, err
		}
		return nil, errors.New(messageModel.Message)
	}
	bt, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bt, nil
}

func (h *httpClient) Post(url string, body interface{}) ([]byte, int, error) {
	bytesBuffer := new(bytes.Buffer)
	err := json.NewEncoder(bytesBuffer).Encode(body)
	if err != nil {
		return nil, 0, err
	}
	resp, err := sendRequest(http.MethodPost, url, bytesBuffer)
	if err != nil {
		return nil, 500, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error on defer close body : %v", err)
			return
		}
	}(resp.Body)

	bt, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 500, err
	}

	return bt, resp.StatusCode, nil
}

func (h *httpClient) Patch(url string, body interface{}) ([]byte, int, error) {
	bytesBuffer := new(bytes.Buffer)
	err := json.NewEncoder(bytesBuffer).Encode(body)
	if err != nil {
		return nil, 0, err
	}
	resp, err := sendRequest(http.MethodPatch, url, bytesBuffer)
	if err != nil {
		return nil, 500, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error on defer close body : %v", err)
			return
		}
	}(resp.Body)

	bt, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 500, err
	}

	return bt, resp.StatusCode, nil
}

func (h *httpClient) PostContext(ctx context.Context, url string, body interface{}, header map[string]string) ([]byte, error) {
	resp, err := h.buildRequestContext(ctx, http.MethodPost, url, body, header)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error on defer close body : %v", err)
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var messageModel struct {
			Message string `json:"message"`
		}
		bt, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(bt, &messageModel); err != nil {
			return nil, err
		}
		return nil, errors.New(messageModel.Message)
	}
	bt, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bt, nil
}

func (h *httpClient) PutContext(ctx context.Context, url string, body interface{}, header map[string]string) ([]byte, error) {
	resp, err := h.buildRequestContext(ctx, http.MethodPut, url, body, header)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error on defer close body : %v", err)
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var messageModel struct {
			Message string `json:"message"`
		}
		bt, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(bt, &messageModel); err != nil {
			return nil, err
		}
		return nil, errors.New(messageModel.Message)
	}
	bt, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bt, nil
}

func (h *httpClient) DeleteContext(ctx context.Context, url string, header map[string]string) ([]byte, error) {
	resp, err := h.buildRequestContext(ctx, http.MethodDelete, url, nil, header)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error on defer close body : %v", err)
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var messageModel struct {
			Message string `json:"message"`
		}
		bt, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(bt, &messageModel); err != nil {
			return nil, err
		}
		return nil, errors.New(messageModel.Message)
	}
	bt, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bt, nil
}

func (h *httpClient) buildRequestContext(ctx context.Context, method string, url string, body interface{}, header map[string]string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		bytesBuffer := new(bytes.Buffer)
		if err := json.NewEncoder(bytesBuffer).Encode(body); err != nil {
			return nil, err
		}
		reqBody = bytesBuffer
	} else {
		reqBody = nil
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	if header != nil {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func sendRequest(method string, url string, body io.Reader) (*http.Response, error) {
	// Create request
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// Set Header
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	// Send request
	c := &http.Client{}
	response, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	return response, err
}
