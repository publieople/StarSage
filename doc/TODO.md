# StarSage 开发计划

本文档整合了项目的所有设计文档，形成一个统一的开发计划，指导项目的进一步开发和维护。

## 1. 项目概述

### 1.1 核心问题

GitHub Stars 本质上是一个被动的书签系统。开发者收藏了成百上千个项目后，这个列表很快就变成了一个"只写不读"的信息坟场。当需要时，我们很难记起当初为何收藏某个项目，它解决了什么问题，或者它与其他类似项目的区别。一个巨大的潜在知识库被闲置了。

### 1.2 产品使命

将开发者的 GitHub Stars 从一个被动的书签列表，转变为一个活跃的、个性化的、可搜索的知识库。StarSage 旨在帮助开发者重新发现、理解并利用他们自己收藏的项目中所蕴含的集体智慧。

### 1.3 独特定位

StarSage 不仅仅是另一个书签管理器，它是一个在您本地运行的 **AI 知识助理**。它通过数据同步、本地 AI 处理（用于摘要和分类）和强大的搜索功能，为您创建一个私有的、个性化的知识发现引擎。我们的核心差异点在于 **深度理解和智能组织**，而不仅仅是存储。

## 2. 目标用户

### 2.1 用户画像

**主要用户："博学开发者 / 终身学习者"**。这类开发者常常在多种语言和技术栈之间切换。他们不断探索新工具、新库和新技术，频繁地使用 Star 来"稍后阅读"，但在真正需要这些信息时，却难以有效地组织和检索。

### 2.2 核心痛点

- "我记得我收藏过一个解决这个问题的库，但想不起名字了。"
- "我收藏了 10 个不同的 Web 框架，哪个最适合我的新项目？"
- "我想找到我所有收藏过的 AI 相关工具，但 GitHub 的搜索功能太宽泛了。"
- "我的收藏夹一团糟，但我没时间手动给它们打标签分类。"

## 3. 功能特性

### 3.1 已实现功能

1. **安全认证**: 通过 GitHub OAuth Device Flow 进行安全认证，令牌存储在本地。
2. **全量同步**: 一键同步所有 GitHub Stars，包括项目元数据和 README 文件。
3. **AI 摘要**: 使用本地或远程 AI 模型（当前支持 Ollama）为项目 README 生成精炼摘要。
4. **全文搜索**: 基于 SQLite FTS5 的高性能全文搜索，快速在名称、描述和 README 中找到需要的项目。
5. **代理支持**: 内置代理支持，轻松应对各种网络环境。

### 3.2 计划中功能

1. **主题发现**: AI 自动分析整个收藏库，并提炼出关键主题标签（如 `go`, `web-framework`, `data-visualization`）。
2. **知识库问答**: 允许用户直接向自己的知识库提问，例如："我收藏的哪个 Go 库最适合做 JWT 认证？"
3. **增强的后端功能**:
   - 过滤器功能：允许用户根据不同的属性（如语言、标签、星标数等）过滤仓库列表
   - 排序方式：支持多种排序方式（如按星标数、创建时间、更新时间等）
   - 标签列：实现类似 Notion 的多选属性，允许用户为仓库添加自定义标签
   - AI 交互：增强 AI 功能，支持更自然的交互方式
4. **完整的后端增删改查功能**:
   - 仓库管理：更新仓库信息、从本地数据库中删除仓库
   - 列表管理：更新列表信息、删除列表、手动添加/移除仓库到列表
   - 主题管理：获取所有主题、获取特定主题下的所有仓库

## 4. 技术架构

### 4.1 系统概览

系统采用在本地运行的客户端-服务器架构。

- **后端服务器 (Go)**: 独立的 Go 应用程序，负责处理所有核心逻辑
- **数据流**: `GitHub API -> Go Backend -> SQLite DB -> Go Backend`

### 4.2 技术选型

- **后端**: Go - 高性能，适合构建 CLI 和 Web 服务
- **数据库**: SQLite with FTS5 - 轻量、本地、无需配置，全文搜索性能优异
- **AI 集成**: Ollama - 强大的本地大语言模型支持，符合隐私优先原则

### 4.3 数据模型扩展

为支持新功能，需要在现有表的基础上新增以下数据表：

```sql
-- 标签表
CREATE TABLE IF NOT EXISTS tags (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    color TEXT
);

-- 仓库标签关联表
CREATE TABLE IF NOT EXISTS repository_tags (
    repository_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (repository_id, tag_id),
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- 主题表
CREATE TABLE IF NOT EXISTS topics (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

-- 项目和主题的多对多关系表
CREATE TABLE IF NOT EXISTS repository_topics (
    repository_id INTEGER NOT NULL,
    topic_id INTEGER NOT NULL,
    confidence_score REAL, -- AI 对该分类的置信度
    PRIMARY KEY (repository_id, topic_id),
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE,
    FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE
);

-- 为仓库表添加时间字段
ALTER TABLE repositories ADD COLUMN created_at TIMESTAMP;
ALTER TABLE repositories ADD COLUMN updated_at TIMESTAMP;

-- 为仓库表添加向量嵌入字段
ALTER TABLE repositories ADD COLUMN embedding BLOB;
```

