# Vector 机器人自定义意图功能实现流程

本文档详细说明了 wire-pod 项目中自定义意图功能的代码实现流程，包括通过 Web 界面添加和直接编辑 `customIntents.json` 文件两种方式。

## 1. 自定义意图数据结构

自定义意图在项目中通过 `CustomIntent` 结构体定义，包含了意图的所有必要信息：

```go
// vars.go 中定义的 CustomIntent 结构体
type CustomIntent struct {
    Name        string   `json:"name"`        // 意图名称
    Description string   `json:"description"` // 意图描述
    Utterances  []string `json:"utterances"`  // 触发该意图的语音短语
    Intent      string   `json:"intent"`      // 机器人执行的意图类型
    Params      struct {
        ParamName  string `json:"paramname"`  // 参数名称
        ParamValue string `json:"paramvalue"` // 参数值
    } `json:"params"`
    Exec           string   `json:"exec"`           // 执行的外部脚本路径
    ExecArgs       []string `json:"execargs"`       // 外部脚本参数
    IsSystemIntent bool     `json:"issystem"`       // 是否为系统意图
    LuaScript      string   `json:"luascript"`      // Lua 脚本内容
}
```

## 2. 自定义意图存储机制

### 2.1 文件路径定义

自定义意图存储在 `customIntents.json` 文件中，文件路径在 `vars.go` 中定义：

```go
// vars.go 中定义的文件路径
var (
    CustomIntentsPath string = "./customIntents.json"
    // 其他文件路径...
)
```

### 2.2 意图加载流程

系统启动时，通过 `LoadCustomIntents()` 函数加载自定义意图：

```go
// vars.go 中的 LoadCustomIntents 函数
func LoadCustomIntents() {
    jsonBytes, err := os.ReadFile(CustomIntentsPath)
    if err == nil {
        json.Unmarshal(jsonBytes, &CustomIntents)
        CustomIntentsExist = true
        logger.Println("Loaded custom intents:")
        for _, intent := range CustomIntents {
            logger.Println(intent.Name)
        }
    }
}
```

### 2.3 意图保存机制

当添加、编辑或删除自定义意图时，通过 `saveCustomIntents()` 函数将变更持久化到文件：

```go
// webserver.go 中的 saveCustomIntents 函数
func saveCustomIntents() {
    customIntentJSONFile, _ := json.Marshal(vars.CustomIntents)
    os.WriteFile(vars.CustomIntentsPath, customIntentJSONFile, 0644)
}
```

## 3. Web 界面交互流程

### 3.1 界面元素

Web 界面通过 `webroot/index.html` 中的表单实现自定义意图的添加：

```html
<div id="section-intents" style="display: none">
  <h2 id="foldable-add" onclick="toggleSection('content-add', 'content-edit')">
    <span>+</span>
    Add a custom intent
  </h2>
  <div class="content" id="content-add">
    <!-- 意图添加表单 -->
    <form id="intentAddForm">
      <label for="nameAdd">Custom intent name:</label>
      <input type="text" name="nameAdd" id="nameAdd" /><br />
      <label for="descriptionAdd">Custom intent description:</label>
      <input type="text" name="descriptionAdd" id="descriptionAdd" /><br />
      <label for="utterancesAdd">Utterances that will trigger the intent (separated by ,):</label>
      <input type="text" name="utterancesAdd" id="utterancesAdd" /><br />
      <div id="intentAddSelect"></div>
      <!-- 其他配置项 -->
    </form>
  </div>
</div>
```

### 3.2 JavaScript 交互

`webroot/js/main.js` 包含处理意图添加、编辑和删除的客户端脚本：

