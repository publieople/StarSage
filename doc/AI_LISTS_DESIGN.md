# 设计文档: AI 驱动的仓库分类 (AI Lists)

**版本:** 1.0
**作者:** Cline

## 1. 概述

此功能旨在扩展 StarSage 的能力，允许用户通过自然语言指令，利用 AI 将其收藏的 GitHub 仓库自动分类到不同的“列表” (Lists) 中。用户可以在前端界面输入分类要求（例如，“帮我找到所有关于数据可视化的库”），后端将调用 AI 模型来分析并匹配相应的仓库，然后将结果保存为一个新的列表。

## 2. 数据库设计

为了支持列表功能，我们需要对现有的 SQLite 数据库进行扩展，增加两张新表：`lists` 和 `repository_lists`。

### 2.1. `lists` 表

这张表用于存储用户创建的每一个列表的信息。

**SQL 定义:**

```sql
CREATE TABLE IF NOT EXISTS lists (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    user_prompt TEXT, -- 存储用户创建此列表时的原始指令
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

- `id`: 列表的唯一标识。
- `name`: 列表的名称，可以由 AI 根据用户指令生成，或由用户指定。
- `description`: 对列表的简短描述。
- `user_prompt`: 保存用户输入的原始自然语言指令，用于追溯和未来可能的重新分类。
- `created_at`: 列表的创建时间。

### 2.2. `repository_lists` 表

这张表用于建立仓库 (`repositories`) 和列表 (`lists`) 之间的多对多关系。

**SQL 定义:**

```sql
CREATE TABLE IF NOT EXISTS repository_lists (
    repository_id INTEGER NOT NULL,
    list_id INTEGER NOT NULL,
    PRIMARY KEY (repository_id, list_id),
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE,
    FOREIGN KEY (list_id) REFERENCES lists(id) ON DELETE CASCADE
);
```

- `repository_id`: 指向 `repositories` 表的外键。
- `list_id`: 指向 `lists` 表的外键。
- `ON DELETE CASCADE`: 确保当一个仓库或一个列表被删除时，相关的链接记录也会被自动清除。

## 3. 后端实现

后端需要实现新的业务逻辑来处理 AI 分类任务，并提供相应的 API 端点供前端调用。

### 3.1. AI 分类服务

这是功能的核心。当收到用户的分类请求时，后端需要执行以下步骤：

1. **准备数据**: 从数据库中查询出所有仓库的 `id`, `full_name`, `description`, 和 `summary`。
2. **构建 Prompt**: 设计一个高效的 Prompt，将其发送给 AI 模型 (例如 Ollama)。这个 Prompt 需要包含：
    - **任务指令**: 清晰地告诉 AI 它的任务是什么。例如：“你是一个 GitHub 项目分类专家。请根据用户的要求，从下面的项目列表中找出所有符合条件的项目。只返回匹配项目的 ID 列表 (JSON 格式的数组)。”
    - **用户要求**: 嵌入用户的原始输入，例如：“帮我找到所有关于数据可视化的库”。
    - **项目数据**: 将所有仓库的信息格式化后附加到 Prompt 中。为了避免 Prompt 过长，可以考虑分批处理。
3. **调用 AI 模型**: 将构建好的 Prompt 发送给 AI Provider (如 Ollama)。
4. **解析结果**: AI 模型应返回一个包含匹配仓库 `id` 的 JSON 数组，例如 `[101, 205, 340]`。后端需要解析这个 JSON。
5. **存储结果**:
    - 在 `lists` 表中创建一个新条目，记录列表的名称和用户的原始指令。
    - 将 AI 返回的 `repository_id` 数组和新创建的 `list_id` 存入 `repository_lists` 表中。

### 3.2. API 端点设计

我们需要在 `internal/server/server.go` 中添加新的路由和处理函数。

- **`POST /api/lists`**: 创建一个新的列表。

  - **Request Body**: `{ "prompt": "用户的自然语言指令" }`
  - **Action**: 触发上述的 AI 分类服务。
  - **Response**: `{ "id": 123, "name": "数据可视化库", "count": 5 }` (返回新列表的信息)

- **`GET /api/lists`**: 获取所有已创建的列表。

  - **Response**: `[{ "id": 1, "name": "列表1", "count": 10 }, { "id": 2, "name": "列表2", "count": 25 }]`

- **`GET /api/lists/{id}`**: 获取特定列表下的所有仓库。

  - **Action**: 查询 `repository_lists` 表，并返回完整的仓库信息。
  - **Response**: `[ { ...repo1... }, { ...repo2... } ]` (标准的 `Repository` 对象数组)

- **`DELETE /api/lists/{id}`**: 删除一个列表。
  - **Action**: 从 `lists` 表和 `repository_lists` 表中删除相关记录。
  - **Response**: `204 No Content`

## 4. 前端实现

前端需要更新以支持列表的创建、浏览和管理。

### 4.1. UI/UX 设计

1. **列表面板**: 在主界面的侧边栏或顶部，增加一个“我的列表”面板，用于展示所有已创建的列表。
2. **创建列表表单**: 提供一个输入框和一个“创建”按钮。用户在输入框中填写分类要求，点击按钮后调用 `POST /api/lists` API。
3. **交互流程**:
    - 用户在输入框中输入 "所有和 k8s 相关的工具"。
    - 点击“创建”按钮，前端显示加载状态。
    - 前端调用 `POST /api/lists`。
    - 后端完成 AI 分类后，API 返回成功响应。
    - 前端刷新列表面板，显示出新的“k8s 相关工具”列表。
4. **浏览列表**:
    - 点击列表面板中的某个列表项 (例如“数据可视化库”)。
    - 前端调用 `GET /api/lists/{id}` API。
    - 主仓库展示区被清空，并渲染从 API 获取到的该列表下的仓库。

### 4.2. JavaScript (`script.js`) 更新

- 添加新的函数来调用上述 API 端点。
- 添加渲染列表面板的逻辑。
- 修改现有的 `renderRepos` 函数或创建新函数，以支持根据选择的列表来显示仓库。
- 处理加载和错误状态。

## 5. 实施步骤 (Checklist)

1. **[后端]** 更新 `internal/db/db.go`，添加 `lists` 和 `repository_lists` 表的创建逻辑。
2. **[后端]** 实现 AI 分类服务的核心逻辑。
3. **[后端]** 在 `internal/server/server.go` 中添加新的 API 端点和处理函数。
4. **[后端]** 优化数据库连接，改为使用全局连接池（在实现新功能的同时重构）。
5. **[前端]** 在 `index.html` 中添加列表面板和创建表单的 HTML 结构。
6. **[前端]** 在 `style.css` 中为新元素添加样式。
7. **[前端]** 在 `script.js` 中实现与新 API 的交互逻辑。
8. **[前端]** 实现前端的列表展示和筛选逻辑。
9. **[文档]** 更新 `README.md`，介绍新的 AI 分类功能。
