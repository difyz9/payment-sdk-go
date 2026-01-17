# Payment SDK - 统一支付平台 Golang SDK

统一支付平台的官方 Golang SDK，支持支付宝、微信支付、PayPal 等多种支付方式。

## 特性

- ✅ 简洁易用的 API 设计
- ✅ 完整的类型定义和文档
- ✅ 自动签名认证（HMAC-SHA256）
- ✅ 支持多种支付方式（支付宝/微信/PayPal）
- ✅ 订单状态轮询功能
- ✅ 完善的错误处理
- ✅ 线程安全
- ✅ 零依赖（仅使用标准库）

## 安装

### 方式一：直接复制（推荐用于快速集成）

将 `payment-sdk` 目录复制到你的项目中：

```bash
cp -r payment-sdk /path/to/your/project/pkg/
```

### 方式二：Go Modules（推荐用于生产环境）

如果将此 SDK 发布到 Git 仓库，可以通过 go get 安装：

```bash
go get github.com/difyz9/payment-sdk
```

## 快速开始

### 1. 初始化客户端

```go
package main

import (
    "fmt"
    "github.com/difyz9/payment-sdk" // 或使用本地路径
)

func main() {
    // 创建客户端配置
    config := &paymentsdk.Config{
        BaseURL:   "https://api.example.com",
        AppID:     "your-app-id",
        AppSecret: "your-app-secret",
    }
    
    // 创建客户端实例
    client := paymentsdk.NewClient(config)
}
```

### 2. 创建支付订单

```go
// 创建支付宝订单
req := &paymentsdk.PaymentRequest{
    Subject:   "VIP会员-月卡",
    Amount:    0.01,
    PayWay:    paymentsdk.PayWayAlipay,
    OrderType: "vip",
    UserID:    "user_12345",
    Extra:     `{"period":"30days"}`,
}

paymentData, err := client.CreatePayment(req)
if err != nil {
    fmt.Printf("创建订单失败: %v\n", err)
    return
}

fmt.Printf("支付链接: %s\n", paymentData.PayUrl)
fmt.Printf("订单号: %s\n", paymentData.OrderNo)
```

### 3. 查询订单状态

```go
// 查询单次
orderStatus, err := client.QueryOrder(orderNo)
if err != nil {
    fmt.Printf("查询失败: %v\n", err)
    return
}

if orderStatus.IsPaymentSuccess() {
    fmt.Println("支付成功！")
}
```

### 4. 轮询查询订单状态

```go
// 自动轮询直到支付成功或超时
orderStatus, err := client.PollOrderStatus(orderNo, &paymentsdk.PollOptions{
    Interval:   5 * time.Second,  // 每5秒查询一次
    MaxRetries: 12,                // 最多查询12次
    OnCheck: func(retry int, status *paymentsdk.OrderStatusData) {
        fmt.Printf("[%d] 订单状态: %s\n", retry, paymentsdk.GetOrderStatusText(status.Status))
    },
})

if err != nil {
    fmt.Printf("轮询失败: %v\n", err)
    return
}

fmt.Println("支付成功！")
```

## API 文档

### 客户端配置

#### Config

```go
type Config struct {
    BaseURL    string        // API基础URL（必填）
    AppID      string        // 应用ID（必填）
    AppSecret  string        // 应用密钥（必填）
    Timeout    time.Duration // 请求超时时间（可选，默认30秒）
    HTTPClient *http.Client  // 自定义HTTP客户端（可选）
}
```

### 核心方法

#### CreatePayment - 创建支付订单

```go
func (c *Client) CreatePayment(req *PaymentRequest) (*PaymentData, error)
```

**参数：**
- `req.Subject` (string, 必填) - 商品名称
- `req.Amount` (float64, 必填) - 支付金额（元）
- `req.PayWay` (string, 必填) - 支付方式：`alipay`/`wechat`/`paypal`
- `req.OrderType` (string, 可选) - 订单类型
- `req.UserID` (string, 可选) - 用户ID
- `req.Extra` (string, 可选) - 额外信息（JSON格式）

**返回：**
- `PaymentData` - 支付数据，包含支付链接和订单号

#### QueryOrder - 查询订单状态

```go
func (c *Client) QueryOrder(orderNo string) (*OrderStatusData, error)
```

**参数：**
- `orderNo` (string) - 订单号

**返回：**
- `OrderStatusData` - 订单详细信息

#### PollOrderStatus - 轮询查询订单状态

```go
func (c *Client) PollOrderStatus(orderNo string, opts *PollOptions) (*OrderStatusData, error)
```

