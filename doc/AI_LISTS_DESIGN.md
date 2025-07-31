# AI 驱动的 GitHub Star 列表分类功能设计文档

## 1. 功能概述

本功能旨在为 StarSage 引入一个由 AI 驱动的自动化分类系统。用户可以通过自然语言指令，要求 AI 将其收藏的 GitHub 项目（Stars）整理成不同的、有意义的列表。例如，用户可以要求“创建一个名为‘数据可视化工具’的列表，包含所有与图表、图形和数据可视化相关的项目”，AI 将自动分析数据库中的所有项目，并将符合条件的项目归入该列表。

这将极大地提升 StarSage 的智能化水平，帮助用户从数千个收藏中高效地组织和发现信息。

## 2. 用户交互流程

1. **入口**: 在前端 Web 界面的主导航栏增加一个“智能列表” (AI Lists) 选项。
2. **列表视图**: 点击“智能列表”后，用户进入列表管理页面，该页面展示所有已创建的列表。
3. **创建新列表**:
    - 页面上有一个“创建新列表”的按钮。
    - 点击后，弹出一个对话框，包含两个输入框：
      - **列表名称 (List Name)**: 用户为新列表命名，例如 `Go Web 框架`。
      - **分类指令 (Classification Prompt)**: 用户输入自然语言指令，描述该列表应包含哪些项目。例如 `所有用于构建 Web 应用的 Go 语言框架和库`。
4. **AI 处理**:
    - 用户提交后，前端将“列表名称”和“分类指令”发送到后端新的 API 端点。
    - 后端启动一个异步任务，调用 AI 模型处理该指令。
    - AI 模型分析数据库中所有仓库的 `full_name`, `description`, 和 `summary`，然后返回一个符合指令描述的仓库 ID 列表。
    - 后端将这个仓库 ID 列表与列表名称关联，并存入数据库。
5. **查看与交互**:
    - 列表创建成功后，会出现在列表管理页面。
    - 用户可以点击某个列表，查看其中包含的所有项目。
    - （未来）用户可以对列表内的项目进行微调（增/删），或者重新运行 AI 分类。

## 3. 后端设计

### 3.1. 数据库变更 (`internal/db/db.go`)

为了支持列表功能，我们需要新增两张表：

1. **`lists` 表**: 用于存储用户创建的列表信息。

    ```sql
    CREATE TABLE IF NOT EXISTS lists (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE,
        prompt TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    ```

2. **`list_repositories` 关联表**: 用于存储列表和仓库之间的多对多关系。

    ```sql
    CREATE TABLE IF NOT EXISTS list_repositories (
        list_id INTEGER NOT NULL,
        repository_id INTEGER NOT NULL,
        PRIMARY KEY (list_id, repository_id),
        FOREIGN KEY (list_id) REFERENCES lists(id) ON DELETE CASCADE,
        FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE
    );
    ```

### 3.2. API 端点 (`internal/server/server.go`)

需要新增以下 API 端点：

- **`POST /api/lists`**: 创建一个新的智能列表。
  - **Request Body**: `{ "name": "Go Web 框架", "prompt": "所有用于构建 Web 应用的 Go 语言框架和库" }`
  - **Response**: `{ "task_id": "some-unique-id" }` (因为 AI 处理是异步的)
- **`GET /api/lists`**: 获取所有已创建的列表。
  - **Response**: `[{ "id": 1, "name": "Go Web 框架", "repo_count": 15 }, ...]`
- **`GET /api/lists/{id}`**: 获取特定列表及其包含的所有仓库。
  - **Response**: `{ "id": 1, "name": "Go Web 框架", "repositories": [ ... repo objects ... ] }`
- **`GET /api/tasks/{task_id}`**: 查询异步任务的状态。
  - **Response**: `{ "status": "processing" | "completed" | "failed", "error": "..." }`

### 3.3. AI 交互逻辑 (`internal/ai/classifier.go`)

需要创建一个新的 `classifier.go` 文件。

1. **`ClassifyRepositories` 函数**:

    - **输入**: 分类指令 (prompt) 和所有待分类的仓库信息 (ID, name, description, summary)。
    - **核心逻辑**:

      1. 构造一个发送给 AI 的 Prompt。这个 Prompt 需要精心设计，以确保 AI 能够理解任务并以期望的格式返回结果。
      2. Prompt 示例:

          ```prompt
          你是一个精准的软件项目分类助手。
          我会给你一个分类任务的描述，以及一个 JSON 格式的项目列表。
          请仔细阅读每个项目的描述，并判断它是否符合分类任务的要求。

          分类任务: "{用户的分类指令}"

          项目列表如下:
          [
            { "id": 101, "name": "gin-gonic/gin", "description": "Gin is a HTTP web framework written in Go (Golang)..." },
            { "id": 102, "name": "d3/d3", "description": "Bring data to life with SVG, Canvas and HTML..." }
          ]

          请只返回一个 JSON 数组，其中仅包含符合分类任务要求的项目 ID。
          例如: [101]
          ```

      3. 由于一次传递给 AI 的 token 数量有限，需要对项目列表进行分块 (chunking) 处理。
      4. 合并所有块的返回结果，得到最终的仓库 ID 列表。

    - **输出**: `[]int64` (符合条件的仓库 ID 列表)

## 4. 前端设计 (`frontend/`)

1. **`index.html`**:
    - 在主导航栏添加“智能列表”链接。
    - 创建一个新的 `<section id="lists-view">` 用于显示列表管理界面，默认隐藏。
2. **`script.js`**:
    - 添加路由逻辑，用于在“仓库视图”和“列表视图”之间切换。
    - 实现 `fetchLists` 函数，用于从 `/api/lists` 获取并渲染列表。
    - 实现创建新列表的表单交互逻辑，点击提交后调用 `POST /api/lists`。
    - 实现查看列表详情的逻辑，点击列表后调用 `GET /api/lists/{id}` 并展示其中的仓库。
3. **`style.css`**:
    - 为列表视图、列表项、弹出对话框等新元素添加样式。

## 5. 实现步骤

1. **[后端]** 在 `internal/db/db.go` 中添加 `lists` 和 `list_repositories` 表的创建逻辑，并编写相应的增删改查函数。
2. **[后端]** 创建 `internal/ai/classifier.go` 文件，实现 AI 分类逻辑，包括 Prompt 构造和分块处理。
3. **[后端]** 在 `internal/server/server.go` 中实现新的 API 端点 (`/api/lists`, `/api/tasks/...`)。
4. **[前端]** 修改 `index.html`，添加新视图的骨架。
5. **[前端]** 大幅修改 `script.js`，实现视图切换、列表数据的获取/渲染、创建列表的交互等。
6. **[前端]** 更新 `style.css` 以美化新界面。
7. **[文档]** 更新 `README.md` 和 `DEVELOPMENT.md`，加入关于新功能的说明。
