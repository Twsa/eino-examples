# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

这是 CloudWeGo Eino 框架的快速入门示例项目，展示了如何使用 Eino 构建各种 AI 应用场景。

**文档**: https://www.cloudwego.io/zh/docs/eino/overview/bytedance_eino_practice/
**核心依赖**: github.com/cloudwego/eino v0.4.7

## 项目结构

```
quickstart/
├── chat/                      # 基础 LLM 聊天示例（ChatModel + ChatTemplate）
├── todoagent/                 # Chain + ToolsNode 构建简单 Agent
└── eino_assistant/            # 完整的 RAG + Agent 应用（生产级）
    ├── cmd/
    │   ├── einoagent/         # HTTP 服务器（Hertz）
    │   ├── einoagentcli/      # 命令行交互模式
    │   └── knowledgeindexing/ # 知识库索引工具
    ├── pkg/                   # 共享包（工具、内存、Redis 连接）
    ├── eino/                  # Eino DevOps 生成的编排代码
    ├── data/                  # 数据存储目录
    └── eino_agent.json        # Agent 图定义配置
```

## 常用命令

### 环境准备

```bash
# 启动 Redis Stack（向量数据库）
docker-compose up -d

# 配置环境变量
export ARK_API_KEY=xxx              # 火山引擎 API Key
export ARK_CHAT_MODEL=ep-xxx        # 聊天模型（推荐 Doubao-pro-4k）
export ARK_EMBEDDING_MODEL=ep-xxx   # 向量模型（推荐 Doubao-embedding-large）
```

### 运行示例

```bash
# 基础聊天示例
cd chat && go run main.go

# TodoAgent 示例
cd todoagent && go run main.go

# Eino Assistant（HTTP 服务器）
cd eino_assistant
go run cmd/einoagent/main.go        # 访问 http://127.0.0.1:8080

# Eino Assistant（命令行模式）
cd eino_assistant
go run cmd/einoagentcli/main.go -id=会话ID

# 知识库索引
cd eino_assistant/cmd/knowledgeindexing
go run main.go
```

### 容器运行

```bash
cd eino_assistant
docker build --platform=linux/amd64 . -t eino-assistant:latest
docker run -p 8080:8080 \
  -e ARK_API_KEY=xxx \
  -e ARK_CHAT_MODEL=xxx \
  -e ARK_EMBEDDING_MODEL=xxx \
  eino-assistant:latest
```

## 核心架构概念

### Eino 组件化设计

Eino 采用高度模块化的组件设计：
- **ChatModel**: 大语言模型抽象（支持 ARK、OpenAI、Ollama）
- **Embedding**: 向量化模型
- **Retriever**: 检索器（RAG）
- **Indexer**: 索引器（文档入库）
- **Tool**: 工具接口
- **ChatTemplate**: 消息模板（支持变量替换和对话历史）

### 编排模式

- **Chain**: 线性编排（见 `todoagent/main.go`）
- **Graph**: DAG 图编排（见 `eino_assistant`）
- **ReAct Agent**: 推理-行动循环（Agent 自动决策调用工具）

### DevOps 代码生成

`eino/` 目录下的代码由 Eino DevOps 工具根据 JSON 配置自动生成：
- `eino_agent.json`: Agent 图定义
- 修改 JSON 后重新生成编排代码，无需手动编写

### Eino Assistant 数据流

```
用户输入 → 并行分支:
  ├─ InputToQuery (Lambda) → RedisRetriever → ChatTemplate
  └─ InputToHistory (Lambda) ↗
                                    ↓
                              ChatTemplate → ReactAgent → 结束
```

## 自定义工具开发

自定义工具的四种方式（见 `todoagent/main.go`）：

1. `utils.NewTool()`: 手动定义 ToolInfo
2. `utils.InferTool()`: 通过结构体标签自动推断
3. 实现 `InvokableTool` 接口
4. 使用官方工具（如 DuckDuckGo Search）

工具实现参考 `pkg/tool/task/task.go`。

## 可观测性

通过环境变量启用：
- **APMPlus**: `APMPLUS_APP_KEY`, `APMPLUS_REGION`
- **LangFuse**: `LANGFUSE_PUBLIC_KEY`, `LANGFUSE_SECRET_KEY`
- **CozeLoop**: `COZELOOP_API_TOKEN`, `COZELOOP_WORKSPACE_ID`

## 调试

- 设置 `DEBUG=true` 开启详细日志
- 查看 `log/eino.log` 文件
- Redis 数据管理界面：http://127.0.0.1:8001
