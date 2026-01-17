# 支付宝 ReturnURL 使用指南

## 概述

`ReturnURL` 是支付宝支付完成后，用户浏览器跳转回商户网站的同步返回地址。通过自定义 `ReturnURL`，您可以灵活控制支付成功后的用户体验。

## 基础用法

```go
req := &paymentsdk.PaymentRequest{
    Subject:   "商品名称",
    Amount:    99.00,
    PayWay:    paymentsdk.PayWayAlipay,
    ReturnURL: "https://mystore.com/payment/success", // 自定义返回地址
    OrderType: "product",
    UserID:    "user_12345",
}

paymentData, err := client.CreatePayment(req)
```

## 工作原理

```
┌──────────┐        ┌──────────┐        ┌──────────┐
│  用户    │        │  支付宝  │        │  商户    │
└────┬─────┘        └────┬─────┘        └────┬─────┘
     │                   │                   │
     │  1. 点击支付       │                   │
     ├──────────────────>│                   │
     │                   │                   │
     │  2. 完成支付       │                   │
     │<──────────────────┤                   │
     │                   │                   │
     │  3. 跳转到ReturnURL (同步返回)         │
     ├───────────────────────────────────────>│
     │                   │                   │
     │                   │  4. 异步通知(可靠)  │
     │                   ├──────────────────>│
```

## 使用场景

### 1️⃣ 不同商品跳转不同页面

```go
// VIP充值 - 跳转到会员中心
vipReq := &paymentsdk.PaymentRequest{
    Subject:   "VIP会员月卡",
    Amount:    30.00,
    PayWay:    paymentsdk.PayWayAlipay,
    ReturnURL: "https://mystore.com/vip/center",
    OrderType: "vip",
}

// 商品购买 - 跳转到订单详情
productReq := &paymentsdk.PaymentRequest{
    Subject:   "iPhone 15 Pro",
    Amount:    7999.00,
    PayWay:    paymentsdk.PayWayAlipay,
    ReturnURL: "https://mystore.com/order/detail?id=12345",
    OrderType: "product",
}
```

### 2️⃣ 携带自定义参数

```go
// 在URL中携带业务参数
req := &paymentsdk.PaymentRequest{
    Subject:   "课程购买",
    Amount:    299.00,
    PayWay:    paymentsdk.PayWayAlipay,
    ReturnURL: "https://edu.com/success?course_id=101&from=alipay&utm_source=promo",
    OrderType: "course",
}

// 用户支付成功后会跳转到：
// https://edu.com/success?course_id=101&from=alipay&utm_source=promo&out_trade_no=xxx&...
```

### 3️⃣ 移动端和PC端区分

```go
// 根据用户设备选择不同的返回地址
var returnURL string
if isMobile {
    returnURL = "https://m.mystore.com/payment/success"
} else {
    returnURL = "https://www.mystore.com/payment/success"
}

req := &paymentsdk.PaymentRequest{
    Subject:   "商品购买",
    Amount:    199.00,
    PayWay:    paymentsdk.PayWayAlipay,
    ReturnURL: returnURL,
    OrderType: "product",
}
```

### 4️⃣ 小程序/APP 中的使用

```go
// 小程序场景 - 跳转到小程序页面（需要配置支付宝小程序跳转）
req := &paymentsdk.PaymentRequest{
    Subject:   "商品购买",
    Amount:    99.00,
    PayWay:    paymentsdk.PayWayAlipay,
    ReturnURL: "alipays://platformapi/startapp?appId=xxx&page=pages/success",
    OrderType: "product",
}
```

## 支付宝返回的参数

用户支付完成后，支付宝会在 `ReturnURL` 后追加以下参数：

| 参数名 | 说明 | 示例 |
|--------|------|------|
| `out_trade_no` | 商户订单号 | `202601171234567890` |
| `trade_no` | 支付宝交易号 | `2026011722001457271454630352` |
| `total_amount` | 订单金额 | `0.01` |
| `timestamp` | 通知时间 | `2026-01-17 17:49:07` |
| `sign` | 签名 | `xxx...` |
| `sign_type` | 签名类型 | `RSA2` |
| `charset` | 编码格式 | `utf-8` |
| `seller_id` | 卖家支付宝用户ID | `2088xxx` |