**参数：**
- `orderNo` (string) - 订单号
- `opts.Interval` (time.Duration) - 查询间隔，默认5秒
- `opts.MaxRetries` (int) - 最大重试次数，默认12次
- `opts.OnCheck` (func) - 每次查询的回调函数
- `opts.OnError` (func) - 查询出错的回调函数

#### GetOrderList - 获取订单列表

```go
func (c *Client) GetOrderList(req *OrderListRequest) (*OrderListResponse, error)
```

#### CancelOrder - 取消订单

```go
func (c *Client) CancelOrder(orderNo, reason string) error
```

#### RefundOrder - 申请退款

```go
func (c *Client) RefundOrder(req *RefundRequest) (*RefundResponse, error)
```

## 使用示例

### 支付宝支付

```go
req := &paymentsdk.PaymentRequest{
    Subject:   "测试商品",
    Amount:    0.01,
    PayWay:    paymentsdk.PayWayAlipay,
    OrderType: "product",
    UserID:    "user123",
}

paymentData, err := client.CreatePayment(req)
```

### 微信支付

```go
req := &paymentsdk.PaymentRequest{
    Subject:   "测试商品",
    Amount:    0.01,
    PayWay:    paymentsdk.PayWayWechat,
    OrderType: "product",
    UserID:    "user123",
}

paymentData, err := client.CreatePayment(req)
// 返回的 PayUrl 是微信支付二维码链接
```

### PayPal 支付

```go
req := &paymentsdk.PaymentRequest{
    Subject:   "Test Product",
    Amount:    1.00,
    PayWay:    paymentsdk.PayWayPaypal,
    Currency:  "USD",
    BrandName: "My Store",
    CancelURL: "https://example.com/cancel",
}

paymentData, err := client.CreatePayment(req)
```

### 获取订单列表

```go
listReq := &paymentsdk.OrderListRequest{
    UserID:   "user123",
    Status:   "2",  // 已支付
    Page:     1,
    PageSize: 10,
}

listResp, err := client.GetOrderList(listReq)
if err == nil {
    for _, order := range listResp.List {
        fmt.Printf("订单: %s, 金额: %.2f\n", order.OrderNo, order.Amount)
    }
}
```

### 取消订单

```go
err := client.CancelOrder(orderNo, "用户主动取消")
```

### 申请退款

```go
refundReq := &paymentsdk.RefundRequest{
    OutTradeNo:   orderNo,
    RefundAmount: 0.01,
    RefundReason: "商品质量问题",
}

refundResp, err := client.RefundOrder(refundReq)
```

## 订单状态

| 状态码 | 常量 | 说明 |
|--------|------|------|
| 1 | `OrderStatusNotPaid` | 未支付 |
| 2 | `OrderStatusPaidSuccess` | 已支付 |
| 201 | `OrderStatusClosed` | 已关闭 |

### 状态判断辅助方法

```go
orderStatus, _ := client.QueryOrder(orderNo)

if orderStatus.IsPaymentSuccess() {
    fmt.Println("支付成功")
}

if orderStatus.IsPending() {
    fmt.Println("待支付")
}

if orderStatus.IsClosed() {
    fmt.Println("已关闭")
}
```

## 错误处理

SDK 使用标准的 Go error 处理机制：

```go
paymentData, err := client.CreatePayment(req)
if err != nil {
    // 处理错误
    fmt.Printf("创建订单失败: %v\n", err)
    return
}

// 使用 paymentData
```

## 高级配置

### 自定义 HTTP 客户端

```go
import "net/http"

customClient := &http.Client{
    Timeout: 60 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:       10,
        IdleConnTimeout:    30 * time.Second,
        DisableCompression: true,
    },
}

config := &paymentsdk.Config{
    BaseURL:    "https://api.example.com",
    AppID:      "your-app-id",
    AppSecret:  "your-app-secret",
    HTTPClient: customClient,
}

client := paymentsdk.NewClient(config)
```

### 自定义轮询回调

```go
orderStatus, err := client.PollOrderStatus(orderNo, &paymentsdk.PollOptions{
    Interval:   3 * time.Second,
    MaxRetries: 20,
    OnCheck: func(retry int, status *paymentsdk.OrderStatusData) {
        log.Printf("[重试 %d] 订单 %s 状态: %s", 
            retry, status.OrderNo, paymentsdk.GetOrderStatusText(status.Status))
    },
    OnError: func(retry int, err error) {
        log.Printf("[重试 %d] 查询失败: %v", retry, err)
    },
})
```

## 线程安全

SDK 的所有方法都是线程安全的，可以在多个 goroutine 中并发使用同一个 `Client` 实例。

## 完整示例

参见 `example_usage.go` 文件。

## 许可证

MIT License

## 技术支持

如有问题，请提交 Issue 或联系技术支持。
