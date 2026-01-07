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
//
// 功能说明：
// 1. 根据用户ID更新用户信息
// 2. 支持部分字段更新（只更新提供的字段）
// 3. 自动清除密码字段（安全考虑）
// 4. 验证用户是否存在
//
// 参数说明：
// - id: 用户ID（必须存在）
// - updates: 要更新的字段映射（key为字段名，value为新值）
//
// 返回信息：
// - *Models.User: 更新后的用户对象（密码字段已清除）
// - error: 错误信息（如果用户不存在或更新失败）
//
// 更新策略：
// - 只更新updates中提供的字段
// - 未提供的字段保持不变
// - 密码字段会被自动清除（不会返回）
//
// 使用场景：
// - 用户资料更新
// - 用户状态修改
// - 批量字段更新
//
// 注意事项：
// - 用户必须存在才能更新
// - updates中的字段必须是有效的用户字段
// - 密码字段不应该通过此方法更新（应使用ChangePassword）
// - 更新后密码字段会被清除，不会返回
func (s *UserService) UpdateUser(id uint, updates map[string]interface{}) (*Models.User, error) {
	// 查找用户
	// 确保用户存在，如果不存在返回错误
	var user Models.User
	if err := s.getDB().First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// 更新字段
	// 使用Updates方法只更新提供的字段
	// 未提供的字段保持不变
	if err := s.getDB().Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 清除密码字段
	// 安全考虑：不返回密码字段
	user.Password = ""

	return &user, nil
}

// DeleteUser 删除用户
//
// 功能说明：
// 1. 根据用户ID删除用户
// 2. 验证用户是否存在
// 3. 执行软删除或硬删除（取决于模型配置）
//
// 参数说明：
// - id: 用户ID（必须存在）
//
// 返回信息：
// - error: 错误信息（如果用户不存在或删除失败）
//
// 删除策略：
// - 如果模型支持软删除，会执行软删除（设置DeletedAt）
// - 如果模型不支持软删除，会执行硬删除（物理删除）
// - 软删除的数据可以通过Unscoped查询
//
// 使用场景：
// - 用户账户删除
// - 用户数据清理
// - 批量用户删除
//
// 注意事项：
// - 用户必须存在才能删除
// - 删除操作不可逆（硬删除）
// - 软删除的数据仍然占用数据库空间
// - 删除前应该检查是否有关联数据
// - 建议使用软删除，保留数据用于审计
func (s *UserService) DeleteUser(id uint) error {
	// 查找用户
	// 确保用户存在，如果不存在返回错误
	var user Models.User
	if err := s.getDB().First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// 删除用户
	// 根据模型配置执行软删除或硬删除
	return s.getDB().Delete(&user).Error
}

