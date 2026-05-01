# CHANGELOG

### v1.1

1. 密码存储改用 bcrypt，替代 SHA-256
2. Token 内嵌用户名，格式升级为 hex(time).username.expiry.sig，向下兼容旧 Token
3. CORS 允许来源改为环境变量 CORS_ORIGIN 配置，不再硬编码
4. 新增 GetClientIP 工具函数，支持 X-Forwarded-For / X-Real-IP 解析
5. 合并 config.go 到 main.go，消除重复定义

### v1.0

1. 支持图片上传（jpg/png/gif/webp），自动生成缩略图
2. 支持图片访问、删除、列表，分页查询
3. Bearer Token 认证，支持静态Token和密码登录签发动态Token
4. 游客模式，每小时限频20次
5. 全中文暗色UI，响应式布局
6. SQLite 存储，WAL模式
7. 路径遍历防护，缓存优化
8. Docker 多阶段构建支持
