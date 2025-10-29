# AI 聊天助手

基于 Gin 框架的 HTTP 服务，提供 AI 聊天功能，支持多种模型选择、Markdown 渲染、问答记录和知识库管理。

## 功能特性

- 🚀 基于 Gin 框架的 HTTP 服务
- 🤖 支持多种 AI 模型选择
- 📝 支持 Markdown 渲染和展示
- ⚙️ YAML 配置文件支持
- 🎨 现代化的 Web 界面
- 📚 **知识库管理** - 将有价值的问答保存到本地知识库
- 📋 **问答记录** - 自动记录最近5次问答
- 🏷️ **标签系统** - 为知识库条目添加标签分类
- 🔍 **知识库浏览** - 独立的知识库展示页面

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 配置设置

编辑 `config.yaml` 文件：

```yaml
api:
  base_url: "https://api.openai.com/v1"
  api_key: "your-api-key-here"

server:
  port: ":8080"
  host: "localhost"

models:
  default: "claude-4.5-sonnet"
  available:
    - "claude-4.5-sonnet"
    - "z-ai/glm-4.6"
    - "deepseek/deepseek-v3.2-exp-thinking"
```

### 3. 运行程序

```bash
go run ai.go
```

或者编译后运行：

```bash
go build -o ai-assistant ai.go
./ai-assistant
```

### 4. 访问服务

- **主聊天页面**: http://localhost:8080
- **知识库页面**: http://localhost:8080/knowledge

## API 接口

### POST /api/chat

发送聊天请求

**请求体：**
```json
{
  "message": "你好，请介绍一下自己",
  "model": "claude-4.5-sonnet"
}
```

**响应：**
```json
{
  "response": "你好！我是一个AI助手...",
  "model": "claude-4.5-sonnet"
}
```

### GET /api/models

获取可用模型列表

**响应：**
```json
{
  "default": "claude-4.5-sonnet",
  "available": [
    "claude-4.5-sonnet",
    "z-ai/glm-4.6",
    "deepseek/deepseek-v3.2-exp-thinking"
  ]
}
```

### GET /api/recent

获取最近5次问答记录

**响应：**
```json
{
  "recent_qas": [
    {
      "id": 1,
      "question": "你好",
      "answer": "你好！我是AI助手...",
      "model": "claude-4.5-sonnet",
      "timestamp": "2025-10-22T22:10:00Z"
    }
  ]
}
```

### POST /api/knowledge/add

将问答记录添加到知识库

**请求体：**
```json
{
  "record_id": 1,
  "title": "AI助手介绍",
  "tags": "AI,介绍,助手"
}
```

**响应：**
```json
{
  "message": "已成功添加到知识库",
  "item": {
    "id": 1,
    "title": "AI助手介绍",
    "content": "你好！我是AI助手...",
    "model": "claude-4.5-sonnet",
    "timestamp": "2025-10-22T22:10:00Z",
    "tags": ["AI", "介绍", "助手"]
  }
}
```

### GET /api/knowledge

获取知识库内容

**响应：**
```json
{
  "knowledge_base": [
    {
      "id": 1,
      "title": "AI助手介绍",
      "content": "你好！我是AI助手...",
      "model": "claude-4.5-sonnet",
      "timestamp": "2025-10-22T22:10:00Z",
      "tags": ["AI", "介绍", "助手"]
    }
  ]
}
```

### DELETE /api/knowledge/:id

删除知识库条目

**响应：**
```json
{
  "message": "已删除知识库条目"
}
```

## 配置说明

### config.yaml 配置项

- `api.base_url`: API 基础 URL (支持 OpenAI、Claude 等)
- `api.api_key`: API 密钥
- `server.port`: 服务端口
- `server.host`: 服务主机
- `models.default`: 默认模型
- `models.available`: 可用模型列表

## 技术栈

- **后端**: Go + Gin
- **前端**: HTML + CSS + JavaScript
- **Markdown 渲染**: marked.js
- **配置**: YAML

## 项目结构

```
ai-chat-assistant/
├── ai.go                    # 主程序文件
├── config.yaml             # 配置文件
├── data/                   # 数据存储目录
│   ├── knowledge.json     # 知识库数据文件
│   └── recent_qas.json    # 最近问答数据文件
├── templates/              # 模板目录
│   ├── index.html         # 主聊天页面
│   └── knowledge.html      # 知识库页面
├── go.mod                  # Go 模块文件
├── go.sum                  # 依赖校验文件
└── README.md              # 项目说明
```

## 使用说明

### 基本聊天功能
1. 在 Web 界面左侧选择 AI 模型
2. 在文本框中输入您的问题
3. 点击"发送消息"按钮
4. AI 回复将以 Markdown 格式渲染显示

### 知识库管理
1. **查看最近问答**: 点击主页面"最近问答"按钮
2. **添加到知识库**: 在最近问答页面点击"添加到知识库"按钮
3. **管理知识库**: 访问 `/knowledge` 页面查看和管理所有知识条目
4. **标签分类**: 为知识条目添加标签，便于分类和查找

### 页面导航
- **主聊天页面**: `/` - 进行AI对话
- **知识库页面**: `/knowledge` - 查看和管理知识库
- **最近问答**: 通过主页面按钮访问

## 功能特色

### 📚 知识库系统
- **自动记录**: 每次对话自动记录到最近问答
- **一键保存**: 将有价值的回答保存到知识库
- **标签管理**: 为知识条目添加标签分类
- **独立页面**: 专门的知识库浏览页面
- **数据持久化**: 自动保存到本地JSON文件，重启不丢失

### 📋 问答记录
- **最近5次**: 自动保存最近5次问答记录
- **详细信息**: 包含问题、回答、模型、时间戳
- **快速访问**: 通过模态框快速查看

### 🎨 用户界面
- **响应式设计**: 适配不同屏幕尺寸
- **Markdown渲染**: 支持丰富的文本格式
- **现代化UI**: 美观的卡片式布局
- **交互友好**: 直观的操作流程

## 数据持久化

### 📁 数据存储
- **知识库数据**: 自动保存到 `data/knowledge.json`
- **问答记录**: 自动保存到 `data/recent_qas.json`
- **自动恢复**: 程序启动时自动加载历史数据
- **实时保存**: 每次操作后立即保存到文件

### 🔄 数据管理
- 程序会自动创建 `data/` 目录
- JSON格式存储，便于查看和备份
- 支持手动编辑JSON文件（需要重启服务生效）
- 数据文件采用UTF-8编码，支持中文内容

## 注意事项

- 请确保 API 密钥有效且有足够的额度
- 支持 OpenAI、Claude、DeepSeek 等多种 AI 服务
- 不同模型有不同的特点和优势，请根据需要选择
- 支持 Ctrl+Enter 快捷键发送消息
- 数据文件位于 `data/` 目录，请定期备份
- 修改JSON文件后需要重启服务才能生效
