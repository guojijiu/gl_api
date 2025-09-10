package tests

import (
	"cloud-platform-api/tests/testsetup"
	"net/http"
	"net/http/httptest"
	"testing"
)

// go test -v tests/health_test.go -run TestHealthCheck
func TestHealthCheck(t *testing.T) {
	// 使用公共路由（已注册全部路由）
	req, err := http.NewRequest("GET", "/api/v1/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	testsetup.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusOK, w.Code)
	}

	expected := `{"message":"Service is running","status":"ok","success":true}`
	if w.Body.String() != expected {
		t.Errorf("期望响应 %s, 得到 %s", expected, w.Body.String())
	}
}
