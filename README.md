<div align="center">

# 良辰图床

**一个用纯 Go 写的轻量自托管图床服务**

没有花里胡哨的依赖，编译完一个二进制文件，扔服务器上就能跑。

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)
![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-blue?style=flat-square)

</div>

---

## 这是什么

良辰图床是一个自托管的图片托管服务。你可以把它理解成一个私有的 Imgur —— 图片上传到你自己的服务器，链接你自己掌控，不用担心哪天服务关停图片全没了。

写博客、做笔记、发 Markdown 文档，图片往这一丢，拿到链接直接用。

## 功能

- **图片上传** — 支持 API 调用和 Web 页面拖拽上传、粘贴上传
- **直接访问** — 生成干净的图片 URL，Markdown 里直接引用
- **自动生成缩略图** — 上传时自动压缩，列表页加载更快
- **图片管理** — 列表查看、分页浏览、Lightbox 预览、一键删除
- **密码登录** — 管理员通过密码登录，获得删除等管理权限
- **游客模式** — 未登录用户可以浏览图片库和上传图片（有频率限制）
- **Web UI** — 中文管理界面，拖拽上传，复制链接（支持 Markdown / HTML 格式）
- **跨平台** — Linux、macOS、Windows 随便编译
- **Docker 支持** — 一行命令容器化部署

## 支持格式

| 格式 | 后缀 | 说明 |
|------|------|------|
| JPEG | `.jpg` | 最常见的图片格式 |
| PNG | `.png` | 支持透明通道 |
| GIF | `.gif` | 支持动图 |
| WebP | `.webp` | Google 出品，体积更小 |

单文件大小默认限制 **10MB**，可通过环境变量调整。

## 快速开始

### 方式一：直接编译

```bash
# 克隆仓库
git clone https://github.com/2759069519/liangchen-tuchuang.git
cd liangchen-tuchuang

# 拉依赖
go mod tidy

# 编译
go build -o imgbed .

# 启动
ADMIN_PASSWORD=你的密码 ./imgbed
```

浏览器打开 `http://localhost:8080`，点右下角齿轮图标输入密码登录。

### 方式二：Docker

```bash
docker build -t imgbed .
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/uploads:/app/uploads \
  -v $(pwd)/data:/app/data \
  -e ADMIN_PASSWORD=你的密码 \
  imgbed
```

## 配置项

所有配置通过环境变量传入，没有配置文件，干净利落：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | `8080` | 监听端口 |
| `UPLOAD_DIR` | `./uploads` | 图片存储目录 |
| `DB_PATH` | `./imgbed.db` | SQLite 数据库路径 |
| `ADMIN_PASSWORD` | *(空)* | 管理员密码，设置后可通过 Web 界面登录获取管理权限 |
| `AUTH_TOKEN` | *(空)* | API 认证 Token（兼容旧版，优先级低于密码登录） |
| `MAX_FILE_SIZE` | `10485760` | 最大上传大小（字节），默认 10MB |
| `BASE_URL` | `http://localhost:8080` | 返回给客户端的图片基础 URL |

> **`BASE_URL` 很重要** — 如果你用了 Nginx 反代或者绑了域名，记得改成你的实际地址，不然返回的图片链接会是 `localhost`。

## 权限模型

| 功能 | 游客（未登录） | 管理员（已登录） |
|------|:---:|:---:|
| 浏览图片库 | ✅ | ✅ |
| 上传图片 | ✅ 每小时 20 张 | ✅ 无限制 |
| 删除图片 | ❌ | ✅ |

管理员登录方式：点击页面右下角齿轮图标 → 输入密码 → 登录成功后图标变绿。

登录后获得的 Token 有效期 24 小时，过期需重新登录。

## API 文档

### 管理员登录

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"password": "你的密码"}'
```

返回：

```json
{
  "token": "xxxxx.yyyyy.zzzzz",
  "message": "login ok"
}
```

后续请求在 Header 中携带此 Token：`Authorization: Bearer <token>`

### 上传图片

```bash
curl -X POST http://localhost:8080/api/upload \
  -F "file=@photo.jpg"
