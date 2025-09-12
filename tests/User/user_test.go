package tests

import (
	"cloud-platform-api/tests/testsetup"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func init() {
	// 确保在单文件执行时也完成公共初始化
	if testsetup.Router == nil {
		testsetup.Init()
	}
}

// go test -v tests/User/user_test.go -run TestGetUser
func TestGetUser(t *testing.T) {

	// 创建测试用户和token
	userInfo, err := testsetup.CreateTestUserWithToken("testuser", "test@example.com", "password123", "user")
	if err != nil {
		t.Fatal(err)
	}

	// 发起真实请求
	req, err := http.NewRequest("GET", "/api/v1/users/"+strconv.FormatUint(uint64(userInfo.User.ID), 10), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+userInfo.Token)

	w := httptest.NewRecorder()
	testsetup.Router.ServeHTTP(w, req)

	fmt.Printf("响应状态码: %d\n", w.Code)
	fmt.Printf("响应内容: %s\n", w.Body.String())

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 %d, 得到 %d, 响应: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// 解析并断言响应
	type apiResp struct {
		Success bool                   `json:"success"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data"`
	}
	var body apiResp
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应JSON失败: %v, 原始: %s", err, w.Body.String())
	}

	if !body.Success {
		t.Fatalf("期望 success=true, 实际: %v", body.Success)
	}
	if body.Message != "User retrieved successfully" {
		t.Fatalf("期望 message='User retrieved successfully', 实际: %v", body.Message)
	}
}
