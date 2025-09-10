package Services

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Models"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// SecurityService 安全防护服务
type SecurityService struct {
	db              *gorm.DB
	config          *Config.SecurityConfig
	commonPasswords map[string]bool
	phishingURLs    map[string]bool
	threatIPs       map[string]bool
	threatDetection *ThreatDetectionService
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewSecurityService 创建安全防护服务
func NewSecurityService(db *gorm.DB, config *Config.SecurityConfig) *SecurityService {
	ctx, cancel := context.WithCancel(context.Background())

	// 如果配置为空，使用默认配置
	if config == nil {
		config = &Config.SecurityConfig{}
		config.SetDefaults()
	}

	service := &SecurityService{
		db:              db,
		config:          config,
		commonPasswords: make(map[string]bool),
		phishingURLs:    make(map[string]bool),
		threatIPs:       make(map[string]bool),
		ctx:             ctx,
		cancel:          cancel,
	}

	// 初始化威胁检测服务
	service.threatDetection = NewThreatDetectionService(db, config)

	// 初始化服务
	service.initialize()

	return service
}

// GetDB 获取数据库连接
func (s *SecurityService) GetDB() *gorm.DB {
	return s.db
}

// initialize 初始化服务
func (s *SecurityService) initialize() {
	// 加载常见密码
	s.loadCommonPasswords()

	// 加载钓鱼URL
	s.loadPhishingURLs()

	// 加载威胁情报
	s.loadThreatIntelligence()

	// 启动定期更新任务
	go s.startPeriodicUpdates()
}

// loadCommonPasswords 加载常见密码
func (s *SecurityService) loadCommonPasswords() {
	if !s.config.PasswordPolicy.PreventCommonPasswords {
		return
	}

	// 如果配置文件路径为空，使用内置列表
	if s.config.PasswordPolicy.CommonPasswordsFile == "" {
		commonPasswords := []string{
			"password", "123456", "123456789", "qwerty", "abc123", "password123",
			"admin", "root", "user", "guest", "test", "demo", "welcome",
			"letmein", "login", "pass", "secret", "password1", "12345678",
		}

		s.mu.Lock()
		for _, pwd := range commonPasswords {
			s.commonPasswords[pwd] = true
		}
		s.mu.Unlock()
		return
	}

	content, err := ioutil.ReadFile(s.config.PasswordPolicy.CommonPasswordsFile)
	if err != nil {
		// 使用内置的常见密码列表
		commonPasswords := []string{
			"password", "123456", "123456789", "qwerty", "abc123", "password123",
			"admin", "root", "user", "guest", "test", "demo", "welcome",
			"letmein", "login", "pass", "secret", "password1", "12345678",
		}

		s.mu.Lock()
		for _, pwd := range commonPasswords {
			s.commonPasswords[pwd] = true
		}
		s.mu.Unlock()
		return
	}

	passwords := strings.Split(string(content), "\n")
	s.mu.Lock()
	for _, pwd := range passwords {
		pwd = strings.TrimSpace(pwd)
		if pwd != "" {
			s.commonPasswords[pwd] = true
		}
	}
	s.mu.Unlock()
}

// loadPhishingURLs 加载钓鱼URL
func (s *SecurityService) loadPhishingURLs() {
	if !s.config.ThreatProtection.PhishingProtection {
		return
	}

	// 如果配置文件路径为空，跳过加载
	if s.config.ThreatProtection.PhishingURLsFile == "" {
		return
	}

	content, err := ioutil.ReadFile(s.config.ThreatProtection.PhishingURLsFile)
	if err != nil {
		return
	}

	urls := strings.Split(string(content), "\n")
	s.mu.Lock()
	for _, url := range urls {
		url = strings.TrimSpace(url)
		if url != "" {
			s.phishingURLs[url] = true
		}
	}
	s.mu.Unlock()
}

// loadThreatIntelligence 加载威胁情报
func (s *SecurityService) loadThreatIntelligence() {
	if !s.config.ThreatProtection.ThreatIntelligence {
		return
	}

	// 从数据库加载威胁情报
	var threats []Models.ThreatIntelligence
	s.db.Where("active = ?", true).Find(&threats)

	s.mu.Lock()
	for _, threat := range threats {
		if threat.IPAddress != "" {
			s.threatIPs[threat.IPAddress] = true
		}
	}
	s.mu.Unlock()
}

// startPeriodicUpdates 启动定期更新任务
func (s *SecurityService) startPeriodicUpdates() {
	ticker := time.NewTicker(s.config.ThreatProtection.TIUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.updateThreatIntelligence()
		}
	}
}