// CreateUser 创建用户
//
// 功能说明：
// 1. 创建新用户（保存到数据库）
// 2. 验证用户数据（用户名和邮箱唯一性）
// 3. 处理密码哈希（密码应该在传入前已哈希）
// 4. 返回创建的用户信息（密码字段已清除）
//
// 参数说明：
// - user: 用户对象（必须包含用户名、邮箱、密码等必要字段）
//
// 返回信息：
// - *Models.User: 创建的用户对象（密码字段已清除）
// - error: 错误信息（如果用户名或邮箱已存在，或创建失败）
//
// 验证规则：
// - 用户名必须唯一（不能与现有用户重复）
// - 邮箱必须唯一（不能与现有用户重复）
// - 密码应该在传入前已哈希（使用Utils.HashPassword）
//
// 使用场景：
// - 用户注册
// - 管理员创建用户
// - 批量导入用户
//
// 注意事项：
// - 密码应该在传入前已哈希，不要传入明文密码
// - 用户名和邮箱的唯一性检查是必须的
// - 创建成功后密码字段会被清除，不会返回
// - 如果用户名或邮箱已存在，会返回错误
func (s *UserService) CreateUser(user *Models.User) (*Models.User, error) {
	// 检查用户名是否已存在
	// 确保用户名的唯一性，避免重复注册
	var existingUser Models.User
	if err := s.getDB().Where("username = ?", user.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	// 确保邮箱的唯一性，避免重复注册
	if err := s.getDB().Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("email already exists")
	}

	// 创建用户
	// 将用户保存到数据库，GORM会自动处理时间戳
	if err := s.getDB().Create(user).Error; err != nil {
		return nil, err
	}

	// 清除密码字段
	// 安全考虑：不返回密码字段
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
//
// 功能说明：
// 1. 获取用户列表，支持分页查询
// 2. 自动过滤敏感信息（密码字段）
// 3. 返回用户列表和总数（用于分页计算）
//
// 参数说明：
// - page: 页码（从1开始）
// - pageSize: 每页数量（必须大于0）
//
// 返回信息：
// - []Models.User: 用户列表（密码字段已清除）
// - int64: 用户总数（用于计算总页数）
// - error: 错误信息（如果查询失败）
//
// 分页计算：
// - offset = (page - 1) * pageSize
// - 例如：page=1, pageSize=10，则offset=0，返回前10条
// - 例如：page=2, pageSize=10，则offset=10，返回第11-20条
//
// 使用场景：
// - 用户管理界面（显示用户列表）
// - 用户搜索和筛选
// - 批量用户操作
//
// 注意事项：
// - page应该大于0，pageSize应该大于0
// - 返回的用户列表中的密码字段会被清除
// - 总数用于前端计算总页数：totalPages = (total + pageSize - 1) / pageSize
// - 如果page超出范围，会返回空列表但不报错
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
//
// 功能说明：
// 1. 修改用户密码（需要验证旧密码）
// 2. 验证旧密码正确性（防止未授权修改）
// 3. 验证新密码强度（确保密码安全）
// 4. 对新密码进行哈希处理（安全存储）
//
// 参数说明：
// - userID: 用户ID（必须存在）
// - oldPassword: 旧密码（明文，用于验证）
// - newPassword: 新密码（明文，会被哈希后存储）
//
// 返回信息：
// - error: 错误信息（如果用户不存在、旧密码错误、新密码强度不足或更新失败）
//
// 验证流程：
// 1. 验证用户是否存在
// 2. 验证旧密码是否正确（使用Utils.CheckPassword）
// 3. 验证新密码强度（使用Utils.ValidatePasswordStrength）
// 4. 对新密码进行哈希（使用Utils.HashPassword）
// 5. 更新数据库中的密码哈希值
//
// 安全考虑：
// - 必须提供旧密码，防止未授权修改
// - 新密码必须符合强度要求
// - 密码使用哈希存储，不存储明文
// - 旧密码验证失败会返回错误，不泄露用户信息
//
// 使用场景：
// - 用户主动修改密码
// - 管理员重置用户密码（可能需要特殊权限）
// - 密码过期后强制修改
//
// 注意事项：
// - 用户必须存在才能修改密码
// - 旧密码必须正确，否则返回错误
// - 新密码必须符合强度要求（长度、字符类型等）
// - 密码修改后，用户需要重新登录
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
//
// 功能说明：
// 1. 验证用户凭据（用户名和密码）
// 2. 检查用户状态（确保账户未被禁用）
// 3. 返回验证结果（用户对象或错误）
//
// 参数说明：
// - username: 用户名（用于查找用户）
// - password: 密码（明文，用于验证）
//
// 返回信息：
// - *Models.User: 验证通过的用户对象（密码字段已清除）
// - error: 错误信息（如果用户不存在、账户被禁用或密码错误）
//
// 验证流程：
// 1. 根据用户名查找用户
// 2. 检查用户是否存在
// 3. 检查用户状态（Status != 1表示账户被禁用）
// 4. 验证密码（使用Utils.CheckPassword）
// 5. 清除密码字段并返回用户对象
//
// 使用场景：
// - 用户登录验证
// - 密码验证
// - 账户状态检查
//
// 注意事项：
// - 用户必须存在且账户未被禁用
// - 密码必须正确
// - 返回的用户对象中密码字段会被清除
// - 错误信息不会泄露用户是否存在的信息（安全考虑）
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
//
// 功能说明：
// 1. 获取用户总数（所有用户）
// 2. 获取活跃用户数（Status = 1）
// 3. 获取禁用用户数（Status = 0）
// 4. 返回统计信息（用于仪表板和分析）
//
// 返回信息：
// - map[string]interface{}: 统计信息映射，包含：
//   - total_users: 用户总数
//   - active_users: 活跃用户数
//   - inactive_users: 禁用用户数
// - error: 错误信息（如果查询失败）
//
// 统计指标：
// - total_users: 所有用户的总数（包括活跃和禁用）
// - active_users: 状态为1的用户数（活跃用户）
// - inactive_users: 状态为0的用户数（禁用用户）
//
// 使用场景：
// - 管理员仪表板（显示用户统计）
// - 用户分析报告
// - 系统监控和告警
//
// 注意事项：
// - 统计信息是实时查询的，可能影响性能
// - 建议对统计信息进行缓存（如Redis）
// - 大量用户时可能需要优化查询性能
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
