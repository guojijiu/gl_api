package Services

import (
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Models"
	"errors"
	"gorm.io/gorm"
)

// UserService 用户服务
// 功能说明：
// 1. 用户数据的业务逻辑处理
// 2. 用户CRUD操作
// 3. 用户权限和状态管理
// 4. 用户数据安全处理（密码字段过滤）
type UserService struct {
	BaseService
}

// NewUserService 创建用户服务
// 功能说明：
// 1. 初始化用户服务实例
// 2. 返回配置好的服务对象
func NewUserService() *UserService {
	return &UserService{}
}

// GetUsers 获取用户列表
// 功能说明：
// 1. 获取所有用户的基本信息
// 2. 自动过滤敏感信息（密码字段）
// 3. 用于用户管理和统计
// 4. 支持分页和搜索（可扩展）
func (s *UserService) GetUsers() ([]Models.User, error) {
	var users []Models.User
	if err := Database.DB.Find(&users).Error; err != nil {
		return nil, err
	}

	// 清除密码字段
	for i := range users {
		users[i].Password = ""
	}

	return users, nil
}

// GetUser 获取单个用户
// 功能说明：
// 1. 根据用户ID获取用户详细信息
// 2. 自动过滤敏感信息（密码字段）
// 3. 处理用户不存在的情况
// 4. 用于用户资料查看和编辑
func (s *UserService) GetUser(id uint) (*Models.User, error) {
	var user Models.User
	if err := Database.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// 清除密码字段
	user.Password = ""

	return &user, nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(id uint, updates map[string]interface{}) (*Models.User, error) {
	var user Models.User
	if err := Database.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// 更新字段
	if err := Database.DB.Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 清除密码字段
	user.Password = ""

	return &user, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uint) error {
	var user Models.User
	if err := Database.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	return Database.DB.Delete(&user).Error
}
