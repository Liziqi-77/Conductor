# Go ZMQ 代码逐行详解

这份文档详细解释了 4 个 Go 语言示例文件中的每一行代码。既然你是 Go 语言初学者，我会特别标注出 **Go 语言特有的语法特性**。

---

## 1. `zmq_simple_publisher.go` (基础发布者)

### 核心概念
- **包管理**: `package main` 和 `import`
- **错误处理**: `if err != nil`
- **资源清理**: `defer`
- **类型推断**: `:=`

| 代码行 | 语法 / 概念 | 解释与作用 |
|:---|:---|:---|
| `package main` | **包声明** | 声明这是一个可执行程序（而不是库）。Go 程序的入口文件必须属于 `main` 包。 |
| `import (...)` | **导入块** | 导入依赖包。`<br>`- `fmt`: 格式化输入输出（类似 C 的 printf）。<br>- `log`: 日志记录。<br>- `math/rand`: 随机数生成。<br>- `strings`: 字符串操作。<br>- `time`: 时间处理。<br>- `zmq "github.com/pebbe/zmq4"`: 导入第三方 ZMQ 库，并起个别名 `zmq` 方便调用。 |
| `func main() {` | **主函数** | 程序的入口点。Go 程序从这里开始执行。 |
| `fmt.Println(...)` | **函数调用** | 打印一行文本并换行。`strings.Repeat("=", 60)` 生成 60 个等号。 |
| `publisher, err := zmq.NewSocket(zmq.PUB)` | **短变量声明 (`:=`)**<br>**多返回值** | **关键语法**：<br>1. `:=` 是声明并赋值，Go 会自动推断变量类型。<br>2. Go 函数可以返回多个值。这里返回了 `socket` 对象和 `error` 对象。<br>3. `zmq.PUB` 是库中定义的常量，表示发布者模式。 |
| `if err != nil {` | **错误检查** | **Go 惯例**：Go 没有 try-catch 异常。标准做法是检查返回的 `err` 是否为 `nil` (空)。如果不为空，说明出错了。 |
| `log.Fatal(err)` | **日志并退出** | 打印错误信息并立即终止程序（退出码 1）。 |
| `defer publisher.Close()` | **延迟执行 (`defer`)** | **关键语法**：`defer` 后的语句会被推迟到当前函数 (`main`) 返回前执行。无论函数是正常结束还是出错退出，`Close()` 都会被调用。用于确保资源（Socket）被释放。 |
| `address := "tcp://*:5555"` | **变量定义** | 定义监听地址字符串。 |
| `err = publisher.Bind(address)` | **方法调用** | 调用 socket 的 `Bind` 方法监听端口。注意这里用 `=` 赋值给已存在的 `err` 变量，而不是 `:=`。 |
| `time.Sleep(2 * time.Second)` | **时间休眠** | 暂停程序执行。Go 的时间单位是强类型的，必须乘以 `time.Second` 常量。 |
| `for {` | **无限循环** | Go 只有 `for` 关键字，没有 `while`。不带条件的 `for` 就是死循环 (while true)。 |
| `msgCount++` | **自增** | 变量加 1。 |
| `rand.Intn(16) + 15` | **随机数** | 生成 `[0, 16)` 的随机整数，然后加 15，结果范围 15-30。 |
| `msg := fmt.Sprintf(...)` | **字符串格式化** | 类似 Python 的 f-string 或 C 的 sprintf。`%d` 是整数占位符。Go 必须显式格式化，不能直接拼接不同类型。 |
| `_, err := publisher.Send(msg, 0)` | **空标识符 (`_`)** | `Send` 返回两个值：(发送的字节数, 错误)。我们不关心字节数，所以用 `_` 忽略它（Go 要求所有声明的变量必须被使用，不用的必须用 `_`）。参数 `0` 表示默认标志（阻塞发送）。 |
| `log.Println(...)` | **日志打印** | 打印日志，自动带时间戳。 |
| `break` | **跳出循环** | 退出 `for` 循环。 |
| `time.Now().Format("15:04:05")` | **时间格式化** | **有趣特性**：Go 的时间格式化不用 `%Y-%m-%d`，而是用固定的参考时间 `2006-01-02 15:04:05`。这里 `15:04:05` 代表时分秒。 |

---

## 2. `zmq_simple_subscriber.go` (基础订阅者)

### 核心概念
- **Socket 连接**: `Connect`
- **订阅过滤**: `SetSubscribe` (至关重要)