// updateThreatIntelligence 更新威胁情报
func (s *SecurityService) updateThreatIntelligence() {
	if !s.config.ThreatProtection.ThreatIntelligence {
		return
	}

	for _, url := range s.config.ThreatProtection.TISourceURLs {
		go s.fetchThreatIntelligence(url)
	}
}

// fetchThreatIntelligence 获取威胁情报
func (s *SecurityService) fetchThreatIntelligence(url string) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// 解析威胁情报数据
	s.parseThreatIntelligence(body)
}

// parseThreatIntelligence 解析威胁情报
func (s *SecurityService) parseThreatIntelligence(data []byte) {
	// 这里应该根据具体的威胁情报源格式来解析
	// 示例实现
	lines := strings.Split(string(data), "\n")

	s.mu.Lock()
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			s.threatIPs[line] = true
		}
	}
	s.mu.Unlock()
}

// ValidatePassword 验证密码强度
func (s *SecurityService) ValidatePassword(password, username string) (bool, []string) {
	var errors []string

	// 检查长度
	if len(password) < s.config.PasswordPolicy.MinLength {
		errors = append(errors, fmt.Sprintf("密码长度不能少于%d个字符", s.config.PasswordPolicy.MinLength))
	}
	if len(password) > s.config.PasswordPolicy.MaxLength {
		errors = append(errors, fmt.Sprintf("密码长度不能超过%d个字符", s.config.PasswordPolicy.MaxLength))
	}

	// 检查字符类型
	if s.config.PasswordPolicy.RequireUppercase && !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		errors = append(errors, "密码必须包含大写字母")
	}
	if s.config.PasswordPolicy.RequireLowercase && !regexp.MustCompile(`[a-z]`).MatchString(password) {
		errors = append(errors, "密码必须包含小写字母")
	}
	if s.config.PasswordPolicy.RequireNumbers && !regexp.MustCompile(`[0-9]`).MatchString(password) {
		errors = append(errors, "密码必须包含数字")
	}
	if s.config.PasswordPolicy.RequireSpecialChars {
		specialChars := regexp.QuoteMeta(s.config.PasswordPolicy.SpecialCharsList)
		if !regexp.MustCompile(`[` + specialChars + `]`).MatchString(password) {
			errors = append(errors, "密码必须包含特殊字符")
		}
	}

	// 检查常见密码
	if s.config.PasswordPolicy.PreventCommonPasswords {
		s.mu.RLock()
		if s.commonPasswords[strings.ToLower(password)] {
			errors = append(errors, "密码不能使用常见密码")
		}
		s.mu.RUnlock()
	}

	// 检查用户名
	if s.config.PasswordPolicy.PreventUsernameInPassword && username != "" {
		if strings.Contains(strings.ToLower(password), strings.ToLower(username)) {
			errors = append(errors, "密码不能包含用户名")
		}
	}

	// 检查连续字符
	if s.config.PasswordPolicy.PreventSequentialChars {
		if s.hasSequentialChars(password) {
			errors = append(errors, "密码不能包含连续字符")
		}
	}

	// 检查重复字符
	if s.config.PasswordPolicy.PreventRepeatedChars {
		if s.hasRepeatedChars(password, s.config.PasswordPolicy.MaxRepeatedChars) {
			errors = append(errors, fmt.Sprintf("密码不能包含超过%d个连续重复字符", s.config.PasswordPolicy.MaxRepeatedChars))
		}
	}

	// 计算密码强度
	strength := s.calculatePasswordStrength(password)
	if strength < float64(s.config.PasswordPolicy.PasswordStrengthThreshold) {
		errors = append(errors, fmt.Sprintf("密码强度不足，当前强度: %.1f%%，要求: %.1f%%", strength, float64(s.config.PasswordPolicy.PasswordStrengthThreshold)))
	}

	return len(errors) == 0, errors
}