## 5. API 设计

### 5.1 标签管理 API

- GET /api/tags - 获取所有标签
- POST /api/tags - 创建新标签
- PUT /api/tags/:id - 更新标签
- DELETE /api/tags/:id - 删除标签

### 5.2 仓库标签关联 API

- POST /api/repositories/:id/tags - 为仓库添加标签
- DELETE /api/repositories/:id/tags/:tagId - 移除仓库的标签

### 5.3 仓库管理 API

- PUT /api/repositories/:id - 更新仓库信息
- DELETE /api/repositories/:id - 从本地数据库中删除仓库

### 5.4 列表管理 API

- PUT /api/lists/:id - 更新列表信息
- DELETE /api/lists/:id - 删除列表
- POST /api/lists/:id/repositories - 手动添加仓库到列表
- DELETE /api/lists/:id/repositories/:repoId - 从列表中移除仓库

### 5.5 主题管理 API

- GET /api/topics - 获取所有主题
- GET /api/topics/:id/repositories - 获取特定主题下的所有仓库

### 5.6 增强的查询 API

- GET /api/repositories?filter=...&sort=... - 支持过滤和排序参数的仓库查询
  - `language` - 按语言过滤
  - `tag` - 按标签过滤
  - `min_stars`/`max_stars` - 按星标数范围过滤
  - `sort` - 排序方式
  - `order` - 排序顺序（asc/desc）

## 6. 后端功能实现

### 6.1 功能模块

1. **标签管理模块**:

   - 标签创建、更新、删除功能
   - 仓库与标签关联管理

2. **过滤器模块**:

   - 语言过滤功能
   - 标签过滤功能
   - 星标数范围过滤功能

3. **排序模块**:

   - 多种排序方式支持
   - 排序顺序控制

4. **AI 交互功能**:
   - AI 标签建议功能
   - 自然语言过滤接口

## 7. 开发路线图

### 7.1 v1.0: 智能组织器

- **主题**: 夯实基础，交付"将收藏变为有序知识库"的核心价值
- **核心目标**: 完善并稳定所有"必须有"的功能（数据同步、AI 摘要、统一搜索、AI 智能列表）

### 7.2 v1.1: 探索与发现

- **主题**: 增强知识的探索与发现能力
- **核心目标**: 实现"应该有"的功能：**主题发现**

### 7.3 v2.0: 知识助理

- **主题**: 将工具升级为真正的对话式知识助理
- **核心目标**: 实现"可以有"的功能：**知识库问答**

## 8. 开发计划

### 8.1 第一阶段：数据库和后端 API 扩展

**目标**：实现标签功能的数据库支持和基础 API

**任务列表**：

1. 扩展数据库模式，添加标签相关表
2. 实现标签管理的数据库操作函数
3. 实现仓库标签关联的数据库操作函数
4. 扩展仓库查询功能，支持过滤和排序
5. 添加标签管理 API 端点
6. 添加仓库标签关联 API 端点
7. 扩展仓库查询 API，支持过滤和排序参数

### 8.2 第二阶段：AI 功能增强

**目标**：实现 AI 驱动的标签建议和自然语言过滤

**任务列表**：

1. 实现 AI 标签建议功能
2. 实现自然语言过滤器
3. 添加 AI 交互接口
4. 集成 AI 功能到现有后端中

### 8.3 第三阶段：测试和优化

**目标**：完善功能，修复问题，优化用户体验

**任务列表**：

1. 功能测试
2. 性能优化
3. 用户体验优化
4. 文档更新

## 9. 验收标准

1. 用户能够为仓库添加、编辑、删除标签
2. 用户可以通过标签、语言、星标数等条件过滤仓库列表
3. 用户可以选择不同的排序方式对仓库列表进行排序
4. AI 能够为仓库提供标签建议
5. 用户可以通过自然语言进行仓库过滤
6. 系统性能满足正常使用需求

## 10. 风险评估和缓解措施

### 10.1 技术风险

1. **数据库性能问题**：

   - 风险：大量标签和关联可能导致查询性能下降
   - 缓解：添加适当的索引，优化查询语句

2. **API 性能问题**：
   - 风险：大量仓库和标签可能导致 API 响应变慢
   - 缓解：实现分页和缓存机制，优化查询逻辑

### 10.2 用户体验风险

1. **接口复杂化**：

   - 风险：新增功能可能导致 API 接口过于复杂
   - 缓解：提供清晰的 API 文档，合理组织接口结构

2. **学习成本增加**：
   - 风险：新功能可能增加用户学习成本
   - 缓解：提供使用示例和帮助文档
