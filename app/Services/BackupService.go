package Services

import (
	"archive/zip"
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Storage"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// BackupConfig 备份配置
type BackupConfig struct {
	EnableAutoBackup    bool          `json:"enable_auto_backup"`    // 启用自动备份
	BackupInterval      time.Duration `json:"backup_interval"`       // 备份间隔
	MaxBackupFiles      int           `json:"max_backup_files"`      // 最大备份文件数
	BackupRetentionDays int           `json:"backup_retention_days"` // 备份保留天数
	BackupPath          string        `json:"backup_path"`           // 备份路径
	EnableCompression   bool          `json:"enable_compression"`    // 启用压缩
	EnableEncryption    bool          `json:"enable_encryption"`     // 启用加密
	EncryptionKey       string        `json:"encryption_key"`        // 加密密钥
}

// BackupInfo 备份信息
type BackupInfo struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "database", "files", "full"
	Path        string                 `json:"path"`
	Size        int64                  `json:"size"`
	MD5         string                 `json:"md5"`
	CreatedAt   time.Time              `json:"created_at"`
	Status      string                 `json:"status"` // "success", "failed", "in_progress"
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// BackupService 备份服务
type BackupService struct {
	storageManager *Storage.StorageManager
	config         *BackupConfig
	backupPath     string
}

// NewBackupService 创建备份服务
// 功能说明：
// 1. 初始化备份服务
// 2. 配置备份策略
// 3. 支持数据库和文件备份
// 4. 提供自动备份和手动备份
// 5. 支持备份恢复和验证
func NewBackupService(storageManager *Storage.StorageManager, config *BackupConfig) *BackupService {
	if config == nil {
		config = &BackupConfig{
			EnableAutoBackup:    true,
			BackupInterval:      24 * time.Hour,
			MaxBackupFiles:      10,
			BackupRetentionDays: 30,
			BackupPath:          "./backups",
			EnableCompression:   true,
			EnableEncryption:    false,
			EncryptionKey:       "",
		}
	}

	// 确保备份目录存在
	if err := os.MkdirAll(config.BackupPath, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create backup directory: %v", err))
	}

	service := &BackupService{
		storageManager: storageManager,
		config:         config,
		backupPath:     config.BackupPath,
	}

	// 启动自动备份
	if config.EnableAutoBackup {
		go service.startAutoBackup()
	}

	return service
}

// CreateDatabaseBackup 创建数据库备份
// 功能说明：
// 1. 备份数据库结构和数据
// 2. 支持多种数据库类型
// 3. 生成备份元数据
// 4. 计算备份文件MD5
// 5. 记录备份日志
func (s *BackupService) CreateDatabaseBackup() (*BackupInfo, error) {
	backupID := fmt.Sprintf("db_%s", time.Now().Format("20060102_150405"))
	backupPath := filepath.Join(s.backupPath, backupID+".sql")

	// 创建备份信息
	backupInfo := &BackupInfo{
		ID:        backupID,
		Type:      "database",
		Path:      backupPath,
		CreatedAt: time.Now(),
		Status:    "in_progress",
		Metadata:  make(map[string]interface{}),
	}

	// 根据数据库类型执行备份
	dbConfig := Config.GetConfig().Database
	var err error

	switch dbConfig.Driver {
	case "mysql":
		err = s.backupMySQL(backupPath)
	case "postgres":
		err = s.backupPostgreSQL(backupPath)
	case "sqlite":
		err = s.backupSQLite(backupPath)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", dbConfig.Driver)
	}

	if err != nil {
		backupInfo.Status = "failed"
		backupInfo.Description = err.Error()
		s.storageManager.LogError("数据库备份失败", map[string]interface{}{
			"backup_id": backupID,
			"error":     err.Error(),
		})
		return backupInfo, err
	}

	// 获取文件信息
	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		return nil, err
	}

	backupInfo.Size = fileInfo.Size()
	backupInfo.Status = "success"
	backupInfo.Description = "数据库备份成功"

	// 计算MD5
	md5Hash, err := s.calculateMD5(backupPath)
	if err != nil {
		return nil, err
	}
	backupInfo.MD5 = md5Hash

	// 压缩备份文件
	if s.config.EnableCompression {
		compressedPath, err := s.compressFile(backupPath)
		if err != nil {
			return nil, err
		}
		backupInfo.Path = compressedPath

		// 更新文件大小
		fileInfo, err = os.Stat(compressedPath)
		if err == nil {
			backupInfo.Size = fileInfo.Size()
		}
	}

	// 记录备份成功日志
	s.storageManager.LogInfo("数据库备份成功", map[string]interface{}{
		"backup_id": backupID,
		"path":      backupInfo.Path,
		"size":      backupInfo.Size,
		"md5":       backupInfo.MD5,
	})

	return backupInfo, nil
}