// hasSequentialChars 检查是否有连续字符
func (s *SecurityService) hasSequentialChars(password string) bool {
	for i := 0; i < len(password)-2; i++ {
		if password[i+1] == password[i]+1 && password[i+2] == password[i]+2 {
			return true
		}
	}
	return false
}

// hasRepeatedChars 检查是否有重复字符
func (s *SecurityService) hasRepeatedChars(password string, maxRepeated int) bool {
	count := 1
	for i := 1; i < len(password); i++ {
		if password[i] == password[i-1] {
			count++
			if count > maxRepeated {
				return true
			}
		} else {
			count = 1
		}
	}
	return false
}

// calculatePasswordStrength 计算密码强度
func (s *SecurityService) calculatePasswordStrength(password string) float64 {
	score := 0.0

	// 长度分数
	length := len(password)
	if length >= 8 {
		score += 20
	} else if length >= 6 {
		score += 10
	}

	// 字符类型分数
	if regexp.MustCompile(`[a-z]`).MatchString(password) {
		score += 10
	}
	if regexp.MustCompile(`[A-Z]`).MatchString(password) {
		score += 10
	}
	if regexp.MustCompile(`[0-9]`).MatchString(password) {
		score += 10
	}
	if regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password) {
		score += 15
	}

	// 复杂度分数
	uniqueChars := make(map[rune]bool)
	for _, char := range password {
		uniqueChars[char] = true
	}
	score += float64(len(uniqueChars)) * 2

	// 熵分数
	entropy := 0.0
	charSetSize := 0
	if regexp.MustCompile(`[a-z]`).MatchString(password) {
		charSetSize += 26
	}
	if regexp.MustCompile(`[A-Z]`).MatchString(password) {
		charSetSize += 26
	}
	if regexp.MustCompile(`[0-9]`).MatchString(password) {
		charSetSize += 10
	}
	if regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password) {
		charSetSize += 32
	}

	if charSetSize > 0 {
		entropy = float64(length) * float64(charSetSize) / 100
		score += entropy
	}

	// 限制最大分数
	if score > 100 {
		score = 100
	}

	return score
}

// CheckPasswordHistory 检查密码历史
func (s *SecurityService) CheckPasswordHistory(userID uint, passwordHash string) bool {
	if s.config.BaseSecurity.PasswordHistoryCount <= 0 {
		return true
	}

	var count int64
	s.db.Model(&Models.PasswordHistory{}).
		Where("user_id = ? AND password_hash = ?", userID, passwordHash).
		Count(&count)

	return count == 0
}