```javascript
// 初始化意图选择下拉框
function createIntentSelect(element) {
  const select = document.createElement("select");
  select.name = `${element}intents`;
  select.id = `${element}intents`;
  intentsJson.forEach((intent) => {
    const option = document.createElement("option");
    option.value = intent;
    option.text = intent;
    select.appendChild(option);
  });
  // 添加到表单
}

// 获取自定义意图列表
function updateIntentSelection(element) {
  fetch("/api/get_custom_intents_json")
    .then((response) => response.json())
    .then((listResponse) => {
      // 创建选择器
    });
}

// 编辑意图表单创建
function editFormCreate() {
  const intentNumber = getE("editSelectintents").selectedIndex;
  fetch("/api/get_custom_intents_json")
    .then((response) => response.json())
    .then((intents) => {
      // 创建编辑表单
    });
}
```

## 4. API 接口处理流程

### 4.1 API 路由设置

`webserver.go` 中的 `apiHandler` 函数设置了自定义意图相关的 API 路由：

```go
func apiHandler(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Access-Control-Allow-Origin", "*")
  w.Header().Set("Access-Control-Allow-Headers", "*")

  switch strings.TrimPrefix(r.URL.Path, "/api/") {
  case "add_custom_intent":
    handleAddCustomIntent(w, r)
  case "edit_custom_intent":
    handleEditCustomIntent(w, r)
  case "get_custom_intents_json":
    handleGetCustomIntentsJSON(w)
  case "remove_custom_intent":
    handleRemoveCustomIntent(w, r)
  // 其他 API 路由...
  }
}
```

### 4.2 添加自定义意图

`handleAddCustomIntent` 函数处理添加新自定义意图的请求：

```go
func handleAddCustomIntent(w http.ResponseWriter, r *http.Request) {
  var intent vars.CustomIntent
  // 解析请求体
  if err := json.NewDecoder(r.Body).Decode(&intent); err != nil {
    http.Error(w, "invalid request body", http.StatusBadRequest)
    return
  }
  // 验证必填字段
  if anyEmpty(intent.Name, intent.Description, intent.Intent) || len(intent.Utterances) == 0 {
    http.Error(w, "missing required field (name, description, utterances, and intent are required)", http.StatusBadRequest)
    return
  }
  // 验证 Lua 脚本（如果提供）
  intent.LuaScript = strings.TrimSpace(intent.LuaScript)
  if intent.LuaScript != "" {
    if err := scripting.ValidateLuaScript(intent.LuaScript); err != nil {
      http.Error(w, "lua validation error: "+err.Error(), http.StatusBadRequest)
      return
    }
  }
  // 添加到内存并保存到文件
  vars.CustomIntentsExist = true
  vars.CustomIntents = append(vars.CustomIntents, intent)
  saveCustomIntents()
  fmt.Fprint(w, "Intent added successfully.")
}
```

### 4.3 编辑自定义意图

`handleEditCustomIntent` 函数处理编辑现有自定义意图的请求：

```go
func handleEditCustomIntent(w http.ResponseWriter, r *http.Request) {
  var request struct {
    Number int `json:"number"`
    vars.CustomIntent
  }
  // 解析请求
  if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
    http.Error(w, "invalid request body", http.StatusBadRequest)
    return
  }
  // 验证意图索引
  if request.Number < 1 || request.Number > len(vars.CustomIntents) {
    http.Error(w, "invalid intent number", http.StatusBadRequest)
    return
  }
  // 更新意图属性
  intent := &vars.CustomIntents[request.Number-1]
  if request.Name != "" {
    intent.Name = request.Name
  }
  // 更新其他属性...
  // 保存更改
  saveCustomIntents()
  fmt.Fprint(w, "Intent edited successfully.")
}
```

### 4.4 删除自定义意图

`handleRemoveCustomIntent` 函数处理删除自定义意图的请求：

```go
func handleRemoveCustomIntent(w http.ResponseWriter, r *http.Request) {
  var request struct {
    Number int `json:"number"`
  }
  // 解析请求
  if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
    http.Error(w, "invalid request body", http.StatusBadRequest)
    return
  }
  // 验证意图索引
  if request.Number < 1 || request.Number > len(vars.CustomIntents) {
    http.Error(w, "invalid intent number", http.StatusBadRequest)
    return
  }
  // 从切片中删除意图
  vars.CustomIntents = append(vars.CustomIntents[:request.Number-1], vars.CustomIntents[request.Number:])
  saveCustomIntents()
  fmt.Fprint(w, "Intent removed successfully.")
}
```