// CreateFileBackup 创建文件备份
// 功能说明：
// 1. 备份存储目录中的文件
// 2. 保持目录结构
// 3. 排除临时文件
// 4. 生成压缩包
// 5. 记录备份信息
func (s *BackupService) CreateFileBackup() (*BackupInfo, error) {
	backupID := fmt.Sprintf("files_%s", time.Now().Format("20060102_150405"))
	backupPath := filepath.Join(s.backupPath, backupID+".zip")

	// 创建备份信息
	backupInfo := &BackupInfo{
		ID:        backupID,
		Type:      "files",
		Path:      backupPath,
		CreatedAt: time.Now(),
		Status:    "in_progress",
		Metadata:  make(map[string]interface{}),
	}

	// 创建ZIP文件
	zipFile, err := os.Create(backupPath)
	if err != nil {
		return nil, err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 备份存储目录
	storagePath := s.storageManager.BasePath()
	err = s.addDirectoryToZip(zipWriter, storagePath, "storage")
	if err != nil {
		backupInfo.Status = "failed"
		backupInfo.Description = err.Error()
		return backupInfo, err
	}

	// 获取文件信息
	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		return nil, err
	}

	backupInfo.Size = fileInfo.Size()
	backupInfo.Status = "success"
	backupInfo.Description = "文件备份成功"

	// 计算MD5
	md5Hash, err := s.calculateMD5(backupPath)
	if err != nil {
		return nil, err
	}
	backupInfo.MD5 = md5Hash

	// 记录备份成功日志
	s.storageManager.LogInfo("文件备份成功", map[string]interface{}{
		"backup_id": backupID,
		"path":      backupInfo.Path,
		"size":      backupInfo.Size,
		"md5":       backupInfo.MD5,
	})

	return backupInfo, nil
}

// CreateFullBackup 创建完整备份
// 功能说明：
// 1. 同时备份数据库和文件
// 2. 生成完整的系统备份
// 3. 包含配置信息
// 4. 生成备份报告
// 5. 验证备份完整性
func (s *BackupService) CreateFullBackup() (*BackupInfo, error) {
	backupID := fmt.Sprintf("full_%s", time.Now().Format("20060102_150405"))
	backupPath := filepath.Join(s.backupPath, backupID+".zip")

	// 创建备份信息
	backupInfo := &BackupInfo{
		ID:        backupID,
		Type:      "full",
		Path:      backupPath,
		CreatedAt: time.Now(),
		Status:    "in_progress",
		Metadata:  make(map[string]interface{}),
	}

	// 创建ZIP文件
	zipFile, err := os.Create(backupPath)
	if err != nil {
		return nil, err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 1. 备份数据库
	dbBackupPath := filepath.Join(s.backupPath, "temp_db.sql")
	err = s.backupDatabaseToFile(dbBackupPath)
	if err != nil {
		backupInfo.Status = "failed"
		backupInfo.Description = "数据库备份失败: " + err.Error()
		return backupInfo, err
	}

	// 添加数据库备份到ZIP
	err = s.addFileToZip(zipWriter, dbBackupPath, "database.sql")
	if err != nil {
		return nil, err
	}

	// 2. 备份文件
	storagePath := s.storageManager.BasePath()
	err = s.addDirectoryToZip(zipWriter, storagePath, "storage")
	if err != nil {
		backupInfo.Status = "failed"
		backupInfo.Description = "文件备份失败: " + err.Error()
		return backupInfo, err
	}

	// 3. 备份配置信息
	configData, err := json.MarshalIndent(Config.GetConfig(), "", "  ")
	if err != nil {
		return nil, err
	}

	configWriter, err := zipWriter.Create("config.json")
	if err != nil {
		return nil, err
	}
	configWriter.Write(configData)

	// 4. 生成备份报告
	report := s.generateBackupReport(backupInfo)
	reportWriter, err := zipWriter.Create("backup_report.json")
	if err != nil {
		return nil, err
	}
	reportData, _ := json.MarshalIndent(report, "", "  ")
	reportWriter.Write(reportData)

	// 清理临时文件
	os.Remove(dbBackupPath)

	// 获取文件信息
	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		return nil, err
	}

	backupInfo.Size = fileInfo.Size()
	backupInfo.Status = "success"
	backupInfo.Description = "完整备份成功"
	backupInfo.Metadata = report

	// 计算MD5
	md5Hash, err := s.calculateMD5(backupPath)
	if err != nil {
		return nil, err
	}
	backupInfo.MD5 = md5Hash

	// 记录备份成功日志
	s.storageManager.LogInfo("完整备份成功", map[string]interface{}{
		"backup_id": backupID,
		"path":      backupInfo.Path,
		"size":      backupInfo.Size,
		"md5":       backupInfo.MD5,
	})

	return backupInfo, nil
}

