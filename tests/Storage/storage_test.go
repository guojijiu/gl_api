package Storage

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Storage"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStorageManager(t *testing.T) {
	// 创建临时测试目录
	testDir := filepath.Join(os.TempDir(), "storage_test")
	defer os.RemoveAll(testDir)

	// 创建StorageManager
	storageConfig := &Config.StorageConfig{
		BasePath: testDir,
	}
	sm := Storage.NewStorageManager(storageConfig)

	// 测试日志功能
	t.Run("TestLogService", func(t *testing.T) {
		sm.LogInfo("测试信息日志", map[string]interface{}{
			"test": true,
			"time": time.Now().Unix(),
		})

		sm.LogError("测试错误日志", map[string]interface{}{
			"error": "测试错误",
		})
	})

	// 测试缓存功能
	t.Run("TestCacheService", func(t *testing.T) {
		// 设置缓存
		err := sm.Cache("test_key", "test_value", 1*time.Hour)
		if err != nil {
			t.Errorf("设置缓存失败: %v", err)
		}

		// 获取缓存
		value, err := sm.GetCache("test_key")
		if err != nil {
			t.Errorf("获取缓存失败: %v", err)
		}

		if value != "test_value" {
			t.Errorf("缓存值不匹配，期望: test_value, 实际: %v", value)
		}

		// 删除缓存
		err = sm.DeleteCache("test_key")
		if err != nil {
			t.Errorf("删除缓存失败: %v", err)
		}
	})

	// 测试临时文件功能
	t.Run("TestTempService", func(t *testing.T) {
		// 创建临时文件
		tempFile, err := sm.CreateTempFile("test")
		if err != nil {
			t.Errorf("创建临时文件失败: %v", err)
		}
		defer tempFile.Close()

		// 写入测试数据
		testData := "测试数据"
		_, err = tempFile.WriteString(testData)
		if err != nil {
			t.Errorf("写入临时文件失败: %v", err)
		}

		// 获取临时文件统计信息
		tempService := sm.TempService()
		if tempService == nil {
			t.Errorf("临时文件服务未初始化")
			return
		}

		stats := tempService.GetTempStats()
		if stats == nil {
			t.Errorf("获取临时文件统计信息失败")
		}

		count, ok := stats["total_files"].(int)
		if !ok || count < 1 {
			t.Errorf("临时文件数量不正确，期望 >= 1, 实际: %v", count)
		}

		totalSize, ok := stats["total_size"].(int64)
		if !ok || totalSize < int64(len(testData)) {
			t.Errorf("临时文件大小不正确，期望 >= %d, 实际: %v", len(testData), totalSize)
		}
	})

	// 测试存储信息
	t.Run("TestStorageInfo", func(t *testing.T) {
		// 检查基础路径
		if sm.BasePath() != testDir {
			t.Errorf("基础路径不正确，期望: %s, 实际: %s", testDir, sm.BasePath())
		}
	})
}

func TestFileStorage(t *testing.T) {
	// 创建临时测试目录
	testDir := filepath.Join(os.TempDir(), "filestorage_test")
	defer os.RemoveAll(testDir)

	// 创建FileStorage
	fs := Storage.NewFileStorage(testDir)

	// 测试文件存储
	t.Run("TestFileOperations", func(t *testing.T) {
		// 创建测试文件内容
		testContent := "测试文件内容"
		testPath := "test/subdir"
		testFilename := "test.txt"

		// 存储文件
		_, err := fs.Store(strings.NewReader(testContent), testFilename, testPath)
		if err != nil {
			t.Errorf("存储文件失败: %v", err)
		}

		// 检查文件是否存在
		if !fs.Exists(filepath.Join(testPath, testFilename)) {
			t.Errorf("文件不存在")
		}

		// 获取文件大小
		size, err := fs.Size(filepath.Join(testPath, testFilename))
		if err != nil {
			t.Errorf("获取文件大小失败: %v", err)
		}

		if size != int64(len(testContent)) {
			t.Errorf("文件大小不正确，期望: %d, 实际: %d", len(testContent), size)
		}

		// 读取文件
		file, err := fs.Get(filepath.Join(testPath, testFilename))
		if err != nil {
			t.Errorf("读取文件失败: %v", err)
		}

		// 读取内容
		content, err := io.ReadAll(file)
		if err != nil {
			t.Errorf("读取文件内容失败: %v", err)
		}
		file.Close() // 确保文件被关闭

		if string(content) != testContent {
			t.Errorf("文件内容不正确，期望: %s, 实际: %s", testContent, string(content))
		}

		// 删除文件
		err = fs.Delete(filepath.Join(testPath, testFilename))
		if err != nil {
			t.Errorf("删除文件失败: %v", err)
		}

		// 检查文件是否已删除
		if fs.Exists(filepath.Join(testPath, testFilename)) {
			t.Errorf("文件删除失败，文件仍然存在")
		}
	})
}

func TestFileUploadAndDelete(t *testing.T) {
	// 创建临时测试目录
	testDir := filepath.Join(os.TempDir(), "file_upload_test")
	defer os.RemoveAll(testDir)

	// 创建StorageManager
	storageConfig := &Config.StorageConfig{
		BasePath: testDir,
	}
	sm := Storage.NewStorageManager(storageConfig)

	// 测试文件上传功能
	t.Run("TestFileUpload", func(t *testing.T) {
		// 创建测试文件内容
		testContent := "测试文件内容"
		testPath := "test/upload"
		testFilename := "test.txt"

		// 模拟multipart.FileHeader
		// 由于无法直接创建multipart.FileHeader，我们使用io.Reader测试
		reader := strings.NewReader(testContent)

		// 测试公共文件上传
		_, err := sm.StorePublic(reader, testFilename, testPath)
		if err != nil {
			t.Errorf("公共文件上传失败: %v", err)
		}

		// 检查文件是否存在
		expectedPath := filepath.Join(testPath, testFilename)
		if !sm.FileStorage().Exists(expectedPath) {
			t.Errorf("上传的公共文件不存在: %s", expectedPath)
		}

		// 测试私有文件上传
		reader2 := strings.NewReader(testContent)
		_, err2 := sm.StorePrivate(reader2, testFilename, testPath)
		if err2 != nil {
			t.Errorf("私有文件上传失败: %v", err2)
		}

		// 检查私有文件是否存在
		expectedPath2 := filepath.Join(testPath, testFilename)
		if !sm.FileStorage().Exists(expectedPath2) {
			t.Errorf("上传的私有文件不存在: %s", expectedPath2)
		}
	})

	// 测试文件删除功能
	t.Run("TestFileDelete", func(t *testing.T) {
		// 创建测试文件
		testPath := "test/delete"
		testFilename := "delete_test.txt"
		testContent := "要删除的测试文件"

		// 上传文件
		reader := strings.NewReader(testContent)
		_, err := sm.StorePublic(reader, testFilename, testPath)
		if err != nil {
			t.Errorf("创建测试文件失败: %v", err)
		}

		// 检查文件存在
		filePath := filepath.Join(testPath, testFilename)
		if !sm.FileStorage().Exists(filePath) {
			t.Errorf("测试文件创建失败")
		}

		// 删除文件
		err = sm.FileStorage().Delete(filePath)
		if err != nil {
			t.Errorf("删除文件失败: %v", err)
		}

		// 检查文件是否已删除
		if sm.FileStorage().Exists(filePath) {
			t.Errorf("文件删除失败，文件仍然存在")
		}
	})
}
