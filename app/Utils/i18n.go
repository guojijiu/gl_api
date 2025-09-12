package Utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// I18nManager 国际化管理器
type I18nManager struct {
	translations map[string]map[string]string
	fallbackLang string
	mutex        sync.RWMutex
}

// 全局国际化管理器
var globalI18nManager *I18nManager

// InitI18n 初始化国际化
func InitI18n(translationsDir string, fallbackLang string) error {
	if globalI18nManager != nil {
		return fmt.Errorf("国际化已经初始化")
	}

	manager := &I18nManager{
		translations: make(map[string]map[string]string),
		fallbackLang: fallbackLang,
	}

	// 加载翻译文件
	if err := manager.loadTranslations(translationsDir); err != nil {
		return fmt.Errorf("加载翻译文件失败: %v", err)
	}

	globalI18nManager = manager
	return nil
}

// loadTranslations 加载翻译文件
func (im *I18nManager) loadTranslations(translationsDir string) error {
	// 检查目录是否存在
	if _, err := os.Stat(translationsDir); os.IsNotExist(err) {
		return fmt.Errorf("翻译目录不存在: %s", translationsDir)
	}

	// 遍历翻译文件
	return filepath.Walk(translationsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理JSON文件
		if !strings.HasSuffix(strings.ToLower(path), ".json") {
			return nil
		}

		// 提取语言代码（文件名）
		lang := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

		// 读取翻译文件
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("读取翻译文件失败 %s: %v", path, err)
		}

		// 解析JSON
		var translations map[string]string
		if err := json.Unmarshal(data, &translations); err != nil {
			return fmt.Errorf("解析翻译文件失败 %s: %v", path, err)
		}

		// 存储翻译
		im.mutex.Lock()
		im.translations[lang] = translations
		im.mutex.Unlock()

		return nil
	})
}

// T 翻译文本
func T(key string, lang string, args ...interface{}) string {
	if globalI18nManager == nil {
		return key
	}

	return globalI18nManager.Translate(key, lang, args...)
}

// Translate 翻译文本
func (im *I18nManager) Translate(key string, lang string, args ...interface{}) string {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	// 获取翻译文本
	var text string
	var ok bool

	// 尝试指定语言
	if translations, exists := im.translations[lang]; exists {
		text, ok = translations[key]
	}

	// 如果指定语言没有找到，尝试回退语言
	if !ok && lang != im.fallbackLang {
		if translations, exists := im.translations[im.fallbackLang]; exists {
			text, ok = translations[key]
		}
	}

	// 如果还是没有找到，返回键名
	if !ok {
		return key
	}

	// 格式化参数
	if len(args) > 0 {
		return fmt.Sprintf(text, args...)
	}

	return text
}

// GetSupportedLanguages 获取支持的语言列表
func (im *I18nManager) GetSupportedLanguages() []string {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	languages := make([]string, 0, len(im.translations))
	for lang := range im.translations {
		languages = append(languages, lang)
	}

	return languages
}

// AddTranslation 添加翻译
func (im *I18nManager) AddTranslation(lang string, key string, value string) {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	if im.translations[lang] == nil {
		im.translations[lang] = make(map[string]string)
	}

	im.translations[lang][key] = value
}

// RemoveTranslation 移除翻译
func (im *I18nManager) RemoveTranslation(lang string, key string) {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	if translations, exists := im.translations[lang]; exists {
		delete(translations, key)
	}
}

// GetTranslation 获取翻译
func (im *I18nManager) GetTranslation(lang string, key string) (string, bool) {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	if translations, exists := im.translations[lang]; exists {
		value, ok := translations[key]
		return value, ok
	}

	return "", false
}

// GetTranslations 获取所有翻译
func (im *I18nManager) GetTranslations(lang string) map[string]string {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	if translations, exists := im.translations[lang]; exists {
		// 返回副本，避免外部修改
		result := make(map[string]string)
		for k, v := range translations {
			result[k] = v
		}
		return result
	}

	return nil
}

// SetFallbackLanguage 设置回退语言
func (im *I18nManager) SetFallbackLanguage(lang string) {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	im.fallbackLang = lang
}

// GetFallbackLanguage 获取回退语言
func (im *I18nManager) GetFallbackLanguage() string {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	return im.fallbackLang
}

// GetI18nManager 获取全局国际化管理器
func GetI18nManager() *I18nManager {
	return globalI18nManager
}