```

游客可直接上传（无需 Token），每小时限 20 张。携带 Token 则无限制。

返回：

```json
{
  "id": "a1b2c3d4",
  "url": "https://img.example.com/img/a1b2c3d4.jpg",
  "thumb": "https://img.example.com/img/a1b2c3d4_thumb.jpg",
  "remaining": 19
}
```

### 查看图片列表

```bash
curl http://localhost:8080/api/images?limit=20&offset=0
```

游客可直接浏览，无需 Token。返回：

```json
{
  "items": [
    {
      "id": "a1b2c3d4",
      "filename": "photo.jpg",
      "url": "https://img.example.com/img/a1b2c3d4.jpg",
      "thumb": "https://img.example.com/img/a1b2c3d4_thumb.jpg",
      "size": 204800,
      "width": 1920,
      "height": 1080
    }
  ],
  "total": 42,
  "limit": 20,
  "offset": 0
}
```

### 删除图片

需要管理员权限：

```bash
curl -X DELETE http://localhost:8080/api/images/a1b2c3d4 \
  -H "Authorization: Bearer <token>"
```

### 直接访问图片

GET 请求不需要 Token，图片链接可以直接在浏览器、Markdown、任何地方用：

```
https://img.example.com/img/a1b2c3d4.jpg
```

> 缩略图同理：`https://img.example.com/img/a1b2c3d4_thumb.jpg`

## 项目结构

```
liangchen-tuchuang/
├── main.go              # 入口，路由注册，启动服务器
├── auth/
│   └── token.go         # Token 生成与验证（HMAC-SHA256）
├── handler/
│   ├── upload.go        # 上传处理（类型校验、缩略图生成、频率限制）
│   ├── serve.go         # 图片访问（带缓存头、路径安全校验）
│   ├── delete.go        # 删除（数据库 + 文件同步清理）
│   ├── list.go          # 列表查询（分页）
│   └── login.go         # 管理员密码登录
├── middleware/
│   ├── auth.go          # 认证中间件（静态 Token + 登录 Token）
│   ├── cors.go          # CORS 跨域支持
│   └── logging.go       # 请求日志
├── store/
│   └── store.go         # SQLite 数据层（图片元数据 + 上传频率记录）
├── static/
│   └── index.html       # Web 管理界面
├── uploads/             # 图片存储目录（自动创建）
├── Dockerfile           # Docker 构建文件
├── Makefile             # 快捷构建命令
└── .github/
    └── workflows/
        └── build.yml    # GitHub Actions CI
```

## 技术选型

| 组件 | 选择 | 为什么 |
|------|------|--------|
| Web 框架 | 标准库 `net/http` | Go 1.22+ 的 `ServeMux` 已经够用了，不需要引入第三方路由 |
| 数据库 | SQLite (`modernc.org/sqlite`) | 纯 Go 实现，不需要 CGO，交叉编译无痛 |
| 图片处理 | `disintegration/imaging` | API 简洁，Lanczos 缩放质量好 |
| UUID | `google/uuid` | 生成短 ID 做文件名，避免冲突和乱码 |
| 认证 | HMAC-SHA256 | 标准库实现，无额外依赖，Token 24 小时过期 |

外部依赖一共就 **3 个**，编译出来一个二进制文件，没有任何运行时依赖。

## 跨平台编译

```bash
# Linux x86_64
GOOS=linux GOARCH=amd64 go build -o imgbed-linux .

# Windows
GOOS=windows GOARCH=amd64 go build -o imgbed.exe .

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o imgbed-mac .

# Linux ARM64（树莓派之类的）
GOOS=linux GOARCH=arm64 go build -o imgbed-arm64 .
```

## 部署建议

如果你打算在公网用，几个建议：

1. **设置 `ADMIN_PASSWORD`** — 管理员密码，保护你的图片不被随便删
2. **Nginx 反代** — 前面套一层 Nginx，加上 HTTPS
3. **`BASE_URL` 要改** — 改成你的域名，比如 `https://img.yourdomain.com`
4. **定期备份** — `uploads/` 目录和 `imgbed.db` 是核心数据
5. **磁盘空间** — 图片这东西涨得快，注意监控磁盘

一个简单的 Nginx 反代配置：

```nginx
server {
    listen 443 ssl;
    server_name img.yourdomain.com;

    ssl_certificate     /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    client_max_body_size 10M;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 许可证

MIT License — 随便用，随便改。