| 代码行 | 语法 / 概念 | 解释与作用 |
|:---|:---|:---|
| `subscriber, err := zmq.NewSocket(zmq.SUB)` | **创建 Socket** | 创建一个 `SUB` (订阅者) 类型的 Socket。 |
| `err = subscriber.Connect(address)` | **连接** | 订阅者使用 `Connect` 连接到发布者的地址。注意：ZMQ 中 Bind/Connect 的顺序不重要，可以先连接再启动发布者。 |
| `subscriber.SetSubscribe("")` | **方法调用** | **核心逻辑**：ZMQ 的 SUB socket 默认过滤掉所有消息。必须调用此方法设置订阅前缀。空字符串 `""` 表示接收所有消息。 |
| `msg, err := subscriber.Recv(0)` | **接收消息** | 阻塞等待接收消息。`0` 是默认标志。如果收到消息，`msg` 将包含字符串内容。 |

---

## 3. `zmq_topic_publisher.go` (主题发布者)

### 核心概念
- **切片 (Slice)**: `[]string`
- **Switch 语句**: 多条件分支
- **类型零值**: `var` 声明

| 代码行 | 语法 / 概念 | 解释与作用 |
|:---|:---|:---|
| `topics := []string{"T", "H", "P"}` | **切片字面量** | **关键语法**：`[]string` 定义了一个字符串切片（类似 Python 的 List）。`{...}` 初始化了内容。 |
| `topic := topics[rand.Intn(len(topics))]` | **切片索引** | `len(topics)` 获取长度。`rand.Intn` 生成随机索引。`topics[...]` 取出元素。 |
| `var value int` | **变量声明** | 声明一个 `int` 类型的变量 `value`。Go 会自动初始化为**零值** (int 为 0)，不需要显式赋值。 |
| `var unit string` | **变量声明** | 声明字符串变量，默认初始化为 `""` (空字符串)。 |
| `switch topic {` | **Switch 语句** | 根据 `topic` 的值进行分支判断。 |
| `case "Temperature":` | **分支** | 如果 `topic` 等于 "Temperature"，执行下面的代码。**注意**：Go 的 switch 默认不需要 `break`，执行完 case 会自动跳出。 |
| `value = rand.Intn(16) + 15` | **赋值** | 给之前声明的变量赋值。 |
| `msg := fmt.Sprintf("%s %d%s", ...)` | **格式化** | `%s`: 字符串, `%d`: 整数。构造类似 "Temperature 25°C" 的消息。 |
| `fmt.Printf(...)` | **格式化打印** | `%-12s`: 字符串左对齐，占用 12 个字符宽度。用于对齐日志输出，更美观。 |

---

## 4. `zmq_topic_subscriber.go` (主题订阅者)

### 核心概念
- **命令行参数**: `os.Args`
- **条件判断**: `if/else`

| 代码行 | 语法 / 概念 | 解释与作用 |
|:---|:---|:---|
| `import "os"` | **导入包** | `os` 包提供了操作系统相关的功能，如命令行参数。 |
| `filter := ""` | **变量初始化** | 默认过滤器为空（接收所有）。 |
| `if len(os.Args) > 1 {` | **参数检查** | `os.Args` 是一个字符串切片，包含命令行参数。`os.Args[0]` 是程序本身的名字，`os.Args[1]` 是第一个参数。 |
| `filter = os.Args[1]` | **获取参数** | 如果用户提供了参数（如 `go run main.go Temperature`），则将过滤器设置为 "Temperature"。 |
| `subscriber.SetSubscribe(filter)` | **设置订阅** | 设置 ZMQ 的订阅前缀。ZMQ 会在内核层面过滤消息，只有以前缀开头的消息才会传给应用程序。 |
| `if filter == "" {` | **条件判断** | 根据是否设置了过滤器打印不同的提示信息。 |

---

## 💡 Go 语言初学者常见疑惑解答

1.  **为什么很多函数首字母是大写的？**
    *   **Go 的访问控制**：首字母**大写**的函数/变量是 **Public** (导出的)，可以被其他包访问（如 `fmt.Println`, `zmq.NewSocket`）。首字母**小写**的是 **Private** (私有的)，只能在当前包内使用。

2.  **`:=` 和 `=` 有什么区别？**
    *   `:=` 是**声明 + 赋值**，只能用于新变量（第一次出现）。
    *   `=` 是**赋值**，用于已经声明过的变量。

3.  **为什么没有 `try...catch`？**
    *   Go 的设计哲学认为错误处理是显式的逻辑，不是异常。每个可能出错的操作都返回一个 `error` 值，程序员必须显式处理它。这虽然写起来繁琐，但能让代码逻辑更清晰、健壮。

4.  **`defer` 到底什么时候执行？**
    *   `defer` 就像是给函数设置了一个"临终遗言"。不管函数是正常跑完、半路 `return`、还是发生了 `panic` 崩溃，只要执行到了 `defer` 这一行，它注册的操作（如 `Close()`）就一定会在函数退出前被执行。

希望这份逐行详解能帮你快速建立起 Go 语言和 ZMQ 的概念！

