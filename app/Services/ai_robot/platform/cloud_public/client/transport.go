package cloudclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

func (a *API) CallCloudPlatformGET(ctx context.Context, fullURL string, userToken string) ([]byte, int, error) {
	if strings.TrimSpace(userToken) == "" {
		return nil, 0, errors.New("缺少用户 token")
	}

	// 失败可重试场景：
	// - 网络错误/超时等（Do 返回 err）
	// - 云平台返回 >=500 或 429
	maxAttempts := 3
	backoffBase := 200 * time.Millisecond

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, 0, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
		if err != nil {
			return nil, 0, err
		}
		req.Header.Set("Token", strings.TrimSpace(userToken))
		req.Header.Set("Accept", "application/json")

		resp, err := a.hc.Do(req)
		if err != nil {
			// 网络错误/超时：允许有限重试
			if attempt < maxAttempts-1 {
				if sleepErr := sleepWithContext(ctx, backoffBase*time.Duration(1<<attempt)); sleepErr != nil {
					return nil, 0, sleepErr
				}
				continue
			}
			return nil, 0, err
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			// 读取失败通常与连接/网络相关，允许有限重试
			if attempt < maxAttempts-1 {
				if sleepErr := sleepWithContext(ctx, backoffBase*time.Duration(1<<attempt)); sleepErr != nil {
					return nil, resp.StatusCode, sleepErr
				}
				continue
			}
			return nil, resp.StatusCode, readErr
		}

		if shouldRetryCloudStatus(resp.StatusCode) && attempt < maxAttempts-1 {
			if sleepErr := sleepWithContext(ctx, backoffBase*time.Duration(1<<attempt)); sleepErr != nil {
				return nil, resp.StatusCode, sleepErr
			}
			continue
		}
		return body, resp.StatusCode, nil
	}

	return nil, 0, errors.New("cloud platform GET failed after retries")
}

func (a *API) CallCloudPlatformPOSTJSON(ctx context.Context, fullURL string, userToken string, payload interface{}) ([]byte, int, error) {
	if strings.TrimSpace(userToken) == "" {
		return nil, 0, errors.New("缺少用户 token")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}

	// 失败可重试场景：
	// - 网络错误/超时等（Do 返回 err）
	// - 云平台返回 >=500 或 429
	maxAttempts := 3
	backoffBase := 200 * time.Millisecond

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, 0, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(raw))
		if err != nil {
			return nil, 0, err
		}
		req.Header.Set("Token", strings.TrimSpace(userToken))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := a.hc.Do(req)
		if err != nil {
			// 网络错误/超时：允许有限重试
			if attempt < maxAttempts-1 {
				if sleepErr := sleepWithContext(ctx, backoffBase*time.Duration(1<<attempt)); sleepErr != nil {
					return nil, 0, sleepErr
				}
				continue
			}
			return nil, 0, err
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			// 读取失败通常与连接/网络相关，允许有限重试
			if attempt < maxAttempts-1 {
				if sleepErr := sleepWithContext(ctx, backoffBase*time.Duration(1<<attempt)); sleepErr != nil {
					return nil, resp.StatusCode, sleepErr
				}
				continue
			}
			return nil, resp.StatusCode, readErr
		}

		if shouldRetryCloudStatus(resp.StatusCode) && attempt < maxAttempts-1 {
			if sleepErr := sleepWithContext(ctx, backoffBase*time.Duration(1<<attempt)); sleepErr != nil {
				return nil, resp.StatusCode, sleepErr
			}
			continue
		}
		return body, resp.StatusCode, nil
	}

	return nil, 0, errors.New("cloud platform POST failed after retries")
}

func shouldRetryCloudStatus(statusCode int) bool {
	// 429：限流
	// 5xx：服务端错误（通常短暂）
	return statusCode == http.StatusTooManyRequests || statusCode >= 500
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
