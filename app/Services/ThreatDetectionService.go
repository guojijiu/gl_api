package Services

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// ThreatDetectionService 威胁检测服务
type ThreatDetectionService struct {
	db            *gorm.DB
	config        *Config.SecurityConfig
	threatIPs     map[string]ThreatInfo
	malwareHashes map[string]MalwareInfo
	phishingURLs  map[string]PhishingInfo
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

// ThreatInfo 威胁信息
type ThreatInfo struct {
	IP          string    `json:"ip"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Source      string    `json:"source"`
	Description string    `json:"description"`
	LastSeen    time.Time `json:"last_seen"`
	Confidence  float64   `json:"confidence"`
}

// MalwareInfo 恶意软件信息
type MalwareInfo struct {
	Hash        string    `json:"hash"`
	Type        string    `json:"type"`
	Family      string    `json:"family"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	LastSeen    time.Time `json:"last_seen"`
	Confidence  float64   `json:"confidence"`
}

// PhishingInfo 钓鱼网站信息
type PhishingInfo struct {
	URL         string    `json:"url"`
	Domain      string    `json:"domain"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	LastSeen    time.Time `json:"last_seen"`
	Confidence  float64   `json:"confidence"`
}

// ThreatIntelligenceSource 威胁情报源
type ThreatIntelligenceSource struct {
	Name           string        `json:"name"`
	URL            string        `json:"url"`
	APIKey         string        `json:"api_key"`
	Enabled        bool          `json:"enabled"`
	UpdateInterval time.Duration `json:"update_interval"`
}

// NewThreatDetectionService 创建威胁检测服务
func NewThreatDetectionService(db *gorm.DB, config *Config.SecurityConfig) *ThreatDetectionService {
	ctx, cancel := context.WithCancel(context.Background())

	service := &ThreatDetectionService{
		db:            db,
		config:        config,
		threatIPs:     make(map[string]ThreatInfo),
		malwareHashes: make(map[string]MalwareInfo),
		phishingURLs:  make(map[string]PhishingInfo),
		ctx:           ctx,
		cancel:        cancel,
	}

	// 初始化服务
	service.initialize()

	return service
}

// initialize 初始化服务
func (s *ThreatDetectionService) initialize() {
	// 加载本地威胁情报
	s.loadLocalThreatIntelligence()

	// 启动威胁情报更新任务
	go s.startThreatIntelligenceUpdates()

	// 启动威胁检测任务
	go s.startThreatDetection()
}

// loadLocalThreatIntelligence 加载本地威胁情报
func (s *ThreatDetectionService) loadLocalThreatIntelligence() {
	// 从数据库加载威胁情报
	var threats []Models.ThreatIntelligence
	s.db.Where("active = ?", true).Find(&threats)

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, threat := range threats {
		if threat.IPAddress != "" {
			s.threatIPs[threat.IPAddress] = ThreatInfo{
				IP:          threat.IPAddress,
				Type:        threat.ThreatType,
				Severity:    threat.Severity,
				Source:      threat.Source,
				Description: threat.Description,
				LastSeen:    threat.LastSeen,
				Confidence:  threat.Confidence,
			}
		}
	}
}

// startThreatIntelligenceUpdates 启动威胁情报更新任务
func (s *ThreatDetectionService) startThreatIntelligenceUpdates() {
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
func (s *ThreatDetectionService) updateThreatIntelligence() {
	// 更新各个威胁情报源
	sources := s.getThreatIntelligenceSources()

	for _, source := range sources {
		if source.Enabled {
			go s.fetchThreatIntelligence(source)
		}
	}
}

// getThreatIntelligenceSources 获取威胁情报源
func (s *ThreatDetectionService) getThreatIntelligenceSources() []ThreatIntelligenceSource {
	return []ThreatIntelligenceSource{
		{
			Name:           "AbuseIPDB",
			URL:            "https://api.abuseipdb.com/api/v2/blacklist",
			APIKey:         s.config.ThreatProtection.VirusTotalAPIKey, // 复用API密钥
			Enabled:        true,
			UpdateInterval: 24 * time.Hour,
		},
		{
			Name:           "Blocklist.de",
			URL:            "https://api.blocklist.de/get.php",
			Enabled:        true,
			UpdateInterval: 12 * time.Hour,
		},
		{
			Name:           "Malware Domain List",
			URL:            "https://malwaredomains.com/feeds/domains.txt",
			Enabled:        true,
			UpdateInterval: 6 * time.Hour,
		},
	}
}

// fetchThreatIntelligence 获取威胁情报
func (s *ThreatDetectionService) fetchThreatIntelligence(source ThreatIntelligenceSource) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(s.ctx, "GET", source.URL, nil)
	if err != nil {
		return
	}

	if source.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+source.APIKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// 解析威胁情报数据
	s.parseThreatIntelligence(source.Name, body)
}

// parseThreatIntelligence 解析威胁情报数据
func (s *ThreatDetectionService) parseThreatIntelligence(source string, data []byte) {
	lines := strings.Split(string(data), "\n")

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 根据源类型解析数据
		switch source {
		case "AbuseIPDB":
			s.parseAbuseIPDBData(line)
		case "Blocklist.de":
			s.parseBlocklistData(line)
		case "Malware Domain List":
			s.parseMalwareDomainData(line)
		}
	}
}

// parseAbuseIPDBData 解析AbuseIPDB数据
func (s *ThreatDetectionService) parseAbuseIPDBData(line string) {
	// 简单的IP地址提取
	ipRegex := regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	ips := ipRegex.FindAllString(line, -1)

	for _, ip := range ips {
		s.threatIPs[ip] = ThreatInfo{
			IP:          ip,
			Type:        "malicious",
			Severity:    "high",
			Source:      "AbuseIPDB",
			Description: "Malicious IP from AbuseIPDB",
			LastSeen:    time.Now(),
			Confidence:  0.8,
		}
	}
}

// parseBlocklistData 解析Blocklist.de数据
func (s *ThreatDetectionService) parseBlocklistData(line string) {
	// 简单的IP地址提取
	ipRegex := regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	ips := ipRegex.FindAllString(line, -1)

	for _, ip := range ips {
		s.threatIPs[ip] = ThreatInfo{
			IP:          ip,
			Type:        "attacker",
			Severity:    "medium",
			Source:      "Blocklist.de",
			Description: "Attacker IP from Blocklist.de",
			LastSeen:    time.Now(),
			Confidence:  0.7,
		}
	}
}

// parseMalwareDomainData 解析恶意软件域名数据
func (s *ThreatDetectionService) parseMalwareDomainData(line string) {
	// 简单的域名提取
	domainRegex := regexp.MustCompile(`[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*`)
	domains := domainRegex.FindAllString(line, -1)

	for _, domain := range domains {
		s.phishingURLs[domain] = PhishingInfo{
			URL:         "http://" + domain,
			Domain:      domain,
			Type:        "malware",
			Severity:    "high",
			Description: "Malware domain",
			LastSeen:    time.Now(),
			Confidence:  0.9,
		}
	}
}

// startThreatDetection 启动威胁检测任务
func (s *ThreatDetectionService) startThreatDetection() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.performThreatDetection()
		}
	}
}

// performThreatDetection 执行威胁检测
func (s *ThreatDetectionService) performThreatDetection() {
	// 检测可疑的登录尝试
	s.detectSuspiciousLogins()

	// 检测异常的网络活动
	s.detectAnomalousNetworkActivity()

	// 检测恶意文件上传
	s.detectMaliciousFileUploads()
}

// detectSuspiciousLogins 检测可疑登录
func (s *ThreatDetectionService) detectSuspiciousLogins() {
	// 查找最近1小时内的失败登录
	var failedLogins []Models.LoginAttempt
	s.db.Where("success = ? AND attempt_time > ?", false, time.Now().Add(-time.Hour)).
		Find(&failedLogins)

	// 按IP地址分组统计
	ipCounts := make(map[string]int)
	for _, login := range failedLogins {
		ipCounts[login.IPAddress]++
	}

	// 检测异常IP
	for ip, count := range ipCounts {
		if count > 10 { // 1小时内失败登录超过10次
			s.recordThreatEvent("suspicious_login", "high", ip, fmt.Sprintf("IP %s has %d failed login attempts in 1 hour", ip, count))
		}
	}
}

// detectAnomalousNetworkActivity 检测异常网络活动
func (s *ThreatDetectionService) detectAnomalousNetworkActivity() {
	// 查找最近1小时内的安全事件
	var events []Models.SecurityEvent
	s.db.Where("created_at > ?", time.Now().Add(-time.Hour)).
		Find(&events)

	// 按IP地址分组统计
	ipCounts := make(map[string]int)
	for _, event := range events {
		ipCounts[event.IPAddress]++
	}

	// 检测异常IP
	for ip, count := range ipCounts {
		if count > 50 { // 1小时内安全事件超过50次
			s.recordThreatEvent("anomalous_network_activity", "medium", ip, fmt.Sprintf("IP %s has %d security events in 1 hour", ip, count))
		}
	}
}

// detectMaliciousFileUploads 检测恶意文件上传
func (s *ThreatDetectionService) detectMaliciousFileUploads() {
	// 这里可以集成文件扫描服务
	// 暂时使用简单的文件扩展名检查
	maliciousExtensions := []string{".exe", ".bat", ".cmd", ".com", ".pif", ".scr", ".vbs", ".js", ".jar", ".php"}

	// 查找最近1小时内的文件上传事件
	var events []Models.SecurityEvent
	s.db.Where("event_type = ? AND created_at > ?", "file_upload", time.Now().Add(-time.Hour)).
		Find(&events)

	for _, event := range events {
		// 检查文件扩展名
		for _, ext := range maliciousExtensions {
			if strings.Contains(strings.ToLower(event.Details), ext) {
				s.recordThreatEvent("malicious_file_upload", "high", event.IPAddress, fmt.Sprintf("Malicious file upload detected: %s", event.Details))
				break
			}
		}
	}
}

// recordThreatEvent 记录威胁事件
func (s *ThreatDetectionService) recordThreatEvent(eventType, severity, ipAddress, description string) {
	event := Models.SecurityEvent{
		EventType:    eventType,
		EventLevel:   severity,
		IPAddress:    ipAddress,
		Details:      description,
		RiskScore:    s.calculateRiskScore(severity),
		AnomalyScore: 0.8,
		Blocked:      false,
		Alerted:      true,
	}

	s.db.Create(&event)
}

// calculateRiskScore 计算风险分数
func (s *ThreatDetectionService) calculateRiskScore(severity string) float64 {
	switch severity {
	case "critical":
		return 100.0
	case "high":
		return 80.0
	case "medium":
		return 60.0
	case "low":
		return 40.0
	default:
		return 50.0
	}
}

// CheckThreatIP 检查威胁IP
func (s *ThreatDetectionService) CheckThreatIP(ipAddress string) (bool, ThreatInfo) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if threat, exists := s.threatIPs[ipAddress]; exists {
		return true, threat
	}

	return false, ThreatInfo{}
}

// CheckMalwareHash 检查恶意软件哈希
func (s *ThreatDetectionService) CheckMalwareHash(hash string) (bool, MalwareInfo) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if malware, exists := s.malwareHashes[hash]; exists {
		return true, malware
	}

	return false, MalwareInfo{}
}

// CheckPhishingURL 检查钓鱼URL
func (s *ThreatDetectionService) CheckPhishingURL(url string) (bool, PhishingInfo) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 提取域名
	domain := s.extractDomain(url)
	if phishing, exists := s.phishingURLs[domain]; exists {
		return true, phishing
	}

	return false, PhishingInfo{}
}

// extractDomain 提取域名
func (s *ThreatDetectionService) extractDomain(url string) string {
	// 简单的域名提取
	url = strings.ToLower(url)
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")

	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[0]
	}

	return url
}

// GetThreatStatistics 获取威胁统计信息
func (s *ThreatDetectionService) GetThreatStatistics() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]interface{})

	// 威胁IP统计
	stats["threat_ips"] = len(s.threatIPs)

	// 恶意软件哈希统计
	stats["malware_hashes"] = len(s.malwareHashes)

	// 钓鱼URL统计
	stats["phishing_urls"] = len(s.phishingURLs)

	// 按严重程度统计威胁IP
	severityCounts := make(map[string]int)
	for _, threat := range s.threatIPs {
		severityCounts[threat.Severity]++
	}
	stats["threat_ip_severity"] = severityCounts

	// 按类型统计威胁IP
	typeCounts := make(map[string]int)
	for _, threat := range s.threatIPs {
		typeCounts[threat.Type]++
	}
	stats["threat_ip_types"] = typeCounts

	return stats
}

// ExportThreatIntelligence 导出威胁情报
func (s *ThreatDetectionService) ExportThreatIntelligence(format string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := map[string]interface{}{
		"threat_ips":     s.threatIPs,
		"malware_hashes": s.malwareHashes,
		"phishing_urls":  s.phishingURLs,
		"export_time":    time.Now(),
		"total_threats":  len(s.threatIPs) + len(s.malwareHashes) + len(s.phishingURLs),
	}

	switch format {
	case "json":
		return json.MarshalIndent(data, "", "  ")
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// Close 关闭服务
func (s *ThreatDetectionService) Close() {
	s.cancel()
}
