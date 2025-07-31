# StarSage

[![Go Report Card](https://goreportcard.com/badge/github.com/publieople/StarSage)](https://goreportcard.com/report/github.com/publieople/StarSage)

**StarSage** 是一个轻量级、功能强大的命令行工具，旨在通过 AI 技术，帮助您自动化地管理、搜索和理解您收藏的数千个 GitHub 项目。

告别无尽的滚动和遗忘，让您的 GitHub Stars 真正为您所用！

## ✨ 功能特性

- **安全认证**: 通过 GitHub OAuth Device Flow 进行安全认证，令牌存储在本地。
- **全量同步**: 一键同步您所有的 GitHub Stars，包括项目元数据和 `README` 文件。
- **AI 摘要**: 使用本地或远程 AI 模型（当前支持 Ollama）为项目 `README` 生成精炼摘要。
- **全文搜索**: 基于 SQLite FTS5 的高性能全文搜索，快速在名称、描述和 `README` 中找到您需要的项目。
- **智能列表 (AI Lists)**: 在 Web 界面中，通过自然语言指令（例如“所有关于数据可视化的库”）创建智能列表，AI 会自动为您分类和组织项目。
- **Web 用户界面**: 通过 `serve` 命令启动一个本地 Web 服务器，提供一个简洁的界面来浏览、搜索和管理您的 Stars。
- **代理支持**: 内置 `--proxy` 标志，轻松应对各种网络环境。

## 🚀 安装与使用

### 1. 环境准备

- **Go**: 确保已安装 Go 1.22 或更高版本。
- **Git**: 用于克隆本项目。
- **Ollama (可选)**: 如果您希望使用 AI 摘要功能，请访问 [ollama.com](https://ollama.com/) 下载并安装，然后拉取一个模型，例如 `ollama pull llama3:8b`。

### 2. 安装

```bash
# 克隆项目
git clone https://github.com/publieople/StarSage.git
cd StarSage

# 下载依赖
go mod tidy

# (可选) 构建二进制文件
go build -o starsage ./cmd/starsage
# 您可以将 starsage 移动到您的 PATH 路径下，方便全局使用
```

### 3. 首次配置

在使用之前，您需要一个 GitHub OAuth App 的 Client ID。

1. 访问 [GitHub 应用创建页面](https://github.com/settings/applications/new)。
2. 填写任意应用名称和主页/回调 URL (例如 `http://localhost`)。
3. 创建后，复制页面上的 **Client ID**。
4. **重要**: 在 `internal/gh/auth.go` 文件中，将 `clientID` 常量的值替换为您刚刚复制的 Client ID。

### 4. 使用命令

所有命令都在项目根目录下执行（或通过构建好的二进制文件执行）。

a. 登录 (首次运行必须)

```bash
# 如果不需要代理
go run ./cmd/starsage login

# 如果需要通过代理
go run ./cmd/starsage --proxy http://127.0.0.1:7890 login
```

根据提示在浏览器中完成授权，StarSage 会自动保存您的令牌。

b. 同步数据

```bash
# 该命令会自动使用您在登录时配置的代理
go run ./cmd/starsage sync
```

此过程可能会花费一些时间，具体取决于您收藏项目的数量。

c. AI 摘要

```bash
# 确保您的 Ollama 服务正在运行
go run ./cmd/starsage summarize --provider=ollama --model=llama3:8b
```

您可以使用 `--limit` 标志来限制本次处理的项目数量，例如 `--limit 10`。

d. 搜索仓库

```bash
# 搜索包含 "data visualization" 的项目
go run ./cmd/starsage search data visualization

# 限制返回结果数量
go run ./cmd/starsage search "data visualization" --limit 5
```

e. 启动 Web 界面

```bash
# 启动服务器 (默认端口 8080)
go run ./cmd/starsage serve

# 使用指定端口
go run ./cmd/starsage serve --port 9090
```

然后，您可以在浏览器中打开 `http://localhost:8080` (或您指定的端口) 来访问 Web 界面。在 Web 界面中，您可以：

- 浏览和搜索所有已同步的仓库。
- 切换到“AI 列表”视图，创建和查看由 AI 自动分类的项目列表。

## 🛠️ 未来计划

- **`export` 命令**: 实现将数据库内容导出为 Markdown 或静态 HTML 网站。
- **更多 AI 支持**: 增加对 OpenAI、Gemini 等更多 AI 提供商的支持。

## 🤝 贡献

欢迎提交 Pull Requests 或 Issues 来帮助改进 StarSage！
