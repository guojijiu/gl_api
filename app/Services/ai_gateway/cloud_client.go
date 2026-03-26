package aigateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func (s *Service) validateGatewayForCloud() error { return nil }

func (s *Service) CallCloudPlatformGET(ctx context.Context, fullURL string, userToken string) ([]byte, int, error) {
	if strings.TrimSpace(userToken) == "" {
		return nil, 0, errors.New("缺少用户 token")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Token", strings.TrimSpace(userToken))
	req.Header.Set("Accept", "application/json")
	resp, err := s.hc.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

func (s *Service) CallCloudPlatformPOSTJSON(ctx context.Context, fullURL string, userToken string, payload interface{}) ([]byte, int, error) {
	if strings.TrimSpace(userToken) == "" {
		return nil, 0, errors.New("缺少用户 token")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(raw))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Token", strings.TrimSpace(userToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := s.hc.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}
