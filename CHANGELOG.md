# Payment SDK 更新说明

## 🎉 新增功能：支付宝自定义 ReturnURL 支持

### 📝 更新内容

1. ✅ 在 `PaymentRequest` 中添加 `ReturnURL` 字段
2. ✅ 更新 API 文档说明
3. ✅ 更新示例代码
4. ✅ 新增详细的使用指南

### 📦 更新的文件

- `types.go` - 添加 ReturnURL 字段定义
- `README.md` - 更新 API 文档和使用示例
- `example_usage.go` - 更新示例代码
- `RETURN_URL_GUIDE.md` - 新增详细使用指南（NEW）

### 🚀 快速使用

```go
import paymentsdk "github.com/difyz9/payment-sdk"

// 创建客户端
client := paymentsdk.NewClient(&paymentsdk.Config{
    BaseURL:   "https://api.example.com",
    AppID:     "your-app-id",
    AppSecret: "your-app-secret",
})

// 创建支付订单（带自定义返回地址）
req := &paymentsdk.PaymentRequest{
    Subject:   "VIP会员年卡",
    Amount:    299.00,
    PayWay:    paymentsdk.PayWayAlipay,
    ReturnURL: "https://mystore.com/payment/success", // 🎯 自定义返回地址
    OrderType: "vip",
    UserID:    "user_12345",
}

paymentData, err := client.CreatePayment(req)
```

### 📚 文档链接

- **主文档**: [README.md](./README.md)
- **ReturnURL 详细指南**: [RETURN_URL_GUIDE.md](./RETURN_URL_GUIDE.md)
- **完整示例**: [example_usage.go](./example_usage.go)

### 💡 使用场景

1. **不同商品不同页面** - VIP充值跳转会员中心，商品购买跳转订单详情
2. **携带自定义参数** - 在URL中传递业务参数
3. **移动端/PC端区分** - 根据设备跳转不同页面
4. **A/B测试** - 不同用户跳转不同落地页

### ⚠️ 重要提示

1. **必须验证订单** - 在返回页面中调用 `QueryOrder()` 验证订单状态
2. **HTTPS 协议** - 生产环境必须使用 HTTPS
3. **同步 vs 异步** - ReturnURL 是同步返回（页面展示），订单状态以异步通知为准
4. **不携带敏感信息** - 不要在 URL 中传递密码等敏感数据

### 📖 详细说明

更多详细信息和最佳实践，请查看：
👉 [ReturnURL 使用指南](./RETURN_URL_GUIDE.md)

### 🔄 向后兼容

- ✅ ReturnURL 是可选参数，不设置则使用服务端默认配置
- ✅ 不影响现有代码，完全向后兼容
- ✅ 所有现有功能正常工作

---

**更新日期**: 2026-01-17
**版本**: v1.1.0