### 4.5 获取自定义意图

`handleGetCustomIntentsJSON` 函数提供当前配置的自定义意图：

```go
func handleGetCustomIntentsJSON(w http.ResponseWriter) {
  if !vars.CustomIntentsExist {
    http.Error(w, "you must create an intent first", http.StatusBadRequest)
    return
  }
  // 读取并返回意图文件内容
  customIntentJSONFile, err := os.ReadFile(vars.CustomIntentsPath)
  if err != nil {
    http.Error(w, "could not read custom intents file", http.StatusInternalServerError)
    logger.Println(err)
    return
  }
  w.Header().Set("Content-Type", "application/json")
  w.Write(customIntentJSONFile)
}
```

## 5. 自定义意图匹配与执行流程

### 5.1 意图匹配逻辑

当机器人接收到语音指令后，`customIntentHandler` 函数负责匹配自定义意图：

```go
func customIntentHandler(req interface{}, voiceText string, botSerial string) bool {
  var successMatched bool = false
  if vars.CustomIntentsExist {
    for _, c := range vars.CustomIntents {
      for _, v := range c.Utterances {
        // 检查是否匹配语音文本
        var seekText = strings.ToLower(strings.TrimSpace(v))
        // 系统意图支持通配符匹配
        if (c.IsSystemIntent && strings.HasPrefix(seekText, "*")) || strings.Contains(voiceText, seekText) {
          logger.Println("Bot " + botSerial + " Custom Intent Matched: " + c.Name + " - " + c.Description + " - " + c.Intent)
          
          // 处理意图参数
          var intentParams map[string]string
          var isParam bool = false
          if c.Params.ParamValue != "" {
            logger.Println("Bot " + botSerial + " Custom Intent Parameter: " + c.Params.ParamName + " - " + c.Params.ParamValue)
            intentParams = map[string]string{c.Params.ParamName: c.Params.ParamValue}
            isParam = true
          }
          
          // 执行后续操作...
          successMatched = true
          break
        }
      }
    }
  }
  return successMatched
}
```

### 5.2 Lua 脚本执行

匹配到意图后，如果配置了 Lua 脚本，则异步执行：

```go
// 在 customIntentHandler 中
if c.LuaScript != "" {
  go func() {
    err := scripting.RunLuaScript(botSerial, c.LuaScript)
    if err != nil {
      logger.Println("Error running Lua script: " + err.Error())
    }
  }()
}
```

### 5.3 外部命令执行

如果配置了外部执行命令，则执行指定的命令：

```go
// 处理命令参数
var args []string
for _, arg := range c.ExecArgs {
  // 特殊参数替换
  if arg == "!botSerial" {
    arg = botSerial
  } else if arg == "!speechText" {
    arg = "\"" + voiceText + "\""
  } else if arg == "!intentName" {
    arg = c.Name
  } else if arg == "!locale" {
    arg = vars.APIConfig.STT.Language
  }
  args = append(args, arg)
}

// 执行命令
var customIntentExec *exec.Cmd
if len(args) == 0 {
  customIntentExec = exec.Command(c.Exec)
} else {
  customIntentExec = exec.Command(c.Exec, args...)
}

// 处理命令输出
var out bytes.Buffer
var stderr bytes.Buffer
customIntentExec.Stdout = &out
customIntentExec.Stderr = &stderr
err := customIntentExec.Run()
```

### 5.4 意图发送

最后，将匹配的意图发送给机器人：