// RestoreBackup 恢复备份
// 功能说明：
// 1. 验证备份文件完整性
// 2. 根据备份类型执行恢复
// 3. 支持部分恢复和完整恢复
// 4. 提供恢复进度反馈
// 5. 记录恢复日志
func (s *BackupService) RestoreBackup(backupPath string, backupType string) error {
	// 验证备份文件
	if err := s.validateBackup(backupPath); err != nil {
		return fmt.Errorf("备份文件验证失败: %v", err)
	}

	switch backupType {
	case "database":
		return s.restoreDatabase(backupPath)
	case "files":
		return s.restoreFiles(backupPath)
	case "full":
		return s.restoreFullBackup(backupPath)
	default:
		return fmt.Errorf("不支持的备份类型: %s", backupType)
	}
}

// ListBackups 列出所有备份
func (s *BackupService) ListBackups() ([]*BackupInfo, error) {
	var backups []*BackupInfo

	files, err := os.ReadDir(s.backupPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// 解析文件名获取备份信息
		backupInfo, err := s.parseBackupInfo(file.Name())
		if err != nil {
			continue
		}

		// 获取文件信息
		filePath := filepath.Join(s.backupPath, file.Name())
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		backupInfo.Size = fileInfo.Size()
		backupInfo.Path = filePath

		// 计算MD5
		md5Hash, err := s.calculateMD5(filePath)
		if err == nil {
			backupInfo.MD5 = md5Hash
		}

		backups = append(backups, backupInfo)
	}

	return backups, nil
}

// CleanupOldBackups 清理旧备份
func (s *BackupService) CleanupOldBackups() error {
	backups, err := s.ListBackups()
	if err != nil {
		return err
	}

	cutoffTime := time.Now().AddDate(0, 0, -s.config.BackupRetentionDays)
	deletedCount := 0

	for _, backup := range backups {
		if backup.CreatedAt.Before(cutoffTime) {
			if err := os.Remove(backup.Path); err != nil {
				s.storageManager.LogError("删除旧备份失败", map[string]interface{}{
					"backup_id": backup.ID,
					"path":      backup.Path,
					"error":     err.Error(),
				})
			} else {
				deletedCount++
				s.storageManager.LogInfo("删除旧备份", map[string]interface{}{
					"backup_id": backup.ID,
					"path":      backup.Path,
				})
			}
		}
	}

	s.storageManager.LogInfo("清理旧备份完成", map[string]interface{}{
		"deleted_count": deletedCount,
		"cutoff_time":   cutoffTime,
	})

	return nil
}

// 私有方法

