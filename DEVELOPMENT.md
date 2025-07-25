# StarSage 开发文档

本文档旨在为 StarSage 项目的开发者提供指导，帮助他们理解项目架构、参与后续开发或进行项目交接。

## 1. 项目概述

StarSage 是一个轻量级、功能强大的 GitHub Stars 整理和分析工具。它旨在通过 AI 技术，帮助用户自动化地管理、搜索和理解他们收藏的数千个 GitHub 项目。

核心目标是：低成本、可部署、轻量高效、兼容多种 AI API。

## 2. 当前项目状态

项目正处于 MVP（最小可行版本）的开发阶段。

**已实现功能:**

- **`login` 命令**:

  - 通过 GitHub OAuth Device Flow 进行安全认证。
  - 支持 `--proxy` 全局标志，可配置 HTTP/HTTPS 代理。
  - 自动将获取的 Token 存储在 `~/.config/starsage/config.yaml`。

- **`sync` 命令**:

  - 读取已保存的 Token 进行认证。
  - 完整获取用户所有的 Starred 项目列表（处理分页）。
  - 逐一获取每个项目的 `README.md` 文件内容。
  - 所有网络请求均包含重试逻辑，以应对不稳定网络。
  - 将项目元数据和 `README` 内容存入本地 SQLite 数据库 (`~/.config/starsage/stars.db`)。

- **`summarize` 命令**:
  - 命令框架及 `--provider` 和 `--model` 标志已创建。
  - 已实现从数据库读取待摘要仓库、调用 AI Provider、将结果写回数据库的完整逻辑。
  - **状态**: 等待本地 Ollama 模型下载完成后，即可进行端到端测试。

## 3. 项目架构

项目采用 Go 语言编写，核心是一个命令行工具（CLI），数据存储在本地的 SQLite 数据库中。

### 3.1. 目录结构与模块职责

```text
/cmd/starsage/       # CLI 命令入口 (main.go) 和各个子命令的定义
/internal/           # 项目内部逻辑，不作为库导出
  /ai/               # AI 抽象层和具体实现 (Ollama, OpenAI, etc.)
    - provider.go    # 定义了所有 AI Provider 都必须实现的通用接口
    - ollama.go      # Ollama 的具体实现
  /config/           # 配置管理 (基于 Viper)
    - config.go      # 处理配置文件的读写 (config.yaml)
  /db/               # 数据库交互
    - db.go          # 初始化数据库、建表、CRUD 操作
  /gh/               # 与 GitHub API 交互的所有逻辑
    - auth.go        # 处理 OAuth Device Flow 认证
    - client.go      # 封装带认证和代理的 HTTP 客户端，获取 Stars 和 README
```

### 3.2. 关键技术选型

- **语言**: Go 1.22+
- **CLI 框架**: `cobra`
- **配置管理**: `viper`
- **数据库**: `modernc.org/sqlite` (纯 Go 实现，无 CGO 依赖)
- **GitHub API**: Go 标准库 `net/http`

## 4. 如何构建和运行

### 4.1. 环境准备

1. **Go**: 确保已安装 Go 1.22 或更高版本。
2. **Git**: 用于版本控制。
3. **Ollama (可选，用于摘要)**:
   - 访问 [https://ollama.com/](https://ollama.com/) 下载并安装。
   - 运行 `ollama pull llama3:8b` (或其他模型) 来下载摘要所需的模型。

### 4.2. 首次配置

1. **创建 GitHub OAuth App**:
   - 访问 [https://github.com/settings/applications/new](https://github.com/settings/applications/new)。
   - 填写任意应用名称和主页/回调 URL (例如 `http://localhost`)。
   - 创建后，复制页面上的 **Client ID**。
2. **更新代码**:
   - 将复制的 Client ID 替换掉 `internal/gh/auth.go` 文件中 `clientID` 常量的值。

### 4.3. 运行命令

所有命令都在项目根目录下执行。

1. **下载依赖**:

   ```bash
   go mod tidy
   ```

2. **登录 (首次运行必须)**:

   ```bash
   # 如果需要代理
   go run ./cmd/starsage --proxy http://127.0.0.1:7890 login
   # 如果不需要代理
   go run ./cmd/starsage login
   ```

   根据提示在浏览器中完成授权。

3. **同步数据**:

   ```bash
   # sync 命令也会自动使用 --proxy 标志
   go run ./cmd/starsage --proxy http://127.0.0.1:7890 sync
   ```

4. **AI 摘要**:

   ```bash
   # 确保 Ollama 正在运行
   go run ./cmd/starsage summarize --provider=ollama --model=llama3:8b
   ```

## 5. 后续开发计划 (MVP)

- **`search` 命令**:

  - 实现基于 SQLite FTS5 的全文关键字搜索。
  - (可选) 集成 `sqlite-vss` 实现向量语义搜索。

- **`export` 命令**:

  - 实现将数据库内容导出为 Markdown 或静态 HTML 网站。

- **完善 `sync` 命令**:
  - 实现基于 `ETag` 的增量同步，避免每次都全量拉取。