// RecordPasswordChange 记录密码更改
func (s *SecurityService) RecordPasswordChange(userID, changedBy uint, passwordHash, reason, ipAddress, userAgent string) error {
	history := Models.PasswordHistory{
		UserID:       userID,
		PasswordHash: passwordHash,
		ChangedAt:    time.Now(),
		ChangedBy:    changedBy,
		Reason:       reason,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	// 保存密码历史
	if err := s.db.Create(&history).Error; err != nil {
		return err
	}

	// 清理旧的密码历史
	if s.config.BaseSecurity.PasswordHistoryCount > 0 {
		var histories []Models.PasswordHistory
		s.db.Where("user_id = ?", userID).
			Order("changed_at DESC").
			Limit(s.config.BaseSecurity.PasswordHistoryCount + 1).
			Find(&histories)

		if len(histories) > s.config.BaseSecurity.PasswordHistoryCount {
			oldHistories := histories[s.config.BaseSecurity.PasswordHistoryCount:]
			for _, old := range oldHistories {
				s.db.Delete(&old)
			}
		}
	}

	return nil
}

// CheckLoginAttempts 检查登录尝试
func (s *SecurityService) CheckLoginAttempts(username, ipAddress string) (bool, string) {
	// 检查账户锁定
	var lockout Models.AccountLockout
	if err := s.db.Where("username = ? AND active = ? AND expiry_time > ?", username, true, time.Now()).First(&lockout).Error; err == nil {
		return false, "账户已被锁定"
	}

	// 检查IP锁定
	if err := s.db.Where("ip_address = ? AND active = ? AND expiry_time > ?", ipAddress, true, time.Now()).First(&lockout).Error; err == nil {
		return false, "IP地址已被锁定"
	}

	// 检查登录尝试次数
	var attemptCount int64
	s.db.Model(&Models.LoginAttempt{}).
		Where("username = ? AND ip_address = ? AND success = ? AND attempt_time > ?",
			username, ipAddress, false, time.Now().Add(-s.config.BaseSecurity.LoginLockoutDuration)).
		Count(&attemptCount)

	if int(attemptCount) >= s.config.BaseSecurity.MaxLoginAttempts {
		// 创建锁定记录
		lockout = Models.AccountLockout{
			UserID:       0, // 未知用户ID
			Username:     username,
			IPAddress:    ipAddress,
			LockoutType:  "login_attempts",
			Reason:       "登录尝试次数过多",
			LockoutTime:  time.Now(),
			ExpiryTime:   time.Now().Add(s.config.BaseSecurity.LoginLockoutDuration),
			AttemptCount: int(attemptCount),
			Active:       true,
		}
		s.db.Create(&lockout)

		return false, "登录尝试次数过多，账户已被锁定"
	}

	return true, ""
}

// RecordLoginAttempt 记录登录尝试
func (s *SecurityService) RecordLoginAttempt(username, ipAddress, userAgent, failureReason string, success bool, location, deviceInfo string) error {
	attempt := Models.LoginAttempt{
		Username:      username,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		Success:       success,
		FailureReason: failureReason,
		AttemptTime:   time.Now(),
		Location:      location,
		DeviceInfo:    deviceInfo,
		RiskScore:     s.calculateLoginRiskScore(username, ipAddress, userAgent),
		Blocked:       false,
	}

	return s.db.Create(&attempt).Error
}

// calculateLoginRiskScore 计算登录风险评分
func (s *SecurityService) calculateLoginRiskScore(username, ipAddress, userAgent string) float64 {
	score := 0.0

	// 检查威胁IP
	s.mu.RLock()
	if s.threatIPs[ipAddress] {
		score += 50
	}
	s.mu.RUnlock()

	// 检查可疑用户代理
	suspiciousUserAgents := []string{
		"bot", "crawler", "spider", "scraper", "curl", "wget", "python", "java",
	}
	for _, suspicious := range suspiciousUserAgents {
		if strings.Contains(strings.ToLower(userAgent), suspicious) {
			score += 20
			break
		}
	}

	// 检查登录模式
	var recentAttempts int64
	s.db.Model(&Models.LoginAttempt{}).
		Where("ip_address = ? AND attempt_time > ?", ipAddress, time.Now().Add(-time.Hour)).
		Count(&recentAttempts)

	if recentAttempts > 10 {
		score += 30
	}

	// 检查地理位置异常
	// 这里可以集成地理位置服务来检测异常登录位置

	return score
}

// CheckAccessControl 检查访问控制
func (s *SecurityService) CheckAccessControl(userID uint, resource, action string) (bool, string) {
	var controls []Models.AccessControl
	s.db.Where("user_id = ? AND resource = ? AND action = ? AND active = ?",
		userID, resource, action, true).
		Order("priority DESC").
		Find(&controls)

	if len(controls) == 0 {
		// 默认策略
		if s.config.AccessControl.DefaultDenyPolicy {
			return false, "访问被拒绝"
		}
		return true, ""
	}

	// 按优先级排序
	sort.Slice(controls, func(i, j int) bool {
		return controls[i].Priority > controls[j].Priority
	})

	// 检查第一个匹配的控制规则
	control := controls[0]

	// 检查时间限制
	if control.TimeRestriction != "" {
		if !s.checkTimeRestriction(control.TimeRestriction) {
			return false, "访问时间受限"
		}
	}

	// 检查位置限制
	if control.LocationRestriction != "" {
		// 这里可以集成地理位置检查
	}

	// 检查设备限制
	if control.DeviceRestriction != "" {
		// 这里可以集成设备指纹检查
	}

	return control.Permission == "allow", control.Permission
}

// checkTimeRestriction 检查时间限制
func (s *SecurityService) checkTimeRestriction(restriction string) bool {
	// 示例：检查工作时间限制
	now := time.Now()
	weekday := now.Weekday()
	hour := now.Hour()

	// 简单的工作时间检查（周一到周五，9:00-18:00）
	if weekday >= time.Monday && weekday <= time.Friday && hour >= 9 && hour < 18 {
		return true
	}

	return false
}

// DetectAnomaly 异常检测
func (s *SecurityService) DetectAnomaly(userID uint, eventType, resource, action, ipAddress, userAgent string) (bool, float64) {
	if !s.config.AnomalyDetection.Enabled {
		return false, 0
	}

	score := 0.0

	// 行为分析
	if s.config.AnomalyDetection.BehavioralAnalysis {
		behaviorScore := s.analyzeBehavior(userID, eventType, resource, action)
		score += behaviorScore
	}

	// 模式识别
	if s.config.AnomalyDetection.PatternRecognition {
		patternScore := s.analyzePattern(userID, eventType, resource, action, ipAddress)
		score += patternScore
	}

	// 检查是否超过阈值
	isAnomaly := score > s.config.AnomalyDetection.AnomalyScoreThreshold

	// 记录安全事件
	s.RecordSecurityEvent(userID, eventType, "medium", ipAddress, userAgent, resource, action, "", score, score, false, false, "", "")

	return isAnomaly, score
}

// analyzeBehavior 行为分析
func (s *SecurityService) analyzeBehavior(userID uint, eventType, resource, action string) float64 {
	score := 0.0

	// 检查用户历史行为
	var eventCount int64
	s.db.Model(&Models.SecurityEvent{}).
		Where("user_id = ? AND event_type = ? AND created_at > ?",
			userID, eventType, time.Now().Add(-24*time.Hour)).
		Count(&eventCount)

	// 如果事件频率异常高
	if eventCount > 100 {
		score += 30
	}

	// 检查资源访问模式
	var resourceCount int64
	s.db.Model(&Models.SecurityEvent{}).
		Where("user_id = ? AND resource = ? AND created_at > ?",
			userID, resource, time.Now().Add(-time.Hour)).
		Count(&resourceCount)

	if resourceCount > 50 {
		score += 20
	}

	return score
}

// analyzePattern 模式识别
func (s *SecurityService) analyzePattern(userID uint, eventType, resource, action, ipAddress string) float64 {
	score := 0.0

	// 检查IP地址变化
	var recentIPs []string
	s.db.Model(&Models.SecurityEvent{}).
		Where("user_id = ? AND created_at > ?", userID, time.Now().Add(-time.Hour)).
		Pluck("ip_address", &recentIPs)

	uniqueIPs := make(map[string]bool)
	for _, ip := range recentIPs {
		uniqueIPs[ip] = true
	}

	if len(uniqueIPs) > 3 {
		score += 25
	}

	// 检查时间模式
	var recentEvents []Models.SecurityEvent
	s.db.Where("user_id = ? AND created_at > ?", userID, time.Now().Add(-time.Hour)).
		Find(&recentEvents)

	if len(recentEvents) > 0 {
		// 检查事件间隔是否异常
		for i := 1; i < len(recentEvents); i++ {
			interval := recentEvents[i].CreatedAt.Sub(recentEvents[i-1].CreatedAt)
			if interval < time.Second {
				score += 15
			}
		}
	}

	return score
}

// RecordSecurityEvent 记录安全事件
func (s *SecurityService) RecordSecurityEvent(userID uint, eventType, eventLevel, ipAddress, userAgent, resource, action, details string, riskScore, anomalyScore float64, blocked, alerted bool, location, deviceInfo string) error {
	event := Models.SecurityEvent{
		EventType:    eventType,
		EventLevel:   eventLevel,
		UserID:       &userID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Resource:     resource,
		Action:       action,
		Details:      details,
		RiskScore:    riskScore,
		AnomalyScore: anomalyScore,
		Blocked:      blocked,
		Alerted:      alerted,
		Location:     location,
		DeviceInfo:   deviceInfo,
	}

	return s.db.Create(&event).Error
}

// CheckThreatProtection 威胁防护检查
func (s *SecurityService) CheckThreatProtection(ipAddress, url, fileHash string) (bool, string) {
	// 使用威胁检测服务检查威胁IP
	if s.threatDetection != nil {
		if isThreat, threatInfo := s.threatDetection.CheckThreatIP(ipAddress); isThreat {
			return false, fmt.Sprintf("IP地址被识别为威胁源: %s (严重程度: %s)", threatInfo.Description, threatInfo.Severity)
		}
	} else {
		// 回退到原有检查
		s.mu.RLock()
		if s.threatIPs[ipAddress] {
			s.mu.RUnlock()
			return false, "IP地址被识别为威胁源"
		}
		s.mu.RUnlock()
	}

	// 使用威胁检测服务检查钓鱼URL
	if url != "" && s.config.ThreatProtection.PhishingProtection {
		if s.threatDetection != nil {
			if isPhishing, phishingInfo := s.threatDetection.CheckPhishingURL(url); isPhishing {
				return false, fmt.Sprintf("URL被识别为钓鱼网站: %s (严重程度: %s)", phishingInfo.Description, phishingInfo.Severity)
			}
		} else {
			// 回退到原有检查
			s.mu.RLock()
			if s.phishingURLs[url] {
				s.mu.RUnlock()
				return false, "URL被识别为钓鱼网站"
			}
			s.mu.RUnlock()
		}
	}

	// 使用威胁检测服务检查恶意文件
	if fileHash != "" && s.config.ThreatProtection.MalwareScanning {
		if s.threatDetection != nil {
			if isMalware, malwareInfo := s.threatDetection.CheckMalwareHash(fileHash); isMalware {
				return false, fmt.Sprintf("文件被识别为恶意软件: %s (严重程度: %s)", malwareInfo.Description, malwareInfo.Severity)
			}
		} else {
			// 回退到原有检查
			if s.isMalwareFile(fileHash) {
				return false, "文件被识别为恶意软件"
			}
		}
	}

	return true, ""
}

// isMalwareFile 检查是否为恶意文件
func (s *SecurityService) isMalwareFile(fileHash string) bool {
	// 这里可以集成VirusTotal API或其他恶意软件检测服务
	// 示例实现
	malwareHashes := map[string]bool{
		"d41d8cd98f00b204e9800998ecf8427e": true, // 示例哈希
	}

	return malwareHashes[fileHash]
}

// GenerateSecurityReport 生成安全报告
func (s *SecurityService) GenerateSecurityReport(reportType, period string, startDate, endDate time.Time, generatedBy uint) (*Models.SecurityReport, error) {
	report := &Models.SecurityReport{
		ReportType:  reportType,
		Title:       fmt.Sprintf("%s安全报告 - %s", reportType, period),
		Description: fmt.Sprintf("生成时间: %s 到 %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		Period:      period,
		StartDate:   startDate,
		EndDate:     endDate,
		GeneratedBy: generatedBy,
		Status:      "draft",
		Published:   false,
	}

	// 生成报告内容
	content, err := s.generateReportContent(reportType, startDate, endDate)
	if err != nil {
		return nil, err
	}

	report.Content = content
	report.Summary = s.generateReportSummary(content)

	return report, s.db.Create(report).Error
}

// generateReportContent 生成报告内容
func (s *SecurityService) generateReportContent(reportType string, startDate, endDate time.Time) (string, error) {
	var content map[string]interface{}

	switch reportType {
	case "login_attempts":
		content = s.generateLoginAttemptsReport(startDate, endDate)
	case "security_events":
		content = s.generateSecurityEventsReport(startDate, endDate)
	case "threat_intelligence":
		content = s.generateThreatIntelligenceReport(startDate, endDate)
	default:
		content = make(map[string]interface{})
	}

	jsonContent, err := json.Marshal(content)
	if err != nil {
		return "", err
	}

	return string(jsonContent), nil
}

// generateLoginAttemptsReport 生成登录尝试报告
func (s *SecurityService) generateLoginAttemptsReport(startDate, endDate time.Time) map[string]interface{} {
	var totalAttempts, successfulAttempts, failedAttempts int64
	var lockouts int64

	s.db.Model(&Models.LoginAttempt{}).
		Where("attempt_time BETWEEN ? AND ?", startDate, endDate).
		Count(&totalAttempts)

	s.db.Model(&Models.LoginAttempt{}).
		Where("attempt_time BETWEEN ? AND ? AND success = ?", startDate, endDate, true).
		Count(&successfulAttempts)

	s.db.Model(&Models.LoginAttempt{}).
		Where("attempt_time BETWEEN ? AND ? AND success = ?", startDate, endDate, false).
		Count(&failedAttempts)

	s.db.Model(&Models.AccountLockout{}).
		Where("lockout_time BETWEEN ? AND ?", startDate, endDate).
		Count(&lockouts)

	return map[string]interface{}{
		"total_attempts":      totalAttempts,
		"successful_attempts": successfulAttempts,
		"failed_attempts":     failedAttempts,
		"success_rate":        float64(successfulAttempts) / float64(totalAttempts) * 100,
		"lockouts":            lockouts,
		"top_failed_ips":      s.getTopFailedIPs(startDate, endDate),
		"top_failed_users":    s.getTopFailedUsers(startDate, endDate),
	}
}

// generateSecurityEventsReport 生成安全事件报告
func (s *SecurityService) generateSecurityEventsReport(startDate, endDate time.Time) map[string]interface{} {
	var totalEvents int64
	var highRiskEvents int64

	s.db.Model(&Models.SecurityEvent{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&totalEvents)

	s.db.Model(&Models.SecurityEvent{}).
		Where("created_at BETWEEN ? AND ? AND risk_score > ?", startDate, endDate, 70).
		Count(&highRiskEvents)

	return map[string]interface{}{
		"total_events":      totalEvents,
		"high_risk_events":  highRiskEvents,
		"risk_distribution": s.getRiskDistribution(startDate, endDate),
		"event_types":       s.getEventTypeDistribution(startDate, endDate),
		"top_sources":       s.getTopEventSources(startDate, endDate),
	}
}

// generateThreatIntelligenceReport 生成威胁情报报告
func (s *SecurityService) generateThreatIntelligenceReport(startDate, endDate time.Time) map[string]interface{} {
	var totalThreats int64
	var activeThreats int64

	s.db.Model(&Models.ThreatIntelligence{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&totalThreats)

	s.db.Model(&Models.ThreatIntelligence{}).
		Where("active = ?", true).
		Count(&activeThreats)

	return map[string]interface{}{
		"total_threats":         totalThreats,
		"active_threats":        activeThreats,
		"threat_types":          s.getThreatTypeDistribution(startDate, endDate),
		"severity_distribution": s.getThreatSeverityDistribution(startDate, endDate),
		"top_sources":           s.getTopThreatSources(startDate, endDate),
	}
}

// 辅助方法
func (s *SecurityService) getTopFailedIPs(startDate, endDate time.Time) []map[string]interface{} {
	var results []map[string]interface{}

	rows, err := s.db.Model(&Models.LoginAttempt{}).
		Select("ip_address, COUNT(*) as count").
		Where("attempt_time BETWEEN ? AND ? AND success = ?", startDate, endDate, false).
		Group("ip_address").
		Order("count DESC").
		Limit(10).
		Rows()

	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		var ipAddress string
		var count int
		rows.Scan(&ipAddress, &count)
		results = append(results, map[string]interface{}{
			"ip_address": ipAddress,
			"count":      count,
		})
	}

	return results
}

func (s *SecurityService) getTopFailedUsers(startDate, endDate time.Time) []map[string]interface{} {
	var results []map[string]interface{}

	rows, err := s.db.Model(&Models.LoginAttempt{}).
		Select("username, COUNT(*) as count").
		Where("attempt_time BETWEEN ? AND ? AND success = ?", startDate, endDate, false).
		Group("username").
		Order("count DESC").
		Limit(10).
		Rows()

	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		var username string
		var count int
		rows.Scan(&username, &count)
		results = append(results, map[string]interface{}{
			"username": username,
			"count":    count,
		})
	}

	return results
}

func (s *SecurityService) getRiskDistribution(startDate, endDate time.Time) map[string]int {
	var results []struct {
		RiskRange string
		Count     int
	}

	s.db.Model(&Models.SecurityEvent{}).
		Select("CASE WHEN risk_score < 30 THEN 'low' WHEN risk_score < 70 THEN 'medium' ELSE 'high' END as risk_range, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("risk_range").
		Scan(&results)

	distribution := make(map[string]int)
	for _, result := range results {
		distribution[result.RiskRange] = result.Count
	}

	return distribution
}

func (s *SecurityService) getEventTypeDistribution(startDate, endDate time.Time) []map[string]interface{} {
	var results []map[string]interface{}

	rows, err := s.db.Model(&Models.SecurityEvent{}).
		Select("event_type, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("event_type").
		Order("count DESC").
		Rows()

	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		var eventType string
		var count int
		rows.Scan(&eventType, &count)
		results = append(results, map[string]interface{}{
			"event_type": eventType,
			"count":      count,
		})
	}

	return results
}

func (s *SecurityService) getTopEventSources(startDate, endDate time.Time) []map[string]interface{} {
	var results []map[string]interface{}

	rows, err := s.db.Model(&Models.SecurityEvent{}).
		Select("ip_address, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("ip_address").
		Order("count DESC").
		Limit(10).
		Rows()

	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		var ipAddress string
		var count int
		rows.Scan(&ipAddress, &count)
		results = append(results, map[string]interface{}{
			"ip_address": ipAddress,
			"count":      count,
		})
	}

	return results
}

func (s *SecurityService) getThreatTypeDistribution(startDate, endDate time.Time) []map[string]interface{} {
	var results []map[string]interface{}

	rows, err := s.db.Model(&Models.ThreatIntelligence{}).
		Select("threat_type, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("threat_type").
		Order("count DESC").
		Rows()

	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		var threatType string
		var count int
		rows.Scan(&threatType, &count)
		results = append(results, map[string]interface{}{
			"threat_type": threatType,
			"count":       count,
		})
	}

	return results
}

func (s *SecurityService) getThreatSeverityDistribution(startDate, endDate time.Time) map[string]int {
	var results []struct {
		Severity string
		Count    int
	}

	s.db.Model(&Models.ThreatIntelligence{}).
		Select("severity, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("severity").
		Scan(&results)

	distribution := make(map[string]int)
	for _, result := range results {
		distribution[result.Severity] = result.Count
	}

	return distribution
}

func (s *SecurityService) getTopThreatSources(startDate, endDate time.Time) []map[string]interface{} {
	var results []map[string]interface{}

	rows, err := s.db.Model(&Models.ThreatIntelligence{}).
		Select("source, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("source").
		Order("count DESC").
		Limit(10).
		Rows()

	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		var source string
		var count int
		rows.Scan(&source, &count)
		results = append(results, map[string]interface{}{
			"source": source,
			"count":  count,
		})
	}

	return results
}

// generateReportSummary 生成报告摘要
func (s *SecurityService) generateReportSummary(content string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return "报告生成失败"
	}

	summary := "安全报告摘要：\n"

	if totalAttempts, ok := data["total_attempts"].(float64); ok {
		summary += fmt.Sprintf("- 总登录尝试次数: %.0f\n", totalAttempts)
	}

	if totalEvents, ok := data["total_events"].(float64); ok {
		summary += fmt.Sprintf("- 总安全事件数: %.0f\n", totalEvents)
	}

	if totalThreats, ok := data["total_threats"].(float64); ok {
		summary += fmt.Sprintf("- 总威胁情报数: %.0f\n", totalThreats)
	}

	return summary
}

// GetThreatDetectionService 获取威胁检测服务
func (s *SecurityService) GetThreatDetectionService() *ThreatDetectionService {
	return s.threatDetection
}

// GetThreatStatistics 获取威胁统计信息
func (s *SecurityService) GetThreatStatistics() map[string]interface{} {
	if s.threatDetection != nil {
		return s.threatDetection.GetThreatStatistics()
	}
	return make(map[string]interface{})
}

// ExportThreatIntelligence 导出威胁情报
func (s *SecurityService) ExportThreatIntelligence(format string) ([]byte, error) {
	if s.threatDetection != nil {
		return s.threatDetection.ExportThreatIntelligence(format)
	}
	return nil, fmt.Errorf("threat detection service not available")
}

// Close 关闭服务
func (s *SecurityService) Close() {
	// 关闭威胁检测服务
	if s.threatDetection != nil {
		s.threatDetection.Close()
	}
	s.cancel()
}