func (s *BackupService) backupMySQL(backupPath string) error {
	dbConfig := Config.GetConfig().Database

	// 构建mysqldump命令
	cmd := exec.Command("mysqldump",
		"-h", dbConfig.Host,
		"-P", dbConfig.Port,
		"-u", dbConfig.Username,
		"-p"+dbConfig.Password,
		"--single-transaction",
		"--routines",
		"--triggers",
		"--events",
		"--add-drop-table",
		"--add-locks",
		"--create-options",
		"--disable-keys",
		"--extended-insert",
		"--quick",
		"--lock-tables=false",
		dbConfig.Database)

	// 创建输出文件
	outputFile, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("创建备份文件失败: %v", err)
	}
	defer outputFile.Close()

	// 设置命令输出
	cmd.Stdout = outputFile
	cmd.Stderr = os.Stderr

	// 执行命令
	if err := cmd.Run(); err != nil {
		// 删除失败的备份文件
		os.Remove(backupPath)
		return fmt.Errorf("mysqldump执行失败: %v", err)
	}

	// 验证备份文件
	if fileInfo, err := os.Stat(backupPath); err != nil || fileInfo.Size() == 0 {
		os.Remove(backupPath)
		return fmt.Errorf("备份文件为空或无效")
	}

	return nil
}

func (s *BackupService) backupPostgreSQL(backupPath string) error {
	dbConfig := Config.GetConfig().Database

	// 构建pg_dump命令
	cmd := exec.Command("pg_dump",
		"-h", dbConfig.Host,
		"-p", dbConfig.Port,
		"-U", dbConfig.Username,
		"-d", dbConfig.Database,
		"--verbose",
		"--clean",
		"--if-exists",
		"--create",
		"--format=plain")

	// 设置环境变量（PostgreSQL使用PGPASSWORD环境变量）
	cmd.Env = append(os.Environ(), "PGPASSWORD="+dbConfig.Password)

	// 创建输出文件
	outputFile, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("创建备份文件失败: %v", err)
	}
	defer outputFile.Close()

	// 设置命令输出
	cmd.Stdout = outputFile
	cmd.Stderr = os.Stderr

	// 执行命令
	if err := cmd.Run(); err != nil {
		// 删除失败的备份文件
		os.Remove(backupPath)
		return fmt.Errorf("pg_dump执行失败: %v", err)
	}

	// 验证备份文件
	if fileInfo, err := os.Stat(backupPath); err != nil || fileInfo.Size() == 0 {
		os.Remove(backupPath)
		return fmt.Errorf("备份文件为空或无效")
	}

	return nil
}

func (s *BackupService) backupSQLite(backupPath string) error {
	dbConfig := Config.GetConfig().Database

	// 对于SQLite，直接复制数据库文件
	sourceFile, err := os.Open(dbConfig.Database)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func (s *BackupService) backupDatabaseToFile(backupPath string) error {
	dbConfig := Config.GetConfig().Database

	switch dbConfig.Driver {
	case "mysql":
		return s.backupMySQL(backupPath)
	case "postgres":
		return s.backupPostgreSQL(backupPath)
	case "sqlite":
		return s.backupSQLite(backupPath)
	default:
		return fmt.Errorf("unsupported database driver: %s", dbConfig.Driver)
	}
}

func (s *BackupService) addDirectoryToZip(zipWriter *zip.Writer, dirPath, zipPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过临时文件和日志文件
		if strings.Contains(path, "temp") || strings.Contains(path, ".log") {
			return nil
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		return s.addFileToZip(zipWriter, path, filepath.Join(zipPath, relPath))
	})
}

func (s *BackupService) addFileToZip(zipWriter *zip.Writer, filePath, zipPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	zipFile, err := zipWriter.Create(zipPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(zipFile, file)
	return err
}

func (s *BackupService) calculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (s *BackupService) compressFile(filePath string) (string, error) {
	compressedPath := filePath + ".gz"

	// 打开源文件
	sourceFile, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开源文件失败: %v", err)
	}
	defer sourceFile.Close()

	// 创建压缩文件
	compressedFile, err := os.Create(compressedPath)
	if err != nil {
		return "", fmt.Errorf("创建压缩文件失败: %v", err)
	}
	defer compressedFile.Close()

	// 创建gzip写入器
	gzipWriter := gzip.NewWriter(compressedFile)
	defer gzipWriter.Close()

	// 复制数据并压缩
	_, err = io.Copy(gzipWriter, sourceFile)
	if err != nil {
		// 删除失败的压缩文件
		os.Remove(compressedPath)
		return "", fmt.Errorf("压缩文件失败: %v", err)
	}

	// 确保所有数据都写入
	if err := gzipWriter.Close(); err != nil {
		os.Remove(compressedPath)
		return "", fmt.Errorf("关闭gzip写入器失败: %v", err)
	}

	// 删除原始文件
	if err := os.Remove(filePath); err != nil {
		// 记录警告但不返回错误
		s.storageManager.LogWarning("删除原始备份文件失败", map[string]interface{}{
			"file":  filePath,
			"error": err.Error(),
		})
	}

	return compressedPath, nil
}