**示例完整URL：**
```
https://mystore.com/payment/success?
  out_trade_no=202601171234567890&
  trade_no=2026011722001457271454630352&
  total_amount=0.01&
  timestamp=2026-01-17+17:49:07&
  sign=xxx...&
  sign_type=RSA2
```

## 返回页面处理建议

### ✅ 推荐做法

```go
// 在返回页面中，再次验证订单状态
func PaymentSuccessHandler(w http.ResponseWriter, r *http.Request) {
    // 1. 从URL获取订单号
    outTradeNo := r.URL.Query().Get("out_trade_no")
    
    // 2. 调用SDK查询订单状态（确保订单真的支付成功）
    orderStatus, err := client.QueryOrder(outTradeNo)
    if err != nil {
        // 显示错误页面
        http.Error(w, "查询订单失败", http.StatusInternalServerError)
        return
    }
    
    // 3. 验证订单状态
    if !orderStatus.IsPaymentSuccess() {
        // 订单未支付成功，显示等待页面或错误提示
        renderWaitingPage(w, outTradeNo)
        return
    }
    
    // 4. 订单已支付，显示成功页面
    renderSuccessPage(w, orderStatus)
}
```

### ❌ 不推荐做法

```go
// ❌ 错误：仅根据URL参数判断支付成功
func BadHandler(w http.ResponseWriter, r *http.Request) {
    outTradeNo := r.URL.Query().Get("out_trade_no")
    
    // 没有验证订单状态，直接认为支付成功
    // 这样不安全，因为用户可以伪造URL参数
    renderSuccessPage(w, outTradeNo)
}
```

## 注意事项

### ⚠️ 重要提示

1. **不要依赖同步返回** - `ReturnURL` 仅用于页面展示，订单状态应以异步通知为准
2. **必须验证订单** - 在返回页面中，应调用 `QueryOrder()` 再次验证订单状态
3. **HTTPS 协议** - 生产环境必须使用 HTTPS
4. **公网可访问** - ReturnURL 必须是公网可以访问的地址
5. **参数编码** - URL 参数会被自动编码，注意处理特殊字符
6. **不要携带敏感信息** - 不要在 ReturnURL 中携带密码等敏感数据

### 🔒 安全建议

```go
// ✅ 推荐：使用订单号作为参数
ReturnURL: "https://mystore.com/success?order_no=xxx"

// ❌ 不推荐：携带敏感信息
ReturnURL: "https://mystore.com/success?user_password=xxx"

// ✅ 推荐：在服务端验证
func Handler(w http.ResponseWriter, r *http.Request) {
    orderNo := r.URL.Query().Get("order_no")
    
    // 从数据库或缓存中获取订单信息
    order := getOrderFromDB(orderNo)
    
    // 验证订单归属
    if order.UserID != currentUserID {
        http.Error(w, "无权访问", http.StatusForbidden)
        return
    }
    
    // 查询最新状态
    status, _ := client.QueryOrder(orderNo)
    // ...
}
```

## 默认行为

如果不设置 `ReturnURL`：

```go
// 未设置 ReturnURL
req := &paymentsdk.PaymentRequest{
    Subject:   "商品购买",
    Amount:    99.00,
    PayWay:    paymentsdk.PayWayAlipay,
    // ReturnURL: "",  // 不设置
    OrderType: "product",
}

// 系统会使用服务端配置文件中的默认返回地址
// 例如：https://api.example.com/payment/return
```

## 完整示例

### 创建订单

```go
package main

import (
    "fmt"
    paymentsdk "github.com/difyz9/payment-sdk-go"
)

func CreateOrder() {
    client := paymentsdk.NewClient(&paymentsdk.Config{
        BaseURL:   "https://api.example.com",
        AppID:     "your-app-id",
        AppSecret: "your-app-secret",
    })
    
    req := &paymentsdk.PaymentRequest{
        Subject:   "VIP会员年卡",
        Amount:    299.00,
        PayWay:    paymentsdk.PayWayAlipay,
        ReturnURL: "https://mystore.com/vip/success?plan=yearly&promo=new2026",
        OrderType: "vip",
        UserID:    "user_12345",
        Extra:     `{"plan":"yearly","discount":"new2026"}`,
    }
    
    paymentData, err := client.CreatePayment(req)
    if err != nil {
        fmt.Printf("创建订单失败: %v\n", err)
        return
    }
    
    fmt.Printf("支付链接: %s\n", paymentData.PayUrl)
    fmt.Printf("订单号: %s\n", paymentData.OrderNo)
}
```

