package main

import (
	"fmt"
	"math/rand"
	"time"

	paymentsdk "github.com/difyz9/pay-unify/examples/payment-sdk"
)

func main() {
	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())

	// 创建支付客户端
	client := paymentsdk.NewClient(&paymentsdk.Config{
        BaseURL:   "https://api.example.com",
	    AppID:     "your-app-id",
        AppSecret: "your-app-secret",
	})

	fmt.Println("=== 统一支付平台 SDK 使用示例 ===\n")

	// ========== 示例 1: 创建支付宝支付订单 ==========
	fmt.Println("【示例 1】创建支付宝支付订单")
	paymentReq := &paymentsdk.PaymentRequest{
		Subject:   "VIP会员-月卡",
		Amount:    0.01,
		PayWay:    paymentsdk.PayWayAlipay,
		OrderType: "vip",
		UserID:    "user_12345",
		Extra:     `{"productId":"vip_001","period":"30days"}`,
	}

	paymentData, err := client.CreatePayment(paymentReq)
	if err != nil {
		fmt.Printf("❌ 创建订单失败: %v\n\n", err)
		return
	}

	fmt.Printf("✅ 订单创建成功\n")
	fmt.Printf("   订单号: %s\n", paymentData.OrderNo)
	fmt.Printf("   支付方式: %s\n", paymentData.PayWay)
	fmt.Printf("   支付金额: ¥%.2f\n", paymentData.Amount)
	fmt.Printf("   支付链接: %s\n\n", paymentData.PayUrl)

	orderNo := paymentData.OrderNo

	// ========== 示例 2: 单次查询订单状态 ==========
	fmt.Println("【示例 2】查询订单状态")
	orderStatus, err := client.QueryOrder(orderNo)
	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n\n", err)
	} else {
		fmt.Printf("✅ 查询成功\n")
		fmt.Printf("   订单号: %s\n", orderStatus.OrderNo)
		fmt.Printf("   状态: %s (%d)\n", paymentsdk.GetOrderStatusText(orderStatus.Status), orderStatus.Status)
		fmt.Printf("   金额: ¥%.2f\n\n", orderStatus.Amount)
	}

	// ========== 示例 3: 轮询查询订单状态 ==========
	fmt.Println("【示例 3】轮询查询订单状态")
	fmt.Println("请在浏览器中完成支付，程序将自动检测支付状态...\n")

	finalStatus, err := client.PollOrderStatus(orderNo, &paymentsdk.PollOptions{
		Interval:   5 * time.Second,
		MaxRetries: 12,
		OnCheck: func(retry int, status *paymentsdk.OrderStatusData) {
			fmt.Printf("   [%d/12] %s\n", retry, paymentsdk.GetOrderStatusText(status.Status))
		},
		OnError: func(retry int, err error) {
			fmt.Printf("   [%d/12] 查询出错: %v\n", retry, err)
		},
	})

	if err != nil {
		fmt.Printf("\n⚠️  轮询结束: %v\n\n", err)
	} else if finalStatus.IsPaymentSuccess() {
		fmt.Printf("\n✅ 支付成功！\n")
		fmt.Printf("   订单号: %s\n", finalStatus.OrderNo)
		fmt.Printf("   商品: %s\n", finalStatus.Subject)
		fmt.Printf("   金额: ¥%.2f\n", finalStatus.Amount)
		fmt.Printf("   支付平台交易号: %s\n", finalStatus.TradeNo)
		fmt.Printf("   支付时间: %s\n\n", time.Unix(finalStatus.PayTime, 0).Format("2006-01-02 15:04:05"))
	}

	// ========== 示例 4: 获取订单列表 ==========
	fmt.Println("【示例 4】获取订单列表")
	listReq := &paymentsdk.OrderListRequest{
		Page:     1,
		PageSize: 5,
	}

	listResp, err := client.GetOrderList(listReq)
	if err != nil {
		fmt.Printf("❌ 获取列表失败: %v\n\n", err)
	} else {
		fmt.Printf("✅ 共找到 %d 条订单\n", listResp.Total)
		for i, order := range listResp.List {
			fmt.Printf("   %d. 订单号: %s | 金额: ¥%.2f | 状态: %s\n",
				i+1, order.OrderNo, order.Amount, paymentsdk.GetOrderStatusText(order.Status))
		}
		fmt.Println()
	}

	// ========== 示例 5: 创建微信支付订单 ==========
	fmt.Println("【示例 5】创建微信支付订单")
	wechatReq := &paymentsdk.PaymentRequest{
		Subject:   "测试商品",
		Amount:    0.01,
		PayWay:    paymentsdk.PayWayWechat,
		OrderType: "product",
		UserID:    "user_12345",
	}

	wechatData, err := client.CreatePayment(wechatReq)
	if err != nil {
		fmt.Printf("❌ 创建微信订单失败: %v\n\n", err)
	} else {
		fmt.Printf("✅ 微信订单创建成功\n")
		fmt.Printf("   订单号: %s\n", wechatData.OrderNo)
		fmt.Printf("   二维码链接: %s\n", wechatData.PayUrl)
		fmt.Println("   提示: 请使用微信扫描二维码完成支付\n")
	}

	// ========== 示例 6: 取消订单（如果未支付）==========
	if orderStatus != nil && orderStatus.IsPending() {
		fmt.Println("【示例 6】取消订单")
		err = client.CancelOrder(orderNo, "用户主动取消")
		if err != nil {
			fmt.Printf("❌ 取消订单失败: %v\n\n", err)
		} else {
			fmt.Printf("✅ 订单已取消: %s\n\n", orderNo)
		}
	}

	fmt.Println("=== 示例完成 ===")
}