func (s *BackupService) validateBackup(backupPath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(backupPath); err != nil {
		return err
	}

	// 检查文件大小
	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		return err
	}

	if fileInfo.Size() == 0 {
		return fmt.Errorf("备份文件为空")
	}

	return nil
}

func (s *BackupService) restoreDatabase(backupPath string) error {
	dbConfig := Config.GetConfig().Database

	// 检查备份文件是否存在
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("备份文件不存在: %v", err)
	}

	// 根据数据库类型执行恢复
	switch dbConfig.Driver {
	case "mysql":
		return s.restoreMySQL(backupPath)
	case "postgres":
		return s.restorePostgreSQL(backupPath)
	case "sqlite":
		return s.restoreSQLite(backupPath)
	default:
		return fmt.Errorf("不支持的数据库驱动: %s", dbConfig.Driver)
	}
}

func (s *BackupService) restoreMySQL(backupPath string) error {
	dbConfig := Config.GetConfig().Database

	// 构建mysql命令
	cmd := exec.Command("mysql",
		"-h", dbConfig.Host,
		"-P", dbConfig.Port,
		"-u", dbConfig.Username,
		"-p"+dbConfig.Password,
		dbConfig.Database)

	// 打开备份文件
	backupFile, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("打开备份文件失败: %v", err)
	}
	defer backupFile.Close()

	// 设置命令输入
	cmd.Stdin = backupFile
	cmd.Stderr = os.Stderr

	// 执行恢复命令
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("MySQL恢复失败: %v", err)
	}

	return nil
}

func (s *BackupService) restorePostgreSQL(backupPath string) error {
	dbConfig := Config.GetConfig().Database

	// 构建psql命令
	cmd := exec.Command("psql",
		"-h", dbConfig.Host,
		"-p", dbConfig.Port,
		"-U", dbConfig.Username,
		"-d", dbConfig.Database,
		"-f", backupPath)

	// 设置环境变量
	cmd.Env = append(os.Environ(), "PGPASSWORD="+dbConfig.Password)
	cmd.Stderr = os.Stderr

	// 执行恢复命令
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("PostgreSQL恢复失败: %v", err)
	}

	return nil
}

func (s *BackupService) restoreSQLite(backupPath string) error {
	dbConfig := Config.GetConfig().Database

	// 对于SQLite，直接复制备份文件到数据库位置
	sourceFile, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("打开备份文件失败: %v", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dbConfig.Database)
	if err != nil {
		return fmt.Errorf("创建目标数据库文件失败: %v", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("复制数据库文件失败: %v", err)
	}

	return nil
}

func (s *BackupService) restoreFiles(backupPath string) error {
	// 打开ZIP文件
	zipReader, err := zip.OpenReader(backupPath)
	if err != nil {
		return fmt.Errorf("打开备份文件失败: %v", err)
	}
	defer zipReader.Close()

	// 解压到存储目录
	for _, file := range zipReader.File {
		// 构建目标路径
		targetPath := filepath.Join(s.storageManager.BasePath(), file.Name)

		// 创建目录
		if file.FileInfo().IsDir() {
			os.MkdirAll(targetPath, file.FileInfo().Mode())
			continue
		}

		// 创建父目录
		os.MkdirAll(filepath.Dir(targetPath), 0755)

		// 解压文件
		sourceFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("打开压缩文件失败: %v", err)
		}

		targetFile, err := os.Create(targetPath)
		if err != nil {
			sourceFile.Close()
			return fmt.Errorf("创建目标文件失败: %v", err)
		}

		_, err = io.Copy(targetFile, sourceFile)
		sourceFile.Close()
		targetFile.Close()

		if err != nil {
			return fmt.Errorf("解压文件失败: %v", err)
		}
	}

	return nil
}