// 常用翻译键
const (
	// 通用
	MsgSuccess = "msg.success"
	MsgError   = "msg.error"
	MsgWarning = "msg.warning"
	MsgInfo    = "msg.info"

	// 认证
	MsgLoginSuccess     = "auth.login_success"
	MsgLoginFailed      = "auth.login_failed"
	MsgLogoutSuccess    = "auth.logout_success"
	MsgTokenExpired     = "auth.token_expired"
	MsgTokenInvalid     = "auth.token_invalid"
	MsgPermissionDenied = "auth.permission_denied"

	// 用户
	MsgUserCreated  = "user.created"
	MsgUserUpdated  = "user.updated"
	MsgUserDeleted  = "user.deleted"
	MsgUserNotFound = "user.not_found"
	MsgUserExists   = "user.exists"
	MsgUserInvalid  = "user.invalid"

	// 验证
	MsgValidationFailed = "validation.failed"
	MsgFieldRequired    = "validation.field_required"
	MsgFieldInvalid     = "validation.field_invalid"
	MsgFieldTooLong     = "validation.field_too_long"
	MsgFieldTooShort    = "validation.field_too_short"

	// 系统
	MsgSystemError        = "system.error"
	MsgSystemMaintenance  = "system.maintenance"
	MsgServiceUnavailable = "system.service_unavailable"
	MsgRateLimitExceeded  = "system.rate_limit_exceeded"
)

// 创建默认翻译文件
func CreateDefaultTranslations(translationsDir string) error {
	// 创建目录
	if err := os.MkdirAll(translationsDir, 0755); err != nil {
		return fmt.Errorf("创建翻译目录失败: %v", err)
	}

	// 中文翻译
	zhTranslations := map[string]string{
		MsgSuccess: "操作成功",
		MsgError:   "操作失败",
		MsgWarning: "警告",
		MsgInfo:    "信息",

		MsgLoginSuccess:     "登录成功",
		MsgLoginFailed:      "登录失败",
		MsgLogoutSuccess:    "退出成功",
		MsgTokenExpired:     "令牌已过期",
		MsgTokenInvalid:     "令牌无效",
		MsgPermissionDenied: "权限不足",

		MsgUserCreated:  "用户创建成功",
		MsgUserUpdated:  "用户更新成功",
		MsgUserDeleted:  "用户删除成功",
		MsgUserNotFound: "用户不存在",
		MsgUserExists:   "用户已存在",
		MsgUserInvalid:  "用户信息无效",

		MsgValidationFailed: "验证失败",
		MsgFieldRequired:    "字段必填",
		MsgFieldInvalid:     "字段格式无效",
		MsgFieldTooLong:     "字段过长",
		MsgFieldTooShort:    "字段过短",

		MsgSystemError:        "系统错误",
		MsgSystemMaintenance:  "系统维护中",
		MsgServiceUnavailable: "服务不可用",
		MsgRateLimitExceeded:  "请求频率过高",
	}

	// 英文翻译
	enTranslations := map[string]string{
		MsgSuccess: "Operation successful",
		MsgError:   "Operation failed",
		MsgWarning: "Warning",
		MsgInfo:    "Information",

		MsgLoginSuccess:     "Login successful",
		MsgLoginFailed:      "Login failed",
		MsgLogoutSuccess:    "Logout successful",
		MsgTokenExpired:     "Token expired",
		MsgTokenInvalid:     "Token invalid",
		MsgPermissionDenied: "Permission denied",

		MsgUserCreated:  "User created successfully",
		MsgUserUpdated:  "User updated successfully",
		MsgUserDeleted:  "User deleted successfully",
		MsgUserNotFound: "User not found",
		MsgUserExists:   "User already exists",
		MsgUserInvalid:  "User information invalid",

		MsgValidationFailed: "Validation failed",
		MsgFieldRequired:    "Field is required",
		MsgFieldInvalid:     "Field format is invalid",
		MsgFieldTooLong:     "Field is too long",
		MsgFieldTooShort:    "Field is too short",

		MsgSystemError:        "System error",
		MsgSystemMaintenance:  "System maintenance",
		MsgServiceUnavailable: "Service unavailable",
		MsgRateLimitExceeded:  "Rate limit exceeded",
	}

	// 保存中文翻译
	zhData, err := json.MarshalIndent(zhTranslations, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化中文翻译失败: %v", err)
	}

	if err := os.WriteFile(filepath.Join(translationsDir, "zh.json"), zhData, 0644); err != nil {
		return fmt.Errorf("保存中文翻译失败: %v", err)
	}

	// 保存英文翻译
	enData, err := json.MarshalIndent(enTranslations, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化英文翻译失败: %v", err)
	}

	if err := os.WriteFile(filepath.Join(translationsDir, "en.json"), enData, 0644); err != nil {
		return fmt.Errorf("保存英文翻译失败: %v", err)
	}

	return nil
}
