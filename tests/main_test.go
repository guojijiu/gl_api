package tests

import (
	"cloud-platform-api/tests/testsetup"
	"os"
	"testing"
)

// TestMain: 在本包所有测试前统一初始化
func TestMain(m *testing.M) {
	// 初始化测试环境（只执行一次）
	testsetup.Init()
	// 执行测试
	os.Exit(m.Run())
}
