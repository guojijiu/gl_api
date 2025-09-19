package Services

import (
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Models"
	"cloud-platform-api/app/Utils"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// UserService 用户服务
// 功能说明：
// 1. 用户数据的业务逻辑处理
// 2. 用户CRUD操作
// 3. 用户权限和状态管理
// 4. 用户数据安全处理（密码字段过滤）
// 5. 缓存优化（减少数据库查询）
type UserService struct {
	BaseService
	cacheService *CacheService
}

// NewUserService 创建用户服务
// 功能说明：
// 1. 初始化用户服务实例
// 2. 返回配置好的服务对象
func NewUserService() *UserService {
	return &UserService{
		BaseService:  *NewBaseService(),
		cacheService: nil, // 延迟初始化
	}
}

// NewUserServiceWithDB 使用数据库连接创建用户服务
func NewUserServiceWithDB(db *gorm.DB) *UserService {
	service := &UserService{
		BaseService:  *NewBaseService(),
		cacheService: nil, // 延迟初始化
	}
	service.DB = db
	return service
}

// getDB 获取数据库连接
func (s *UserService) getDB() *gorm.DB {
	if s.DB != nil {
		if db, ok := s.DB.(*gorm.DB); ok {
			return db
		}
	}
	// 回退到全局数据库连接
	return Database.DB
}

// getCacheService 获取缓存服务（延迟初始化）
func (s *UserService) getCacheService() *CacheService {
	if s.cacheService == nil {
		// 这里需要传入StorageManager，暂时返回nil
		// 在实际使用中，应该通过依赖注入传入
		return nil
	}
	return s.cacheService
}

// GetUsers 获取用户列表
// 功能说明：
// 1. 获取所有用户的基本信息
// 2. 自动过滤敏感信息（密码字段）
// 3. 用于用户管理和统计
// 4. 支持分页和搜索（可扩展）
func (s *UserService) GetUsers() ([]Models.User, error) {
	var users []Models.User
	// 使用Select优化查询，只选择需要的字段，排除密码字段
	if err := s.getDB().Select("id,username,email,status,created_at,updated_at").Find(&users).Error; err != nil {
		return nil, err
	}

	// 密码字段已经在查询时被排除，无需额外清除
	return users, nil
}

// GetUser 获取单个用户
// 功能说明：
// 1. 根据用户ID获取用户详细信息
// 2. 自动过滤敏感信息（密码字段）
// 3. 处理用户不存在的情况
// 4. 用于用户资料查看和编辑
// 5. 使用缓存优化性能
func (s *UserService) GetUser(id uint) (*Models.User, error) {
	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("user:%d", id)
	if cacheService := s.getCacheService(); cacheService != nil {
		var user Models.User
		if err := cacheService.GetWithJSON(cacheKey, &user); err == nil {
			return &user, nil
		}
	}

	// 从数据库查询
	var user Models.User
	// 使用Select优化查询，只选择需要的字段，排除密码字段
	if err := s.getDB().Select("id,username,email,status,created_at,updated_at").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// 存入缓存
	if cacheService := s.getCacheService(); cacheService != nil {
		cacheService.SetWithJSON(cacheKey, &user, 15*time.Minute) // 15分钟缓存
	}

	return &user, nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(id uint, updates map[string]interface{}) (*Models.User, error) {
	var user Models.User
	if err := s.getDB().First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// 更新字段
	if err := s.getDB().Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 清除密码字段
	user.Password = ""

	return &user, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uint) error {
	var user Models.User
	if err := s.getDB().First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	return s.getDB().Delete(&user).Error
}

// CreateUser 创建用户
// 功能说明：
// 1. 创建新用户
// 2. 验证用户数据
// 3. 处理密码哈希
// 4. 返回创建的用户信息
func (s *UserService) CreateUser(user *Models.User) (*Models.User, error) {
	// 检查用户名是否已存在
	var existingUser Models.User
	if err := s.getDB().Where("username = ?", user.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	if err := s.getDB().Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("email already exists")
	}

	// 创建用户
	if err := s.getDB().Create(user).Error; err != nil {
		return nil, err
	}

	// 清除密码字段
	user.Password = ""

	return user, nil
}

// GetUserByID 根据ID获取用户
// 功能说明：
// 1. 根据用户ID获取用户信息
// 2. 自动过滤敏感信息
// 3. 处理用户不存在的情况
func (s *UserService) GetUserByID(id uint) (*Models.User, error) {
	return s.GetUser(id)
}

// GetUserByUsername 根据用户名获取用户
// 功能说明：
// 1. 根据用户名获取用户信息
// 2. 自动过滤敏感信息
// 3. 处理用户不存在的情况
func (s *UserService) GetUserByUsername(username string) (*Models.User, error) {
	var user Models.User
	if err := s.getDB().Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// 清除密码字段
	user.Password = ""

	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
// 功能说明：
// 1. 根据邮箱获取用户信息
// 2. 自动过滤敏感信息
// 3. 处理用户不存在的情况
func (s *UserService) GetUserByEmail(email string) (*Models.User, error) {
	var user Models.User
	if err := s.getDB().Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// 清除密码字段
	user.Password = ""

	return &user, nil
}

// ListUsers 获取用户列表（带分页）
// 功能说明：
// 1. 获取用户列表
// 2. 支持分页参数
// 3. 自动过滤敏感信息
// 4. 返回用户列表和总数
func (s *UserService) ListUsers(page, pageSize int) ([]Models.User, int64, error) {
	var users []Models.User
	var total int64

	// 获取总数
	if err := s.getDB().Model(&Models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := s.getDB().Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	// 清除密码字段
	for i := range users {
		users[i].Password = ""
	}

	return users, total, nil
}

// ChangePassword 修改密码
// 功能说明：
// 1. 修改用户密码
// 2. 验证旧密码
// 3. 哈希新密码
// 4. 更新数据库
func (s *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user Models.User
	if err := s.getDB().First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// 验证旧密码
	if !Utils.CheckPassword(oldPassword, user.Password) {
		return errors.New("原密码错误")
	}

	// 验证新密码强度
	isValid, validationErrors := Utils.ValidatePasswordStrength(newPassword)
	if !isValid {
		return fmt.Errorf("新密码不符合要求: %s", strings.Join(validationErrors, "; "))
	}

	// 哈希新密码
	hashedPassword, err := Utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("密码哈希失败: %v", err)
	}

	// 更新密码
	updates := map[string]interface{}{
		"password": hashedPassword,
	}

	return s.getDB().Model(&user).Updates(updates).Error
}

// ValidateUser 验证用户
// 功能说明：
// 1. 验证用户凭据
// 2. 检查用户状态
// 3. 返回验证结果
func (s *UserService) ValidateUser(username, password string) (*Models.User, error) {
	var user Models.User
	if err := s.getDB().Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, errors.New("user account is disabled")
	}

	// 验证密码
	if !Utils.CheckPassword(password, user.Password) {
		return nil, errors.New("invalid password")
	}

	// 清除密码字段
	user.Password = ""

	return &user, nil
}

// GetUserStats 获取用户统计信息
// 功能说明：
// 1. 获取用户总数
// 2. 获取活跃用户数
// 3. 获取用户状态分布
// 4. 返回统计信息
func (s *UserService) GetUserStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总用户数
	var totalUsers int64
	if err := s.getDB().Model(&Models.User{}).Count(&totalUsers).Error; err != nil {
		return nil, err
	}
	stats["total_users"] = totalUsers

	// 活跃用户数
	var activeUsers int64
	if err := s.getDB().Model(&Models.User{}).Where("status = ?", 1).Count(&activeUsers).Error; err != nil {
		return nil, err
	}
	stats["active_users"] = activeUsers

	// 禁用用户数
	var inactiveUsers int64
	if err := s.getDB().Model(&Models.User{}).Where("status = ?", 0).Count(&inactiveUsers).Error; err != nil {
		return nil, err
	}
	stats["inactive_users"] = inactiveUsers

	return stats, nil
}
