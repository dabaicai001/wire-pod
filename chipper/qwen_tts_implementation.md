# Qwen-TTS 功能集成文档

## 概述

本文档介绍了如何在 wire-pod 项目中集成和使用 Qwen-TTS（通义千问语音合成）功能。Qwen-TTS 提供了高质量的语音合成服务，特别适合中文语音合成，同时支持英文和其他语言。

## 功能特性

- **高质量语音合成**：基于阿里云的通义千问TTS服务
- **多语言支持**：支持中文、英文等多种语言
- **情感音色**：提供标准音色和情感音色两种模式
- **优先级控制**：Qwen-TTS 优先级高于 OpenAI TTS
- **Web界面配置**：通过Web界面轻松配置语音选项

## 配置选项

### Qwen-TTS 语音选择

在知识图谱配置界面中，当选择 OpenAI 作为提供商时，可以看到 Qwen-TTS 配置选项：

- **Qwen-TTS voice for non-English languages**：选择非英语语言使用的语音
- **Use the Qwen-TTS voice for English as well**：是否对英语也使用 Qwen-TTS 语音（优先级高于 OpenAI TTS）

### 可用语音列表

Qwen-TTS 提供以下14种语音选项：

| 语音ID | 名称 | 类型 |
|--------|------|------|
| zhitian | 知甜 | 标准 |
| zhitian_emo | 知甜-情感 | 情感 |
| zhizhe | 知哲 | 标准 |
| zhizhe_emo | 知哲-情感 | 情感 |
| zhiyan | 知燕 | 标准 |
| zhiyan_emo | 知燕-情感 | 情感 |
| zhiqi | 知琪 | 标准 |
| zhiqi_emo | 知琪-情感 | 情感 |
| zhiling | 知灵 | 标准 |
| zhiling_emo | 知灵-情感 | 情感 |
| zhimei | 知美 | 标准 |
| zhimei_emo | 知美-情感 | 情感 |
| zhibei | 知贝 | 标准 |
| zhibei_emo | 知贝-情感 | 情感 |

## 使用说明

### 1. 启用 Qwen-TTS

1. 打开 wire-pod Web 界面
2. 进入设置页面
3. 选择 "Knowledge Graph API Provider" 为 "OpenAI"
4. 在 OpenAI 配置区域下方找到 Qwen-TTS 配置选项
5. 选择所需的语音和启用选项

### 2. 语音选择逻辑

wire-pod 的语音选择遵循以下优先级：

1. **Qwen-TTS 优先**：如果启用了 Qwen-TTS 英语语音，或当前语言非英语且知识图谱提供商为 OpenAI，则使用 Qwen-TTS
2. **OpenAI TTS 次之**：如果未启用 Qwen-TTS 但启用了 OpenAI TTS 英语语音，或当前语言非英语，则使用 OpenAI TTS
3. **Vector 内置 TTS**：默认使用 Vector 机器人的内置 TTS

### 3. API 配置要求

要使用 Qwen-TTS 功能，需要：

- 有效的阿里云 DashScope API Key
- 在 OpenAI 配置中设置正确的 API Key
- 网络连接能够访问阿里云服务

## 技术实现

### 核心文件

- `pkg/vars/config.go`：添加 Qwen-TTS 配置字段
- `pkg/wirepod/ttr/kgsim_cmds.go`：实现 Qwen-TTS 语音合成功能
- `webroot/setup.html`：Web 界面配置选项
- `webroot/js/main.js`：JavaScript 配置处理逻辑

### API 调用流程

1. **请求构建**：根据配置构建 Qwen-TTS API 请求
2. **语音映射**：将配置的语音ID映射为 API 参数
3. **HTTP 调用**：向阿里云 DashScope API 发送请求
4. **音频处理**：处理返回的音频数据，进行降采样等操作
5. **流式播放**：通过 Vector 机器人的外部音频流接口播放音频

## 故障排除

### 常见问题

1. **Qwen-TTS 不工作**
   - 检查 API Key 是否正确配置
   - 确认网络连接正常
   - 查看日志文件中的错误信息

2. **语音选择不正确**
   - 检查知识图谱提供商设置
   - 确认 Qwen-TTS 配置选项已正确启用
   - 查看语音选择逻辑是否符合预期

3. **音频播放问题**
   - 检查音频流处理是否正确
   - 确认降采样函数正常工作
   - 查看网络延迟是否影响音频播放

### 日志调试

Qwen-TTS 相关的日志信息可以在 wire-pod 的日志中找到：

- Qwen-TTS API 请求和响应信息
- 音频处理过程中的错误信息
- 语音选择决策的日志记录

## 性能考虑

- **网络延迟**：Qwen-TTS 需要网络连接，可能增加响应时间
- **音频质量**：支持 24kHz 高质量音频，通过降采样适配 Vector 的 16kHz 要求
- **并发处理**：支持多机器人同时使用 Qwen-TTS 功能

## 版本历史

- **v1.0**：初始版本，集成 Qwen-TTS 基本功能
- 支持14种语音选项
- 完整的 Web 界面配置
- 优先级控制的语音选择逻辑

## 相关资源

- [阿里云 DashScope 文档](https://help.aliyun.com/zh/dashscope/)
- [Qwen-TTS 官方介绍](https://help.aliyun.com/zh/dashscope/developer-reference/tongyi-qianwen-tts)
- [wire-pod 项目文档](https://github.com/kercre123/wire-pod)

---

**注意**：使用 Qwen-TTS 功能需要遵守阿里云的相关服务条款和API使用限制。