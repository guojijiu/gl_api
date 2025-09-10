# 安全配置指南

## JWT密钥配置

### 重要安全提醒

⚠️ **生产环境必须修改默认JWT密钥！**

默认的JWT密钥是不安全的，仅用于开发环境。在生产环境中使用默认密钥会导致严重的安全漏洞。

### 生成安全密钥

#### 方法1：使用内置工具

```bash
# 运行密钥生成工具
go run scripts/generate-jwt-secret.go
```

#### 方法2：使用OpenSSL

```bash
# 生成64字节随机密钥
openssl rand -hex 64

# 或者生成Base64编码的密钥
openssl rand -base64 64
```

#### 方法3：使用在线工具

访问 [JWT.io](https://jwt.io) 或其他安全的密钥生成工具。

### 配置密钥

#### 环境变量方式

```bash
# 设置环境变量
export JWT_SECRET="your-generated-secret-key-here"

# 或者添加到 ~/.bashrc 或 ~/.zshrc
echo 'export JWT_SECRET="your-generated-secret-key-here"' >> ~/.bashrc
```

#### .env文件方式

```bash
# 复制示例配置文件
cp .env.example .env

# 编辑 .env 文件，修改JWT_SECRET
JWT_SECRET=your-generated-secret-key-here
```

### 密钥强度要求

#### 开发环境
- 最小长度：32字符
- 建议包含：大小写字母、数字、特殊字符

#### 生产环境
- 最小长度：64字符
- 必须包含：大小写字母、数字、特殊字符
- 强度评分：≥80分（满分100分）

### 密钥验证

应用启动时会自动验证JWT密钥：

```bash
# 启动应用，查看验证结果
go run main.go
```

如果密钥不符合要求，应用会显示错误信息并退出。

### 常见错误

#### 1. 使用默认密钥
```
错误: JWT密钥使用了不安全的默认值，请设置一个强密钥
解决: 使用上述方法生成新密钥
```

#### 2. 密钥长度不足
```
错误: JWT密钥长度不足，建议至少32个字符，当前长度: 16
解决: 生成更长的密钥
```

#### 3. 密钥强度不足
```
错误: 生产环境JWT密钥强度不足，当前强度: 60/100，建议至少80分
解决: 使用更复杂的密钥
```

### 密钥管理最佳实践

1. **定期轮换**：建议每6个月更换一次JWT密钥
2. **安全存储**：使用密钥管理服务（如AWS KMS、Azure Key Vault）
3. **环境隔离**：不同环境使用不同的密钥
4. **备份策略**：安全备份密钥，避免丢失
5. **访问控制**：限制密钥的访问权限

### 密钥轮换流程

1. 生成新密钥
2. 更新环境变量或配置文件
3. 重启应用服务
4. 验证新密钥工作正常
5. 安全删除旧密钥

### 监控和告警

建议设置以下监控：

- JWT密钥使用情况
- 异常登录尝试
- Token验证失败率
- 密钥轮换状态

### 相关文档

- [JWT安全最佳实践](https://tools.ietf.org/html/rfc7519)
- [OWASP JWT安全指南](https://owasp.org/www-project-cheat-sheets/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)
- [密钥管理最佳实践](https://cloud.google.com/kms/docs/best-practices)