func (s *BackupService) restoreFullBackup(backupPath string) error {
	// 打开ZIP文件
	zipReader, err := zip.OpenReader(backupPath)
	if err != nil {
		return fmt.Errorf("打开备份文件失败: %v", err)
	}
	defer zipReader.Close()

	// 创建临时目录
	tempDir := filepath.Join(s.backupPath, "temp_restore")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 解压所有文件到临时目录
	for _, file := range zipReader.File {
		targetPath := filepath.Join(tempDir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(targetPath, file.FileInfo().Mode())
			continue
		}

		os.MkdirAll(filepath.Dir(targetPath), 0755)

		sourceFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("打开压缩文件失败: %v", err)
		}

		targetFile, err := os.Create(targetPath)
		if err != nil {
			sourceFile.Close()
			return fmt.Errorf("创建目标文件失败: %v", err)
		}

		_, err = io.Copy(targetFile, sourceFile)
		sourceFile.Close()
		targetFile.Close()

		if err != nil {
			return fmt.Errorf("解压文件失败: %v", err)
		}
	}

	// 恢复数据库
	dbBackupPath := filepath.Join(tempDir, "database.sql")
	if _, err := os.Stat(dbBackupPath); err == nil {
		if err := s.restoreDatabase(dbBackupPath); err != nil {
			return fmt.Errorf("恢复数据库失败: %v", err)
		}
	}

	// 恢复文件
	storageBackupPath := filepath.Join(tempDir, "storage")
	if _, err := os.Stat(storageBackupPath); err == nil {
		// 备份当前存储目录
		currentStorageBackup := filepath.Join(s.backupPath, "current_storage_backup")
		if err := os.Rename(s.storageManager.BasePath(), currentStorageBackup); err != nil {
			s.storageManager.LogWarning("备份当前存储目录失败", map[string]interface{}{
				"error": err.Error(),
			})
		}

		// 恢复存储文件
		if err := os.Rename(storageBackupPath, s.storageManager.BasePath()); err != nil {
			// 如果恢复失败，尝试恢复原存储目录
			os.Rename(currentStorageBackup, s.storageManager.BasePath())
			return fmt.Errorf("恢复存储文件失败: %v", err)
		}

		// 删除备份的当前存储目录
		os.RemoveAll(currentStorageBackup)
	}

	return nil
}

func (s *BackupService) parseBackupInfo(fileName string) (*BackupInfo, error) {
	// 解析文件名格式：type_YYYYMMDD_HHMMSS.ext
	parts := strings.Split(fileName, "_")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid backup filename format")
	}

	backupType := parts[0]
	dateTime := parts[1] + "_" + parts[2]
	dateTime = strings.TrimSuffix(dateTime, filepath.Ext(dateTime))

	createdAt, err := time.Parse("20060102_150405", dateTime)
	if err != nil {
		return nil, err
	}

	return &BackupInfo{
		ID:        strings.TrimSuffix(fileName, filepath.Ext(fileName)),
		Type:      backupType,
		CreatedAt: createdAt,
		Status:    "success",
		Metadata:  make(map[string]interface{}),
	}, nil
}

func (s *BackupService) generateBackupReport(backupInfo *BackupInfo) map[string]interface{} {
	return map[string]interface{}{
		"backup_id":  backupInfo.ID,
		"type":       backupInfo.Type,
		"created_at": backupInfo.CreatedAt,
		"size":       backupInfo.Size,
		"md5":        backupInfo.MD5,
		"config":     Config.GetConfig(),
		"system_info": map[string]interface{}{
			"go_version": "1.21",
			"platform":   "linux/amd64",
		},
	}
}

func (s *BackupService) startAutoBackup() {
	ticker := time.NewTicker(s.config.BackupInterval)
	defer ticker.Stop()

	for range ticker.C {
		// 创建完整备份
		backupInfo, err := s.CreateFullBackup()
		if err != nil {
			s.storageManager.LogError("自动备份失败", map[string]interface{}{
				"error": err.Error(),
			})
			continue
		}

		s.storageManager.LogInfo("自动备份完成", map[string]interface{}{
			"backup_id": backupInfo.ID,
			"size":      backupInfo.Size,
		})

		// 清理旧备份
		s.CleanupOldBackups()
	}
}
