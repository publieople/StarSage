# 需求

1. 功能
   • 一键读取用户所有 GitHub Stars（增量/全量）
   • 本地持久化（可断点续传、可离线查看）
   • 多 AI 渠道兼容：OpenAI、Claude、Gemini、Ollama、本地 GGUF、云函数/Serverless 等
   • 可插拔“整理/摘要/标签”策略（Prompt 模板化、函数调用）
   • 轻量全文或向量搜索（关键词、语义、混合检索）
   • 导出：静态网站 / Markdown / JSON / RSS
   • 极低运行成本：≤ 128 MB RAM、≤ 100 MB 磁盘即可跑 1 万级 Stars
   • 一键部署：Docker / 单二进制 / Serverless（Vercel、Cloudflare Workers、Fly.io）

2. 非功能
   • 纯用户侧处理，不存敏感 Token
   • 无状态或极简状态，可水平扩展
   • 100% 开源、MIT 或 Apache-2.0

技术选型（最小可行 + 可演进）
────────────────

1. 语言与框架
   • Go 1.22：单二进制、静态链接、跨平台、内存占用低
   • 备选：Rust（更极致，但编译链重）或 Python（生态好，但镜像大）。
   • CLI 框架：cobra（Go）或 clap（Rust）

2. 数据层
   • SQLite（单文件，无依赖，FTS5 全文索引）
   • 向量检索：
   – 本地轻量：sqlite-vss（SQLite 扩展，基于 Faiss，<15 MB）
   – 云/容器：qdrant-lite（Docker 镜像 40 MB）
   • 缓存/队列：BadgerDB（Go KV，纯内存映射，零 GC）

3. GitHub 数据抓取
   • REST v3（/user/starred）+ ETag 增量同步
   • 速率限制：GraphQL 批量 100 条/次，或 REST + conditional requests
   • 并发：Go routine 池 20-50 并发即可

4. AI 接入层
   • 统一接口：定义 `Provider` interface
   • 内置适配器：
   – openai-compatible（OpenAI、Together、Anyscale、One-API 等）
   – anthropic、googleai、ollama（HTTP）
   • Prompt 模板：Go text/template 或 Rust askama，支持 YAML 热加载
   • 函数调用：若 AI 支持，可自动提取 tag / 摘要 / tech-stack

5. 部署形态
   • CLI：单二进制，下载即用
   • Docker：scratch 镜像 <15 MB
   • Serverless：
   – Cloudflare Workers + D1（SQLite）
   – Fly.io 免费 tier 256 MB RAM 足够
   • 静态站：hugo-book 主题，或纯 Go 内嵌 html/template 生成

6. 开发加速包
   • 配置：viper / figment（YAML/ENV）
   • 日志：zerolog（Go）或 tracing（Rust）
   • 测试：testify + dockertest（CI 里起容器测向量库）
   • 打包：goreleaser（自动发版 + brew tap）

最小可行版本（MVP）步骤
────────────────

1. 初始化：go mod init star-sage
2. CLI：star-sage login（OAuth Device Flow 拿 token）→ star-sage sync（写 SQLite）
3. AI 摘要：star-sage summarize --provider=ollama --model=llama3:8b
4. 搜索：star-sage search "microservice go"
5. 导出：star-sage export --format=html --out=./site

目录结构（Go 示例）

```text
/cmd/starsage        // main.go
/internal/db         // SQLite + vss
/internal/gh         // GitHub client
/internal/ai         // Provider interface + 适配器
/internal/search     // 关键词 + 向量
/internal/export     // 静态站生成
/config              // prompt.yml
```

性能与成本预估
────────────────
• 1 万 Stars → SQLite 数据库 ~30 MB
• 每条 README 平均 8 KB → 向量 384 维 float32 ≈ 15 MB
• 总结 1000 个项目（1k token/项目）→ 1M tokens，OpenAI GPT-3.5 约 0.2 美元
• 本地 Ollama 8B 模型 4-bit → 4.7 GB，但只需推理一次，可关机后纯本地文件

下一步可扩展
────────────────
• Web UI：htmx + Alpine.js，无需 React
• 多用户：加 JWT + 多租户 SQLite 分库
• 插件系统：WASM 或 Lua 沙箱
• 自动 PR：发现 Stars 更新 → AI 重写 README 摘要 → 发 PR 到用户 blog repo

这样即可在 1-2 周内做出一个“低成本、可部署、轻量高效、兼容各种 AI API 的 GitHub Stars 整理器”。