```go
if c.IsSystemIntent {
  // 系统意图特殊处理
  var resp systemIntentResponseStruct
  err := json.Unmarshal(out.Bytes(), &resp)
  if err == nil && resp.Status == "ok" {
    logger.Println("Bot " + botSerial + " System intent parsed and executed successfully")
    IntentPass(req, resp.ReturnIntent, voiceText, intentParams, isParam)
  }
} else {
  // 普通意图直接发送
  IntentPass(req, c.Intent, voiceText, intentParams, isParam)
}
```

## 6. 直接编辑 customIntents.json 文件的流程

除了通过 Web 界面添加自定义意图外，用户还可以直接编辑 `customIntents.json` 文件：

1. **编辑文件**：用户手动编辑 `customIntents.json` 文件，按照 CustomIntent 结构体的 JSON 格式添加或修改意图
2. **重新加载**：系统重启时，通过 `LoadCustomIntents()` 函数重新加载文件中的意图
3. **生效使用**：加载后的意图会立即对新的语音请求生效

## 7. 自定义意图使用示例

### 7.1 示例：创建一个简单的问候意图

**JSON 格式示例**：
```json
[
  {
    "name": "custom_greeting",
    "description": "Custom greeting intent",
    "utterances": ["你好向量", "向量你好", "早上好向量"],
    "intent": "intent_greeting_hello",
    "params": {
      "paramname": "",
      "paramvalue": ""
    },
    "exec": "",
    "execargs": [],
    "issystem": false,
    "luascript": ""
  }
]
```

### 7.2 示例：创建带外部脚本执行的意图

**JSON 格式示例**：
```json
[
  {
    "name": "custom_command",
    "description": "Execute custom command",
    "utterances": ["运行命令", "执行脚本"],
    "intent": "intent_greeting_hello",
    "params": {
      "paramname": "",
      "paramvalue": ""
    },
    "exec": "./scripts/custom.sh",
    "execargs": ["!botSerial", "!speechText"],
    "issystem": false,
    "luascript": ""
  }
]
```

## 8. 自定义意图实现流程图

```
┌────────────────────────┐     ┌─────────────────────────┐     ┌─────────────────────────┐
│                        │     │                         │     │                         │
│   用户交互层           │     │      处理层             │     │      执行层             │
│                        │     │                         │     │                         │
└─────────┬──────────────┘     └─────────┬───────────────┘     └─────────┬───────────────┘
          │                              │                              │
          ▼                              ▼                              ▼
┌────────────────────────┐     ┌─────────────────────────┐     ┌─────────────────────────┐
│                        │     │                         │     │                         │
│ 1. Web界面表单         │────▶│ 2. API请求处理          │────▶│ 3. 意图保存/修改        │
│ 2. 直接编辑JSON文件    │     │ 3. 数据验证              │     │ 4. Lua脚本执行          │
│                        │     │ 4. 意图匹配              │     │ 5. 外部命令执行         │
└────────────────────────┘     └─────────┬───────────────┘     └─────────┬───────────────┘
                                         │                              │
                                         ▼                              ▼
                                  ┌─────────────────────────┐     ┌─────────────────────────┐
                                  │                         │     │                         │
                                  │ 意图配置持久化           │     │ 意图发送给机器人        │
                                  │ customIntents.json      │     │                         │
                                  │                         │     │                         │
                                  └─────────────────────────┘     └─────────────────────────┘
```

## 9. 总结

wire-pod 项目中的自定义意图功能通过以下关键组件实现：

1. **数据结构**：`CustomIntent` 结构体定义意图的所有属性
2. **存储机制**：使用 `customIntents.json` 文件持久化存储自定义意图
3. **Web 界面**：提供直观的用户界面用于添加、编辑和删除意图
4. **API 处理**：通过 RESTful API 处理意图的增删改查操作
5. **意图匹配**：实现语音文本与自定义意图的匹配逻辑
6. **扩展功能**：支持 Lua 脚本和外部命令执行，提供灵活的扩展能力

用户可以通过两种方式配置自定义意图：通过 Web 界面进行交互式配置，或直接编辑 `customIntents.json` 文件进行手动配置。两种方式最终都会更新同一个 JSON 文件，并在系统中生效。