### 返回页面处理

```go
package main

import (
    "fmt"
    "net/http"
    paymentsdk "github.com/difyz9/payment-sdk-go"
)

var client *paymentsdk.Client

func init() {
    client = paymentsdk.NewClient(&paymentsdk.Config{
        BaseURL:   "https://api.example.com",
        AppID:     "your-app-id",
        AppSecret: "your-app-secret",
    })
}

func SuccessHandler(w http.ResponseWriter, r *http.Request) {
    // 1. 获取URL参数
    outTradeNo := r.URL.Query().Get("out_trade_no")
    plan := r.URL.Query().Get("plan")
    promo := r.URL.Query().Get("promo")
    
    if outTradeNo == "" {
        http.Error(w, "订单号为空", http.StatusBadRequest)
        return
    }
    
    // 2. 验证订单状态
    orderStatus, err := client.QueryOrder(outTradeNo)
    if err != nil {
        http.Error(w, "查询订单失败: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    // 3. 检查支付状态
    if !orderStatus.IsPaymentSuccess() {
        // 显示等待页面
        fmt.Fprintf(w, `
            <html>
            <head><title>等待支付</title></head>
            <body>
                <h1>支付处理中...</h1>
                <p>订单号: %s</p>
                <p>状态: %s</p>
                <script>
                    // 每5秒刷新一次页面
                    setTimeout(function(){ location.reload(); }, 5000);
                </script>
            </body>
            </html>
        `, outTradeNo, paymentsdk.GetOrderStatusText(orderStatus.Status))
        return
    }
    
    // 4. 显示支付成功页面
    fmt.Fprintf(w, `
        <html>
        <head><title>支付成功</title></head>
        <body>
            <h1>✅ 支付成功！</h1>
            <p>订单号: %s</p>
            <p>商品: %s</p>
            <p>金额: ¥%.2f</p>
            <p>套餐: %s</p>
            <p>优惠码: %s</p>
            <a href="/vip/center">进入会员中心</a>
        </body>
        </html>
    `, orderStatus.OrderNo, orderStatus.Subject, orderStatus.Amount, plan, promo)
}

func main() {
    http.HandleFunc("/vip/success", SuccessHandler)
    http.ListenAndServe(":8080", nil)
}
```

## 常见问题

### Q1: ReturnURL 和异步通知有什么区别？

**A:** 
- **ReturnURL (同步返回)**: 用户支付完成后浏览器跳转，用于页面展示，不可靠（用户可能关闭浏览器）
- **异步通知 (NotifyURL)**: 支付宝服务器主动回调，用于更新订单状态，可靠且必须处理

### Q2: 用户可以伪造 ReturnURL 的参数吗？

**A:** 可以！所以必须在返回页面中调用 `QueryOrder()` 验证订单真实状态，不要仅依赖URL参数。

### Q3: ReturnURL 可以是局域网地址吗？

**A:** 
- **开发环境**: 可以使用 `http://localhost` 或内网地址
- **生产环境**: 必须是公网可访问的 HTTPS 地址

### Q4: 如果不设置 ReturnURL 会怎样？

**A:** 系统会使用服务端配置的默认返回地址。建议根据业务需求自定义 ReturnURL。

### Q5: ReturnURL 的参数会被覆盖吗？

**A:** 不会。支付宝会在您的 URL 后追加参数，不会覆盖已有参数。

```
您的URL: https://mystore.com/success?plan=vip
最终URL: https://mystore.com/success?plan=vip&out_trade_no=xxx&trade_no=xxx&...
```

## 相关文档

- [SDK 使用文档](./README.md)
- [完整示例代码](./example_usage.go)
- [支付宝开放平台文档](https://opendocs.alipay.com/)

## 技术支持

如有问题，请提交 Issue 或联系技术支持。
