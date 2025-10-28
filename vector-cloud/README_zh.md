# Vector-cloud

让Vector机器人与云服务通信的程序！

## 项目介绍

Vector-cloud是一个为Anki Vector机器人开发的云通信框架，允许Vector机器人与云服务进行交互，实现语音识别、知识图谱查询等功能。该项目是wire-pod的重要组成部分，提供了机器人与云服务之间的桥梁。

## 项目结构

```
vector-cloud/
├── cloud/             # 云通信核心组件
│   ├── main.go        # 云服务主程序入口
│   ├── config_dev.go  # 开发环境配置
│   └── config_shipping.go # 生产环境配置
├── gateway/           # 网关服务组件
│   ├── main.go        # 网关服务主程序入口
│   ├── message_handler.go # 消息处理器
│   └── switchboard_proxy.go # 交换机代理
├── internal/          # 内部包和库
│   ├── clad/          # 通信协议
│   ├── cloudproc/     # 云处理相关
│   ├── config/        # 配置管理
│   ├── ipc/           # 进程间通信
│   ├── voice/         # 语音处理
│   └── robot/         # 机器人相关功能
├── docker-builder/    # Docker构建环境
└── go.mod             # Go模块定义
```

## 核心组件

### 1. vic-cloud

vic-cloud是运行在Vector机器人上的核心组件，负责处理以下功能：
- 与云服务的安全通信（TLS加密）
- 证书管理和验证
- 语音处理和意图识别
- 知识图谱查询处理
- 机器人状态管理

### 2. vic-gateway

vic-gateway提供了一个API网关，处理：
- GRPC和HTTP请求转发
- 认证和授权
- 请求限流和安全控制
- 与机器人内部服务的IPC通信

## 构建方法

为了便于在计算机上交叉编译可以在Vector上运行的二进制文件，您首先需要armbuilder docker镜像。可以通过运行以下命令生成：

```bash
# 生成Docker构建环境
make docker-builder

# 构建vic-cloud
make vic-cloud

# 构建vic-gateway
make vic-gateway
```

## 自定义示例

以下是一个自定义Vector响应的示例，让Vector拒绝提供关于51区的信息，并明确表示所有其他信息请求已获批准：

首先，对`internal/voice/stream/context.go`进行以下修改：

```diff
diff --git a/internal/voice/stream/context.go b/internal/voice/stream/context.go
index 1d5df2c..564b22f 100644
--- a/internal/voice/stream/context.go
+++ b/internal/voice/stream/context.go
@@ -1,7 +1,9 @@
 package stream
 
 import (
-       "bytes"
+       "regexp"
+       
+       "bytes"
        "context"
        "encoding/json"
        "fmt"
@@ -155,6 +157,14 @@ func sendIntentResponse(resp *chipper.IntentResult, receiver Receiver) {
 
 func sendKGResponse(resp *chipper.KnowledgeGraphResponse, receiver Receiver) {
        var buf bytes.Buffer
+
+       found, _ := regexp.MatchString("area fifty one", resp.QueryText)
+       if found {
+         resp.SpokenText = "Information regarding Area Fifty One is classified. The Illuminati High Council has been notified of this request."
+       } else {
+         resp.SpokenText = "Information Request Approved. " + resp.SpokenText
+       }
+
        params := map[string]string{
                "answer":      resp.SpokenText,
                "answer_type": resp.CommandType,
```

接下来编译、复制到Vector并重启：

```bash
make vic-cloud
ssh root@<VECTOR_IP> mount -o remount,rw /
scp build/vic-cloud root@<VECTOR_IP>:/anki/bin
ssh root@<VECTOR_IP> /sbin/reboot
```

重启后，通过说"嘿Vector... 问题... 51区是什么？"和"嘿Vector... 问题... DogeCoin是什么？"进行测试。

## 证书管理

Vector-cloud支持自定义证书，用于安全通信。证书可以通过以下路径加载：
- `/anki/etc/wirepod-cert.crt`
- `/data/data/wirepod-cert.crt`

证书加载过程会在启动时自动完成，并支持自定义证书链的附加。

## 调试

项目提供了详细的日志记录功能，可以通过修改代码中的日志级别来启用更详细的调试信息。主要的日志开关包括：
- `logVerbose`: 启用网关请求日志
- `logMessageContent`: 启用详细消息内容日志

## 依赖关系

主要依赖包括：
- Go语言标准库
- gRPC和gRPC-Gateway
- 自定义的通信协议(CLAD)
- 证书管理库

## 注意事项

- 本项目需要在特定的环境下运行，主要是为Vector机器人设计的
- 修改代码后需要重新编译并部署到机器人上
- 证书管理和安全配置对于与云服务的正常通信至关重要

## 许可证

请参见项目根目录中的LICENSE文件。