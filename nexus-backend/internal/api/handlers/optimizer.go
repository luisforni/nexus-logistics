package handlers

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type OptimizerHandler struct {
	baseURL string
	client  *http.Client
}

func NewOptimizerHandler(host, port string) *OptimizerHandler {
	return &OptimizerHandler{
		baseURL: fmt.Sprintf("http://%s:%s", host, port),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (h *OptimizerHandler) OptimizeRoute(c *gin.Context) {
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "could not read request body"})
		return
	}

	resp, err := h.proxyPost("/api/v1/optimize/route", body)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, errorResponse{Error: "optimizer unavailable"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", respBody)
}

func (h *OptimizerHandler) proxyPost(path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, h.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	req.ContentLength = int64(len(body))
	req.Body = io.NopCloser(byteReader(body))
	return h.client.Do(req)
}

type byteReader []byte

func (b byteReader) Read(p []byte) (int, error) {
	n := copy(p, b)
	if n == 0 {
		return 0, io.EOF
	}
	return n, nil
}
