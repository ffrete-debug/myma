# ARK 服务器管理器

> ⚠️ **开发阶段提示**：本项目目前仍处于开发阶段，功能可能不完整或存在稳定性问题。建议仅用于测试环境，不建议在生产环境中使用。

[English](README.md) | [中文](README-zh.md)

- Linux上的 ARK 生存进化服务器管理工具。
- ARK 服务器自带 ArkApi 插件系统。

## 🎮 功能特性

### ✅ 已实现功能
- 🐳 每个ARK服务器运行在独立的Docker容器中
- 🔌 服务器自带ArkApi
- 🔄 服务器容器支持崩溃自动重启
- ⬆️ 第一次启动时自动更新服务端文件和Mod
- 💾 自动创建和管理Docker卷存储游戏数据
- 🖥️ 添加和管理多个 ARK 服务器
- ⚙️ 配置服务器设置和配置参数（GameUserSettings.ini、Game.ini、启动参数）
- ▶️ 一键启动/停止/重启服务器
- 🖼️ Docker镜像管理（拉取、更新、状态检查）
- 🔐 JWT认证和用户管理
- 📝 完整的API文档（Swagger）
- 🧩 **插件管理器** — 文件浏览器，支持拖拽上传、重命名、删除、创建文件夹
- 📝 **JSON/INI/配置文件编辑器** — 内联模态编辑器，支持 `.json`、`.ini`、`.txt`、`.cfg`、`.yaml`、`.xml`、`.conf`
- 📦 **Zip/Unzip** — 上传自动解压、手动解压、文件夹下载为ZIP
- 📋 **配置导入/导出** — 一键导入导出所有配置（GameUserSettings.ini + Game.ini + server_args）为单个JSON文件；每页标签页支持单独下载/导入
- 🌐 **i18n 国际化** — 英文和中文

### 🚧 待实现功能
- 🎮 RCON 命令执行
- 📊 服务器运行状态监控
- 🎨 Mod管理对接steam创意工坊
- 📋 服务器日志查看
- 💾 服务器存档及配置备份
- 🔍 工具版本更新检查
- ⚡ 可选更新服务端文件和Mod
- 🔄 容器镜像更新功能
- 🔌 MCP 支持
   
### 🚀 未来计划
- ☸️ 多主机管理，可能基于K8S实现
- 🌍 服务器收录网站，脱离糟糕的steam搜服
- 👥 玩家使用界面

## 🔒 安全提示

### ⚠️ JWT密钥配置（重要）

**在部署此应用程序之前，您必须配置一个强JWT密钥！**

#### 为什么这很重要？
- JWT（JSON Web Token）用于用户认证和会话管理
- 弱密钥或默认密钥允许攻击者伪造认证令牌
- 这可能导致**系统完全被攻破**和对所有服务器的未授权访问

#### 如何配置：

**1. 生成强随机密钥（推荐）：**
```bash
openssl rand -base64 48
```

**2. 设置环境变量：**

对于 Docker Compose 部署，编辑 `docker-compose.yml`：
```yaml
environment:
  - JWT_SECRET=your-generated-secret-here  # 替换为生成的密钥
```

对于直接部署：
```bash
export JWT_SECRET='your-generated-secret-here'
```

#### 安全要求：
- ✅ 最小长度：32 字符
- ✅ 使用加密随机生成
- ✅ 永远不要将密钥提交到版本控制
- ✅ 不同环境使用不同密钥（开发/测试/生产）
- ❌ 永远不要使用默认值如 "your-secret-key-here"
- ❌ 永远不要使用常见密码或字典单词

#### 验证：
应用程序将**拒绝启动**如果：
- JWT_SECRET 未设置
- JWT_SECRET 短于 32 字符
- JWT_SECRET 包含弱/常见密码模式

---

## 🚀 快速开始

### 🔧 系统要求

- 每个ARK服务器 8GB+ 内存 (推荐)
- 每个ARK服务器 10GB+ 磁盘空间

### 🔧 本地开发（Docker）

构建包含 Go 后端和 Next.js 前端的自定义镜像：
```bash
git clone https://github.com/21oramaster/ark-commander.git
cd ark-commander
docker build -t ark-commander-fixed:latest .
docker compose up -d
```

访问地址：`http://<your-ip>:3000`。默认登录：`admin` / `admin123`。

### 🐳 Docker容器化部署

拷贝docker-compose.yml，或直接复制：
```yml
version: '3.8'

services:
  ark-commander:
    image: tbro98/arkservercommander:latest
    container_name: ark-commander
    ports:
      - "8080:8080"
    environment:
      - JWT_SECRET=your-secret-key-here
      - DB_PATH=/data/ark_server.db
      - SERVER_PORT=8080
    volumes:
      - ./data:/data
      - /var/run/docker.sock:/var/run/docker.sock
    restart: unless-stopped
    privileged: true
```

```bash
sudo docker compose up -d
```

## 📖 使用说明

### 🆕 首次使用
1. 系统会自动跳转到初始化页面
2. 设置您的管理员账号和密码
3. 初始化完成后登录系统

### 🖥️ 管理服务器
1. 登录后点击"服务器管理"
2. 点击"添加服务器"创建新的服务器配置
3. 点击铅笔图标编辑服务器 — 在4个标签页中配置基本参数、GameUserSettings.ini、Game.ini 和启动参数

### 🧩 插件管理器
1. 在侧边栏导航到"插件管理"
2. 选择服务器，然后浏览、上传、编辑、重命名、删除或下载插件文件
3. ZIP文件上传时自动解压；使用解压按钮手动解压已有ZIP
4. 使用"下载为ZIP"按钮下载文件夹

### 📋 配置导入/导出
1. 打开服务器的编辑页面
2. 使用"导出所有配置"将所有设置下载为JSON
3. 使用"导入所有配置"从之前导出的JSON恢复
4. 每个标签页有单独的下载/导入按钮用于单文件操作

### 🗺️ 支持的地图
- The Island (孤岛)、The Center (中心岛)、Scorched Earth (焦土)
- Aberration (畸变)、Extinction (灭绝)、Valguero (瓦尔盖罗)
- Genesis (创世纪)、Genesis 2 (创世纪2)、Crystal Isles (水晶岛)
- Lost Island (失落岛)、Fjordur (峡湾)

## ❓ 常见问题

### ❓ Q: 如何备份ARK服务器数据？
A: 服务器数据存储在Docker卷 `ark-server-<服务器编号>` 中。可以手动备份，或使用导出配置功能备份设置。

### ❓ Q: 如何查看ARK服务器日志？
A: 目前需要在容器内查看日志。日志查看功能计划在后续更新中实现。

### ❓ Q: 如何更新ARK服务器镜像？
A: 转到首页，点击服务器卡片上的"检查更新"。系统会比较本地和远程镜像摘要并提示更新。

### ❓ Q: JWT_SECRET 配置错误怎么办？
A: 如果应用启动失败并提示 JWT_SECRET 错误，请确保：
- JWT_SECRET 已在环境变量中设置
- 密钥长度至少 32 字符
- 使用 `openssl rand -base64 48` 生成强随机密钥


### 🖼️ ARK服务器镜像
- 本系统使用 `tbro98/ase-server:latest` 镜像来运行ARK服务器
- 镜像源地址: [ASE-Server-Docker](https://github.com/tbro199803/ASE-Server-Docker)

## 📸 界面展示
![](./docs/zh/images/img_servers.png)
![](./docs/zh/images/ima_base.png)
![](./docs/zh/images/img_GameUserSettings.png)
![](./docs/zh/images/img_GameIni.png)
![](./docs/zh/images/img_args.png) 
