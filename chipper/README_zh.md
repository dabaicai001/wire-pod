# chipper - wire-pod 核心组件

chipper是wire-pod项目的核心组件，作为本地服务器运行在用户的计算机上，负责处理语音识别、意图理解和命令执行，使Vector机器人无需付费即可使用语音命令功能。

## 项目概述

chipper最初由Digital Dream Labs开源，后由wire-pod项目维护并增强，支持Vector 1.0和2.0版本的所有功能，包括：

- 语音识别与转录
- 自然语言理解与意图处理
- 机器人命令执行
- 多语言支持
- 自定义意图
- Web管理界面

## 目录结构

```
chipper/
├── cmd/                  # 不同语音识别引擎的入口点
│   ├── coqui/            # Coqui STT引擎支持
│   ├── experimental/     # 实验性功能（Whisper、Houndify等）
│   ├── leopard/          # Picovoice Leopard STT引擎支持
│   └── vosk/             # Vosk STT引擎支持（默认）
├── epod/                 # Escape Pod相关文件
├── pkg/                  # 核心功能实现
│   ├── initwirepod/      # 服务器初始化
│   ├── logger/           # 日志功能
│   ├── mdnshandler/      # mDNS服务发现
│   ├── scripting/        # 脚本功能
│   ├── servers/          # gRPC和HTTP服务器
│   ├── vars/             # 全局变量
│   ├── vtt/              # Vector语音工具
│   └── wirepod/          # wire-pod特定功能
├── intent-data/          # 多语言意图数据文件
├── plugins/              # 插件系统
├── webroot/              # Web界面文件
├── start.sh              # 启动脚本
└── go.mod/go.sum         # Go模块依赖
```

## 核心功能模块

### 1. 语音识别 (STT)

chipper支持多种语音识别引擎，通过`pkg/wirepod/stt/`目录实现：

- **Vosk**：离线语音识别引擎，默认选项
- **Whisper**：基于OpenAI的Whisper模型的语音识别
- **Whisper.cpp**：Whisper的C++实现，优化了性能
- **Leopard**：Picovoice的高质量商业STT引擎
- **Coqui**：开源语音识别引擎
- **Houndify**：SoundHound的语音识别服务
- **Rhino**：Picovoice的语音到意图引擎（实验性）

### 2. 意图处理系统

意图处理负责将识别的文本转换为机器人可执行的命令，主要在`pkg/wirepod/ttr/`和`pkg/wirepod/preqs/`中实现：

- **语音请求处理** (`speechrequest`)：处理来自机器人的语音数据流，转换为适合STT引擎的格式
- **请求处理** (`preqs`)：处理不同类型的请求（意图识别、知识图谱查询等）
- **文本到响应** (`ttr`)：将识别的文本转换为机器人响应，包括意图解析和参数提取

### 3. 多语言支持

通过`intent-data/`目录下的语言文件支持多种语言：

- 英语 (en-US)
- 中文 (zh-CN)
- 德语 (de-DE)
- 西班牙语 (es-ES)
- 法语 (fr-FR)
- 意大利语 (it-IT)
- 荷兰语 (nl-NL)
- 波兰语 (pl-PL)
- 葡萄牙语 (pt-BR)
- 俄语 (ru-RU)
- 土耳其语 (tr-TR)
- 乌克兰语 (uk-UA)
- 越南语 (vi-VN)

### 4. Web管理界面

chipper提供了Web管理界面，位于`webroot/`目录，功能包括：

- 配置管理
- 自定义意图设置
- 机器人管理
- 系统状态监控
- 语言模型选择

### 5. gRPC和HTTP服务器

chipper实现了与Vector机器人通信的接口：

- gRPC服务器用于与机器人的高效通信
- HTTP服务器用于Web界面和管理功能
- 支持TLS加密通信

### 6. 插件系统

chipper支持通过插件扩展功能，位于`plugins/`目录：

- 自定义命令处理
- 新功能添加
- 第三方服务集成

## 配置与使用

### 基本配置

chipper的配置主要通过以下方式：

1. **环境变量**：通过`source.sh`文件配置
2. **Web界面**：通过默认地址 http://localhost:8080 进行配置
3. **自定义意图**：通过Web界面或`customIntents.json`文件定义

### 启动chipper

```bash
# 在chipper目录下运行
./start.sh
```

### 选择语音识别引擎

在安装过程中，可以选择不同的语音识别引擎，也可以通过修改`source.sh`文件中的`STT_SERVICE`变量来更改：

```bash
# 支持的值: vosk, whisper.cpp, leopard, coqui, houndify
export STT_SERVICE=vosk
```

## 开发说明

### 添加新的STT引擎

要添加新的语音识别引擎支持，需要在以下位置实现：

1. 在`pkg/wirepod/stt/`目录下创建新的包
2. 实现必要的接口函数
3. 在`cmd/`目录下创建相应的入口点
4. 更新启动脚本以支持新引擎

### 自定义意图

自定义意图可以通过Web界面添加，也可以直接编辑`customIntents.json`文件：

```json
{
  "intents": [
    {
      "name": "自定义命令名称",
      "description": "命令描述",
      "phrases": ["触发短语1", "触发短语2"],
      "responses": ["机器人回复1", "机器人回复2"]
    }
  ]
}
```

### 修改响应行为

可以通过修改`pkg/wirepod/ttr/`目录下的代码来自定义机器人的响应行为。例如，`matchIntentSend.go`文件包含意图匹配和响应发送的核心逻辑。

## 依赖关系

chipper的主要依赖包括：

- **Go语言**：主要开发语言
- **gRPC**：与机器人通信的协议
- **TLS**：安全通信
- **语音识别库**：根据选择的STT引擎不同而不同
- **Web框架**：提供Web管理界面
- **JSON处理**：配置和数据处理

## 故障排除

### 连接问题

- 确保Vector机器人和chipper服务器在同一网络
- 检查证书配置是否正确
- 验证服务器是否在正确的端口上运行

### 语音识别问题

- 尝试切换不同的语音识别引擎
- 确保选择了正确的语言模型
- 检查环境噪音是否过大
- 调整麦克风灵敏度设置

### 自定义意图不工作

- 确保触发短语拼写正确
- 检查意图配置是否保存成功
- 尝试使用更独特的触发短语

## 许可证

chipper遵循与wire-pod相同的许可证，基于Digital Dream Labs开源的代码构建。

## 贡献

欢迎对chipper进行贡献，包括：

- 报告bug
- 提交新功能建议
- 编写文档
- 提交代码改进

请通过项目的GitHub页面进行贡献。