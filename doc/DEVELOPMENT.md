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

- **`search` 命令**:
  - **状态**: 已完成。
  - 基于 SQLite FTS5 实现了对仓库名称、描述和 README 的全文关键字搜索。
  - 搜索结果按相关性（BM25）排序。

- **完善 `sync` 命令**:
  - **状态**: 已完成。
  - 基于 `ETag` 实现了对 `README` 文件的增量同步，避免了对未更改文件的不必要下载。

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

- **`serve` 命令 (Web UI)**:
  - **状态**: 已完成 (只读查看和搜索)。
  - **后端**:
    - 创建了 `serve` 子命令，用于在 `http://localhost:8080` 启动本地服务器。
    - 实现了 `GET /api/repositories` 端点，用于获取所有仓库数据。
    - 服务器可以托管 `frontend` 目录下的静态文件。
  - **前端**:
    - 创建了一个简单的单页应用，用于显示所有仓库。
    - 实现了客户端的实时搜索功能。
  - **未来计划**:
    - 实现完整的 CRUD (创建、更新、删除) 功能。
    - 在 API 和前端实现分页。

- **`export` 命令**:

  - 实现将数据库内容导出为 Markdown 或静态 HTML 网站。

## 6. 开发经验与关键决策

在 MVP 的开发过程中，我们遇到并解决了一系列关键问题。这些决策对项目的健壮性、易用性和可维护性至关重要。

1. **应对网络不稳定性**:

    - **问题**: 在初始阶段，频繁遇到连接 Go 模块代理和 GitHub API 的网络超时 (`timeout`) 和连接中断 (`EOF`) 错误。
    - **决策**:
      - **配置 Go Proxy**: 通过 `go env -w` 将 Go 模块代理切换到国内镜像 (`goproxy.cn`)，解决了依赖下载问题。
      - **实现应用内代理**: 添加了全局的 `--proxy` 标志，允许用户通过自己的代理服务器访问 GitHub API，从根本上解决了网络可访问性问题。
      - **增加请求重试**: 为所有对 GitHub API 的调用（获取列表、获取 README）都增加了自动重试逻辑，大大增强了程序在网络抖动环境下的成功率。

2. **移除 CGO 依赖**:

    - **问题**: 最初选用的数据库驱动 `mattn/go-sqlite3` 依赖 CGO，要求开发环境中必须安装和配置 C 语言编译器，给 Windows 用户带来了额外的环境配置负担。
    - **决策**: 果断将数据库驱动更换为纯 Go 实现的 `modernc.org/sqlite`。此举完全移除了项目对 CGO 的依赖，使得项目回归到 Go “跨平台、无依赖”的哲学，极大地简化了编译和部署流程。

3. **处理数据库问题**:

    - **问题 1**: 在为已有数据的表添加 FTS5 全文搜索索引后，再次写入时出现 `database disk image is malformed` 错误，导致数据库损坏。
    - **决策 1**: 添加了 `starsage db reset` 命令，提供了一个简单可靠的方式来删除损坏的数据库，以便重新开始。
    - **问题 2**: `search` 命令在初次同步后无法返回结果。
    - **诊断 2**: 原因是 FTS 索引是在数据插入后才创建的，导致索引为空。通过 `UPSERT` 操作（`sync` 命令）触发数据库的 `UPDATE` 触发器，成功地为存量数据建立了索引。
    - **问题 3**: `search` 命令在读取未被摘要的 `summary` 字段时，因 `NULL` 值无法直接转换为 `string` 而崩溃。
    - **决策 3**: 在数据库扫描逻辑中，使用 `sql.NullString` 等可处理 `NULL` 值的数据类型，增强了代码的健壮性。

4. **提升开发与测试效率**:
    - **问题**: 每次 `sync` 都需要处理全部（600+）仓库，调试周期长，效率低下。
    - **决策**: 根据用户反馈，添加了全局的 `--limit` 标志。该标志可用于 `sync`, `summarize`, `search` 等多个命令，允许开发者只处理少量数据，极大地提升了开发和调试的效率。
