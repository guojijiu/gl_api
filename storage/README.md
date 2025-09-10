# Storage 存储系统

这是一个类似Laravel的存储系统，提供了文件存储、日志管理、缓存和临时文件管理等功能。

## 📋 目录

- [系统概览](#-系统概览)
- [目录结构](#-目录结构)
- [功能特性](#-功能特性)
- [API接口](#-api接口)
- [使用示例](#-使用示例)
- [配置说明](#-配置说明)
- [性能优化](#-性能优化)
- [安全注意事项](#-安全注意事项)
- [故障排除](#-故障排除)

## 🚀 系统概览

### 核心功能
- **文件存储管理** - 支持公共和私有文件存储
- **日志系统** - 结构化日志记录和管理
- **缓存系统** - 多级缓存策略
- **临时文件管理** - 临时文件创建和清理
- **SQL日志记录** - 数据库查询日志

### 技术特性
- **多存储后端** - 支持本地存储和云存储
- **安全防护** - 文件类型验证、大小限制、路径遍历防护
- **性能优化** - 缓存策略、异步处理、连接池
- **监控告警** - 存储使用统计、性能监控、异常告警

### 支持平台
- **Windows** - 完整支持
- **Linux** - 完整支持
- **macOS** - 完整支持
- **Docker** - 容器化支持
- **Kubernetes** - 集群部署支持

## 📁 目录结构

```
storage/
├── app/                    # 应用程序文件存储
│   ├── public/            # 公共文件（可直接访问）
│   │   ├── uploads/       # 用户上传文件
│   │   ├── images/        # 图片文件
│   │   ├── documents/     # 文档文件
│   │   └── media/         # 媒体文件
│   └── private/           # 私有文件（需要认证）
│       ├── user/          # 用户私有文件
│       ├── admin/         # 管理员文件
│       └── system/        # 系统文件
├── framework/             # 框架相关文件
│   ├── cache/             # 缓存文件
│   │   ├── data/          # 数据缓存
│   │   ├── views/         # 视图缓存
│   │   └── sessions/      # 会话缓存
│   ├── sessions/          # 会话文件
│   └── views/             # 视图缓存
├── logs/                  # 日志文件
│   ├── access/            # 访问日志
│   ├── audit/             # 审计日志
│   ├── business/          # 业务日志
│   ├── errors/            # 错误日志
│   ├── requests/          # 请求日志
│   ├── security/          # 安全日志
│   ├── sql/               # SQL日志
│   └── system/            # 系统日志
├── temp/                  # 临时文件
│   ├── uploads/           # 临时上传文件
│   ├── processing/        # 处理中文件
│   └── cleanup/           # 待清理文件
├── log_viewer.html        # 日志查看器
└── test_upload.html       # 文件上传测试页面
```

## 功能特性

### 1. 文件存储
- 支持公共和私有文件存储
- 自动生成唯一文件名
- 文件上传、下载、删除
- 文件存在性检查和大小获取
- 文件类型验证和大小限制

### 2. 日志管理
- 按日期分文件记录日志
- 支持不同日志级别（INFO、WARNING、ERROR、DEBUG）
- JSON格式日志，便于解析
- 支持上下文信息记录
- 日志查看器界面

### 3. 缓存系统
- 内存+文件双重缓存
- 支持TTL（生存时间）
- 自动清理过期缓存
- 线程安全的缓存操作
- 缓存统计和管理

### 4. 临时文件管理
- 创建临时文件
- 自动清理过期文件
- 支持文件扩展名
- 临时文件统计信息
- 定期清理机制

### 5. SQL日志记录
- 记录SQL查询语句
- 查询执行时间统计
- 慢查询检测
- SQL日志文件管理

## API接口

### 文件上传
```
POST /api/v1/storage/upload
Content-Type: multipart/form-data

参数:
- file: 上传的文件
- path: 存储路径（可选，默认uploads）
- type: 存储类型（public/private，默认public）
```

### 文件下载
```
GET /api/v1/storage/download/{path}?type={storage_type}

参数:
- path: 文件路径
- type: 存储类型（public/private）
```

### 获取日志
```
GET /api/v1/storage/logs?level={level}&date={date}

参数:
- level: 日志级别（可选）
- date: 日期（可选，默认今天）
```

### 获取存储信息
```
GET /api/v1/storage/info
```

### 清空缓存
```
POST /api/v1/storage/cache/clear
```

### 清理临时文件
```
POST /api/v1/storage/temp/clean
```

### 获取SQL日志
```
GET /api/v1/storage/sql-logs?date={date}

参数:
- date: 日期（可选，默认今天）
```

## 使用示例

### 在代码中使用Storage

```go
// 获取StorageManager实例
storageManager := app.StorageManager

// 记录日志
storageManager.LogInfo("用户登录成功", map[string]interface{}{
    "user_id": 123,
    "ip": "192.168.1.1",
})

// 设置缓存
storageManager.Cache("user:123", userData, 1*time.Hour)

// 获取缓存
if cached, err := storageManager.GetCache("user:123"); err == nil {
    // 使用缓存数据
}

// 创建临时文件
if tempFile, err := storageManager.CreateTempFile("upload"); err == nil {
    defer tempFile.Close()
    // 使用临时文件
}

// 记录SQL日志
storageManager.LogSQL("SELECT * FROM users WHERE id = ?", []interface{}{123}, 1.5)
```

### 文件上传示例

```html
<form action="/api/v1/storage/upload" method="post" enctype="multipart/form-data">
    <input type="file" name="file" required>
    <input type="text" name="path" placeholder="存储路径（可选）">
    <select name="type">
        <option value="public">公共文件</option>
        <option value="private">私有文件</option>
    </select>
    <button type="submit">上传</button>
</form>
```

### 使用curl上传文件

```bash
curl -X POST http://localhost:8080/api/v1/storage/upload \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@/path/to/file.jpg" \
  -F "path=uploads" \
  -F "type=public"
```

## 配置说明

### 环境变量配置

Storage系统支持以下环境变量配置：

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| STORAGE_UPLOAD_PATH | 上传路径 | ./storage/app/public |
| STORAGE_MAX_FILE_SIZE | 最大文件大小(MB) | 10 |
| STORAGE_ALLOWED_TYPES | 允许的文件类型 | jpg,jpeg,png,gif,pdf,doc,docx |
| STORAGE_PRIVATE_PATH | 私有文件路径 | ./storage/app/private |
| STORAGE_PUBLIC_PATH | 公共文件路径 | ./storage/app/public |
| STORAGE_TEMP_PATH | 临时文件路径 | ./storage/temp |
| STORAGE_LOG_PATH | 日志文件路径 | ./storage/logs |
| STORAGE_CACHE_PATH | 缓存文件路径 | ./storage/framework/cache |

### 文件类型限制

默认允许的文件类型：
- 图片：jpg, jpeg, png, gif
- 文档：pdf, doc, docx
- 其他：txt, csv, zip, rar

### 文件大小限制

- 默认最大文件大小：10MB
- 可通过环境变量 `STORAGE_MAX_FILE_SIZE` 调整

## 安全注意事项

1. **私有文件访问控制**: 私有文件需要用户认证才能访问
2. **文件类型验证**: 上传文件会进行类型验证
3. **文件大小限制**: 防止大文件上传攻击
4. **日志安全**: 日志文件包含敏感信息，需要适当的访问控制
5. **临时文件清理**: 临时文件会定期清理，避免磁盘空间浪费
6. **路径遍历防护**: 防止路径遍历攻击

## 📈 性能优化

### 缓存策略优化

#### 1. 多级缓存
```go
// 内存缓存 + 文件缓存 + Redis缓存
type CacheManager struct {
    memoryCache *MemoryCache
    fileCache   *FileCache
    redisCache  *RedisCache
}

func (c *CacheManager) Get(key string) (interface{}, error) {
    // 1. 尝试从内存缓存获取
    if value, exists := c.memoryCache.Get(key); exists {
        return value, nil
    }
    
    // 2. 尝试从文件缓存获取
    if value, err := c.fileCache.Get(key); err == nil {
        c.memoryCache.Set(key, value)
        return value, nil
    }
    
    // 3. 尝试从Redis缓存获取
    if value, err := c.redisCache.Get(key); err == nil {
        c.memoryCache.Set(key, value)
        c.fileCache.Set(key, value)
        return value, nil
    }
    
    return nil, ErrCacheMiss
}
```

#### 2. 缓存预热
```go
// 缓存预热策略
func (s *StorageService) WarmupCache() error {
    // 预热热门文件信息
    hotFiles, err := s.GetHotFiles()
    if err != nil {
        return err
    }
    
    for _, file := range hotFiles {
        s.Cache.Set(fmt.Sprintf("file:%s", file.Path), file, time.Hour)
    }
    
    // 预热目录结构
    dirs, err := s.GetDirectories()
    if err != nil {
        return err
    }
    
    for _, dir := range dirs {
        s.Cache.Set(fmt.Sprintf("dir:%s", dir.Path), dir, time.Hour)
    }
    
    return nil
}
```

#### 3. 缓存清理策略
```go
// 智能缓存清理
func (c *CacheManager) Cleanup() error {
    // 清理过期缓存
    c.memoryCache.CleanupExpired()
    c.fileCache.CleanupExpired()
    
    // 清理LRU缓存
    c.memoryCache.CleanupLRU(1000) // 保留1000个最常用的缓存项
    
    // 清理大文件缓存
    c.fileCache.CleanupLargeFiles(100 * 1024 * 1024) // 清理大于100MB的文件
    
    return nil
}
```

### 文件存储优化

#### 1. 文件分片存储
```go
// 大文件分片存储
func (s *StorageService) StoreLargeFile(file *File) error {
    if file.Size > s.config.MaxFileSize {
        return s.storeFileInChunks(file)
    }
    return s.storeFile(file)
}

func (s *StorageService) storeFileInChunks(file *File) error {
    chunkSize := s.config.ChunkSize
    chunks := int(math.Ceil(float64(file.Size) / float64(chunkSize)))
    
    for i := 0; i < chunks; i++ {
        start := i * chunkSize
        end := start + chunkSize
        if end > file.Size {
            end = file.Size
        }
        
        chunk := &FileChunk{
            FileID: file.ID,
            Index:  i,
            Data:   file.Data[start:end],
        }
        
        if err := s.storeChunk(chunk); err != nil {
            return err
        }
    }
    
    return nil
}
```

#### 2. 异步文件处理
```go
// 异步文件处理
func (s *StorageService) ProcessFileAsync(file *File) error {
    go func() {
        // 文件压缩
        if s.config.EnableCompression {
            s.compressFile(file)
        }
        
        // 生成缩略图
        if s.isImageFile(file) {
            s.generateThumbnails(file)
        }
        
        // 病毒扫描
        if s.config.EnableVirusScan {
            s.scanFile(file)
        }
        
        // 更新文件状态
        s.updateFileStatus(file.ID, "processed")
    }()
    
    return nil
}
```

#### 3. 文件去重
```go
// 文件去重
func (s *StorageService) DeduplicateFile(file *File) (*File, error) {
    // 计算文件哈希
    hash, err := s.calculateFileHash(file)
    if err != nil {
        return nil, err
    }
    
    // 检查是否已存在
    existingFile, err := s.GetFileByHash(hash)
    if err == nil {
        // 创建硬链接
        if err := s.createHardLink(existingFile.Path, file.Path); err != nil {
            return nil, err
        }
        return existingFile, nil
    }
    
    // 存储新文件
    return s.storeFile(file)
}
```

### 日志优化

#### 1. 异步日志写入
```go
// 异步日志写入
type AsyncLogger struct {
    logChan chan LogEntry
    workers int
}

func (l *AsyncLogger) Log(entry LogEntry) {
    select {
    case l.logChan <- entry:
    default:
        // 日志队列满，丢弃日志
        log.Printf("Log queue full, dropping log entry")
    }
}

func (l *AsyncLogger) startWorkers() {
    for i := 0; i < l.workers; i++ {
        go l.worker()
    }
}

func (l *AsyncLogger) worker() {
    for entry := range l.logChan {
        l.writeLog(entry)
    }
}
```

#### 2. 日志压缩
```go
// 日志压缩
func (l *Logger) compressLogFile(filePath string) error {
    // 压缩日志文件
    compressedPath := filePath + ".gz"
    
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()
    
    compressedFile, err := os.Create(compressedPath)
    if err != nil {
        return err
    }
    defer compressedFile.Close()
    
    gzWriter := gzip.NewWriter(compressedFile)
    defer gzWriter.Close()
    
    _, err = io.Copy(gzWriter, file)
    if err != nil {
        return err
    }
    
    // 删除原文件
    return os.Remove(filePath)
}
```

#### 3. 日志轮转
```go
// 日志轮转
func (l *Logger) rotateLogs() error {
    for _, logType := range l.logTypes {
        filePath := l.getLogFilePath(logType)
        
        // 检查文件大小
        if l.shouldRotate(filePath) {
            // 重命名当前日志文件
            rotatedPath := l.getRotatedFilePath(filePath)
            if err := os.Rename(filePath, rotatedPath); err != nil {
                return err
            }
            
            // 压缩旧日志文件
            if l.config.EnableCompression {
                l.compressLogFile(rotatedPath)
            }
            
            // 清理旧日志文件
            l.cleanupOldLogs(logType)
        }
    }
    
    return nil
}
```

### 数据库优化

#### 1. 连接池优化
```go
// 数据库连接池配置
func (s *StorageService) setupDatabase() error {
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        ConnPool: &sql.DB{
            MaxOpenConns:    100,
            MaxIdleConns:    10,
            ConnMaxLifetime: time.Hour,
            ConnMaxIdleTime: time.Minute * 30,
        },
    })
    
    if err != nil {
        return err
    }
    
    s.db = db
    return nil
}
```

#### 2. 查询优化
```go
// 查询优化
func (s *StorageService) GetFilesByUser(userID uint, limit, offset int) ([]File, error) {
    var files []File
    
    // 使用索引优化查询
    err := s.db.Where("user_id = ?", userID).
        Order("created_at DESC").
        Limit(limit).
        Offset(offset).
        Find(&files).Error
    
    return files, err
}

// 预加载关联数据
func (s *StorageService) GetFileWithDetails(fileID uint) (*File, error) {
    var file File
    
    err := s.db.Preload("User").
        Preload("Tags").
        First(&file, fileID).Error
    
    return &file, err
}
```

### 监控和告警

#### 1. 性能监控
```go
// 性能监控
type PerformanceMonitor struct {
    metrics map[string]float64
    mutex   sync.RWMutex
}

func (p *PerformanceMonitor) RecordMetric(name string, value float64) {
    p.mutex.Lock()
    defer p.mutex.Unlock()
    
    p.metrics[name] = value
}

func (p *PerformanceMonitor) GetMetrics() map[string]float64 {
    p.mutex.RLock()
    defer p.mutex.RUnlock()
    
    result := make(map[string]float64)
    for k, v := range p.metrics {
        result[k] = v
    }
    
    return result
}
```

#### 2. 告警系统
```go
// 告警系统
func (s *StorageService) checkAlerts() {
    // 检查磁盘空间
    if s.getDiskUsage() > 90 {
        s.sendAlert("磁盘空间不足", "磁盘使用率超过90%")
    }
    
    // 检查文件数量
    if s.getFileCount() > 1000000 {
        s.sendAlert("文件数量过多", "文件数量超过100万")
    }
    
    // 检查错误率
    if s.getErrorRate() > 5 {
        s.sendAlert("错误率过高", "错误率超过5%")
    }
}
```

### 性能测试

#### 1. 基准测试
```go
// 基准测试
func BenchmarkFileUpload(b *testing.B) {
    service := NewStorageService()
    file := &File{
        Name: "test.jpg",
        Size: 1024 * 1024, // 1MB
        Data: make([]byte, 1024*1024),
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.StoreFile(file)
    }
}

func BenchmarkFileDownload(b *testing.B) {
    service := NewStorageService()
    fileID := uint(1)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.GetFile(fileID)
    }
}
```

#### 2. 压力测试
```go
// 压力测试
func TestConcurrentFileUpload(t *testing.T) {
    service := NewStorageService()
    concurrency := 100
    
    var wg sync.WaitGroup
    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func(i int) {
            defer wg.Done()
            
            file := &File{
                Name: fmt.Sprintf("test_%d.jpg", i),
                Size: 1024,
                Data: make([]byte, 1024),
            }
            
            err := service.StoreFile(file)
            assert.NoError(t, err)
        }(i)
    }
    
    wg.Wait()
}
```

## 🔧 扩展功能

### 云存储集成

#### 1. AWS S3 集成
```go
// AWS S3 存储服务
type S3StorageService struct {
    client *s3.Client
    bucket string
}

func (s *S3StorageService) StoreFile(file *File) error {
    _, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(file.Path),
        Body:   bytes.NewReader(file.Data),
    })
    return err
}
```

#### 2. 阿里云 OSS 集成
```go
// 阿里云 OSS 存储服务
type OSSStorageService struct {
    client *oss.Client
    bucket *oss.Bucket
}

func (s *OSSStorageService) StoreFile(file *File) error {
    return s.bucket.PutObject(file.Path, bytes.NewReader(file.Data))
}
```

### 文件处理服务

#### 1. 图片处理
```go
// 图片处理服务
type ImageProcessor struct {
    config ImageConfig
}

func (p *ImageProcessor) ProcessImage(file *File) error {
    // 生成缩略图
    thumbnails := []ThumbnailSize{
        {Width: 150, Height: 150, Name: "small"},
        {Width: 300, Height: 300, Name: "medium"},
        {Width: 600, Height: 600, Name: "large"},
    }
    
    for _, size := range thumbnails {
        if err := p.generateThumbnail(file, size); err != nil {
            return err
        }
    }
    
    return nil
}
```

#### 2. 文档处理
```go
// 文档处理服务
type DocumentProcessor struct {
    config DocumentConfig
}

func (p *DocumentProcessor) ProcessDocument(file *File) error {
    // 提取文本内容
    text, err := p.extractText(file)
    if err != nil {
        return err
    }
    
    // 生成预览
    if err := p.generatePreview(file); err != nil {
        return err
    }
    
    return nil
}
```

### 搜索服务

#### 1. 全文搜索
```go
// 全文搜索服务
type SearchService struct {
    indexer *bleve.Index
}

func (s *SearchService) IndexFile(file *File) error {
    doc := map[string]interface{}{
        "id":       file.ID,
        "name":     file.Name,
        "content":  file.Content,
        "tags":     file.Tags,
        "category": file.Category,
        "created":  file.CreatedAt,
    }
    
    return s.indexer.Index(file.ID, doc)
}
```

### 备份和恢复

#### 1. 自动备份
```go
// 自动备份服务
type BackupService struct {
    config BackupConfig
    storage StorageService
}

func (s *BackupService) CreateBackup() error {
    // 创建备份目录
    backupDir := s.getBackupDirectory()
    if err := os.MkdirAll(backupDir, 0755); err != nil {
        return err
    }
    
    // 备份文件
    if err := s.backupFiles(backupDir); err != nil {
        return err
    }
    
    return nil
}
```

### 监控和统计

#### 1. 使用统计
```go
// 使用统计服务
type UsageStatsService struct {
    db *gorm.DB
}

func (s *UsageStatsService) GetFileStats() (*FileStats, error) {
    var stats FileStats
    
    // 总文件数
    if err := s.db.Model(&File{}).Count(&stats.TotalFiles).Error; err != nil {
        return nil, err
    }
    
    // 总存储大小
    if err := s.db.Model(&File{}).Select("SUM(size)").Scan(&stats.TotalSize).Error; err != nil {
        return nil, err
    }
    
    return &stats, nil
}
```

### 安全增强

#### 1. 文件病毒扫描
```go
// 文件病毒扫描服务
type VirusScanService struct {
    scanner *clamav.Scanner
}

func (s *VirusScanService) ScanFile(file *File) (*ScanResult, error) {
    result, err := s.scanner.ScanBytes(file.Data)
    if err != nil {
        return nil, err
    }
    
    return &ScanResult{
        Clean:    result.Status == clamav.StatusClean,
        Virus:    result.Virus,
        Engine:   result.Engine,
        Scanned:  time.Now(),
    }, nil
}
```

### 使用示例

#### 1. 基本使用
```go
// 创建存储服务
storage := NewStorageService(StorageConfig{
    BasePath: "./storage",
    MaxFileSize: 100 * 1024 * 1024, // 100MB
    AllowedExtensions: []string{".jpg", ".png", ".pdf"},
})

// 存储文件
file := &File{
    Name: "document.pdf",
    Data: fileData,
    Size: int64(len(fileData)),
    UserID: 1,
}

storedFile, err := storage.StoreFile(file)
if err != nil {
    log.Fatal(err)
}
```

### 最佳实践

1. **文件命名**: 使用UUID或时间戳避免文件名冲突
2. **目录结构**: 按日期或用户ID分目录存储
3. **缓存策略**: 合理设置缓存TTL和大小限制
4. **备份策略**: 定期备份重要文件
5. **监控告警**: 设置磁盘空间和错误率告警
6. **安全防护**: 启用文件类型验证和病毒扫描
7. **性能优化**: 使用异步处理和连接池
8. **日志记录**: 记录所有文件操作和错误信息

### 云存储支持
- AWS S3
- 阿里云OSS
- 腾讯云COS
- 七牛云存储

### 图片处理
- 缩略图生成
- 图片压缩
- 水印添加
- 格式转换

### 文件处理
- 文件压缩
- 文件加密
- 文件预览
- 批量处理

### 监控和统计
- 存储使用统计
- 文件访问统计
- 性能监控
- 异常告警

## 故障排除

### 常见问题

1. **文件上传失败**
   - 检查文件大小是否超限
   - 检查文件类型是否允许
   - 检查存储目录权限

2. **缓存不生效**
   - 检查缓存目录权限
   - 检查TTL设置
   - 检查缓存键名

3. **日志文件过大**
   - 启用日志轮转
   - 定期清理旧日志
   - 调整日志级别

4. **临时文件堆积**
   - 检查清理任务是否正常运行
   - 手动清理临时文件
   - 调整清理策略

### 调试方法

1. **查看存储信息**
```bash
curl http://localhost:8080/api/v1/storage/info
```

2. **查看日志**
```bash
curl http://localhost:8080/api/v1/storage/logs
```

3. **清空缓存**
```bash
curl -X POST http://localhost:8080/api/v1/storage/cache/clear
```

4. **清理临时文件**
```bash
curl -X POST http://localhost:8080/api/v1/storage/temp/clean
```
