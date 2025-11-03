# Wello Go Common

团队共用组件工具箱，提供 Kafka、AWS、Redis、Logging 等常用组件的封装。

## 目录结构

```text
.
├── kafka/       # Kafka 相关组件
│   ├── config.go      # 配置接口
│   ├── consumer.go    # 消费者池管理
│   ├── producer.go    # 生产者封装
│   └── kafka.go       # 包说明
├── aws/         # AWS 相关组件
│   └── aws.go         # AWS 服务封装（S3、Secrets Manager）
├── redis/       # Redis 相关组件（待实现）
├── logging/     # Logging 相关组件
│   └── logging.go     # 日志接口定义
└── Makefile     # 常用命令
```

## 安装

### 私有仓库配置

由于这是私有仓库，需要先配置 Git 和 Go 环境变量：

#### 方式 1：使用 SSH（推荐）

```bash
# 配置 Git 使用 SSH
git config --global url."git@github.com:".insteadOf "https://github.com/"

# 设置 GOPRIVATE，跳过私有模块的校验和数据库
go env -w GOPRIVATE=github.com/liberty-group-tech/wello-go-common

# 安装
go get github.com/liberty-group-tech/wello-go-common
```

#### 方式 2：使用 HTTPS + Personal Access Token

```bash
# 设置 GOPRIVATE
go env -w GOPRIVATE=github.com/liberty-group-tech/wello-go-common

# 配置 Git 凭证（替换 YOUR_TOKEN 为你的 GitHub Personal Access Token）
git config --global url."https://YOUR_TOKEN@github.com/".insteadOf "https://github.com/"

# 或者使用 GIT_TERMINAL_PROMPT 环境变量
export GIT_TERMINAL_PROMPT=1
go get github.com/liberty-group-tech/wello-go-common
```

### 导入使用

**重要**：必须先完成上面的私有仓库配置，然后才能导入子包。

```bash
# 1. 首先安装主模块（必须）
go get github.com/liberty-group-tech/wello-go-common

# 2. 然后在代码中导入子包
```

```go
// 导入 logging 包
import "github.com/liberty-group-tech/wello-go-common/logging"

// 导入 kafka 包
import "github.com/liberty-group-tech/wello-go-common/kafka"

// 导入 aws 包
import "github.com/liberty-group-tech/wello-go-common/aws"
```

**常见问题：**

- 如果遇到 `cannot find package` 错误：确保已设置 `GOPRIVATE` 并配置了 Git 认证
- 如果遇到 `404 Not Found` 错误：确保 Git 可以访问私有仓库（SSH 或 Token）
- 导入子包时不需要单独 `go get` 子包路径，只需 `go get` 主模块路径

## Logging 组件

提供统一的日志接口，其他组件依赖此接口进行日志记录。

```go
import "github.com/liberty-group-tech/wello-go-common/logging"

// Logger 接口定义
type Logger interface {
    Errorf(format string, args ...interface{})
    Infof(format string, args ...interface{})
    Debugf(format string, args ...interface{})
}

// 使用空操作日志器（不输出日志）
logger := &logging.NoOpLogger{}

// 实现自定义日志器
type MyLogger struct{}

func (l *MyLogger) Errorf(format string, args ...interface{}) {
    log.Printf("[ERROR] "+format, args...)
}

func (l *MyLogger) Infof(format string, args ...interface{}) {
    log.Printf("[INFO] "+format, args...)
}

func (l *MyLogger) Debugf(format string, args ...interface{}) {
    log.Printf("[DEBUG] "+format, args...)
}
```

## Kafka 组件

### 基本配置

```go
import (
    "github.com/liberty-group-tech/wello-go-common/kafka"
    "github.com/liberty-group-tech/wello-go-common/logging"
)

// 方式1：使用 SimpleConfig
cfg := &kafka.SimpleConfig{
    Brokers:     []string{"localhost:9092"},
    ClientID:    "your-app",
    Environment: "prod",
}

// 方式2：实现 Config 接口（适配现有项目配置）
type MyConfig struct {
    KafkaBrokers []string
    AppName      string
    Env          string
}

func (c *MyConfig) GetBrokers() []string {
    return c.KafkaBrokers
}

func (c *MyConfig) GetClientID() string {
    return c.AppName
}

func (c *MyConfig) GetEnvironment() string {
    return c.Env
}
```

### 生产者

```go
// 创建日志器（可选，nil 时使用 NoOpLogger）
var logger logging.Logger = &MyLogger{}

// 创建生产者
producer, err := kafka.NewProducer(cfg, logger)
if err != nil {
    log.Fatal(err)
}
defer producer.Close()

// 确保主题存在
topics := []string{"test.topic"}
if err := producer.EnsureTopics(topics, 3, 1); err != nil {
    log.Fatal(err)
}

// 发送消息
message := []byte(`{"key": "value"}`)
if err := producer.SendMessage("test.topic", message, nil); err != nil {
    log.Fatal(err)
}
```

### 消费者

```go
// 创建消费者管理器
consumerMgr := kafka.NewConsumerManager(cfg, logger, "consumer-group-id")

// 获取消费者池
pool := consumerMgr.GetPool("test.topic")

// 借用消费者
consumer, returnFunc := pool.Borrow()
defer returnFunc()

// 使用消费者
ev := consumer.Poll(1000)
switch e := ev.(type) {
case *kafka.Message:
    // 处理消息
case kafka.Error:
    // 处理错误
}
```

## AWS 组件

提供 AWS 服务封装，支持 Secrets Manager 和 S3。

### 初始化

```go
import "github.com/liberty-group-tech/wello-go-common/aws"

cfg := &aws.AwsConfig{
    Region:          "us-east-1",
    AccessKeyID:     "your-access-key",
    SecretAccessKey: "your-secret-key",
    SessionToken:    "", // 可选，用于临时凭证
    Endpoint:        "", // 可选，用于本地测试
}

service := aws.NewAwsService(cfg)
```

### Secrets Manager

```go
// 获取密钥
secretArn := "arn:aws:secretsmanager:us-east-1:123456789012:secret:my-secret"
secrets, err := service.GetSecrets(&secretArn)
if err != nil {
    log.Fatal(err)
}

// secrets 是 map[string]string 类型
fmt.Println(secrets["username"])
```

### S3 操作

```go
// 上传文件
file, err := os.Open("local-file.txt")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

err = service.UploadPublicFile("my-bucket", "path/to/file.txt", file)
if err != nil {
    log.Fatal(err)
}

// 生成预签名 URL（默认 30 分钟过期）
url, err := service.GetPresignedS3URL("my-bucket", "path/to/file.txt", nil)
if err != nil {
    log.Fatal(err)
}

// 自定义过期时间（秒）
expires := 3600 // 1 小时
url, err = service.GetPresignedS3URL("my-bucket", "path/to/file.txt", &expires)
```

## 特性

- ✅ **消费者池管理**：自动复用 Kafka 消费者连接
- ✅ **配置抽象**：通过接口适配不同项目的配置结构
- ✅ **日志抽象**：统一的日志接口，适配不同日志库
- ✅ **主题自动创建**：Kafka 主题自动创建功能
- ✅ **AWS 服务封装**：Secrets Manager 和 S3 便捷操作
- ✅ **无项目特定依赖**：可被多个项目复用

## 开发

```bash
# 运行测试
make test

# 格式化代码
make fmt

# 代码检查
make lint

# 清理构建文件
make clean
```
