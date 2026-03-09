# ccs - Claude Config Switcher

Claude 配置切换器 - 快速切换不同的 Claude 环境配置

## 功能特性

- 🚀 **快速切换** - 一键切换不同的 Claude API 配置
- 📋 **交互式选择** - 支持交互式菜单选择模型
- 🎨 **绿色高亮** - 选中项绿色显示，清晰直观
- ✅ **配置验证** - 支持验证配置文件格式
- 💡 **详细输出** - 支持查看详细配置信息

## 项目结构

```
ccs-go/
├── cmd/ccs/main.go           # CLI 入口程序
├── internal/
│   ├── config/config.go      # 配置管理器
│   ├── utils/utils.go        # 工具函数
│   └── validator/validator.go # 配置验证器
├── go.mod
├── go.sum
└── README.md
```

## 开发构建

### 环境要求

- Go 1.21 或更高版本

### 安装依赖

```bash
go mod tidy
```

### 构建

#### Windows

```bash
# 普通构建
go build -o ccs.exe ./cmd/ccs

# 优化构建（去除调试信息，减小文件体积）
go build -ldflags="-s -w" -trimpath -o ccs.exe ./cmd/ccs
```

#### macOS / Linux

```bash
go build -ldflags="-s -w" -trimpath -o ccs ./cmd/ccs
```

## 使用说明

### 配置文件位置

配置文件存储在：`~/.claude-config-switch/claudeEnvConfig.json`

### 配置文件格式

```json
{
  "work": {
    "ANTHROPIC_API_KEY": "your-work-api-key",
    "ANTHROPIC_BASE_URL": "https://api.anthropic.com",
    "ANTHROPIC_MODEL": "claude-sonnet-4-20250514"
  },
  "personal": {
    "ANTHROPIC_API_KEY": "your-personal-api-key",
    "ANTHROPIC_BASE_URL": "https://custom-api.example.com",
    "ANTHROPIC_MODEL": "claude-opus-4-20250514"
  }
}
```

### 命令

| 命令 | 说明 |
|------|------|
| `ccs` | 交互式选择模型 |
| `ccs <model>` | 直接切换到指定模型 |
| `ccs -l, --list` | 列出所有可用模型 |
| `ccs -i, --info` | 显示当前 env 配置 |
| `ccs -V, --validate` | 验证配置文件格式 |
| `ccs --verbose` | 详细输出（配合 -V 使用） |
| `ccs -h, --help` | 显示帮助信息 |
| `ccs -v, --version` | 显示版本号 |

### 使用示例

#### 1. 交互式选择模型

```bash
ccs
```

运行后会显示所有可用模型，使用上下箭头选择，回车确认。

#### 2. 直接切换到指定模型

```bash
ccs work
```

直接切换到名为 "work" 的配置。

#### 3. 列出所有可用模型

```bash
ccs -l
```

#### 4. 查看当前配置

```bash
ccs -i
```

显示当前 settings.json 中的 env 配置详情。

#### 5. 验证配置文件

```bash
ccs -V              # 简要验证
ccs -V --verbose    # 详细验证，显示完整配置内容
```

## 工作原理

ccs 通过修改 `~/.claude/settings.json` 文件中的 `env` 字段来切换不同的 Claude API 配置。

1. 所有模型配置存储在 `~/.claude-config-switch/claudeEnvConfig.json`
2. 切换时，读取指定模型的 env 配置
3. 将 env 配置写入 `~/.claude/settings.json`

## 许可证

MIT License
