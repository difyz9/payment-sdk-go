package paymentsdk

import "time"

// PaymentRequest 支付请求结构
type PaymentRequest struct {
	Subject   string  `json:"subject"`             // 支付主题/商品名称
	Amount    float64 `json:"amount"`              // 支付金额
	PayWay    string  `json:"payWay"`              // 支付方式 alipay/wechat/paypal
	OrderType string  `json:"orderType,omitempty"` // 订单类型
	UserID    string  `json:"userId,omitempty"`    // 用户ID（可选）
	Extra     string  `json:"extra,omitempty"`     // 额外信息（JSON格式）
	Currency  string  `json:"currency,omitempty"`  // 货币代码（PayPal支付时使用，默认USD）
	BrandName string  `json:"brandName,omitempty"` // 品牌名称（PayPal支付时显示）
	CancelURL string  `json:"cancelUrl,omitempty"` // 取消支付返回地址（PayPal支付时使用）
}

// PaymentResponse 支付响应结构
type PaymentResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    PaymentData `json:"data"`
}

// PaymentData 支付数据
type PaymentData struct {
	PayUrl   string  `json:"payUrl"`   // 支付链接
	PayWay   string  `json:"payWay"`   // 支付方式
	Amount   float64 `json:"amount"`   // 支付金额
	OrderNo  string  `json:"orderNo"`  // 订单号
	OrderID  string  `json:"orderId"`  // PayPal订单ID（仅PayPal支付返回）
	Currency string  `json:"currency"` // 货币代码（仅PayPal支付返回）
}

// OrderQueryResponse 订单查询响应
type OrderQueryResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    OrderStatusData `json:"data"`
}

// OrderStatusData 订单状态数据
type OrderStatusData struct {
	OrderNo   string  `json:"orderNo"`   // 订单号
	Subject   string  `json:"subject"`   // 商品名称
	Amount    float64 `json:"amount"`    // 订单金额
	Status    int     `json:"status"`    // 订单状态：1-未支付，2-已支付，201-已关闭
	PayWay    string  `json:"payWay"`    // 支付方式
	TradeNo   string  `json:"tradeNo"`   // 支付平台交易号
	PayTime   int64   `json:"payTime"`   // 支付时间戳
	OrderType string  `json:"orderType"` // 订单类型
	UserID    string  `json:"userId"`    // 用户ID
	Extra     string  `json:"extra"`     // 额外信息
	CreatedAt string  `json:"createdAt"` // 创建时间
	UpdatedAt string  `json:"updatedAt"` // 更新时间
}

// OrderListRequest 订单列表请求参数
type OrderListRequest struct {
	UserID   string `json:"userId,omitempty"`   // 用户ID
	Status   string `json:"status,omitempty"`   // 订单状态
	PayWay   string `json:"payWay,omitempty"`   // 支付方式
	Page     int    `json:"page,omitempty"`     // 页码
	PageSize int    `json:"pageSize,omitempty"` // 每页数量
}

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	List     []OrderStatusData `json:"list"`     // 订单列表
	Total    int64             `json:"total"`    // 总数
	Page     int               `json:"page"`     // 当前页
	PageSize int               `json:"pageSize"` // 每页数量
}

// RefundRequest 退款请求参数
type RefundRequest struct {
	OutTradeNo   string  `json:"outTradeNo"`   // 商户订单号
	RefundAmount float64 `json:"refundAmount"` // 退款金额（元）
	RefundReason string  `json:"refundReason"` // 退款原因
}

// RefundResponse 退款响应
type RefundResponse struct {
	Message         string      `json:"message"`         // 响应消息
	OrderNo         string      `json:"orderNo"`         // 订单号
	PayWay          string      `json:"payWay"`          // 支付方式
	TradeNo         string      `json:"tradeNo"`         // 支付平台交易号
	RefundAmount    float64     `json:"refundAmount"`    // 退款金额
	RefundRequestNo string      `json:"refundRequestNo"` // 退款请求号
	Result          interface{} `json:"result"`          // 退款结果详情
}

// PollOptions 轮询查询配置
type PollOptions struct {
	Interval   time.Duration                           // 查询间隔，默认5秒
	MaxRetries int                                     // 最大重试次数，默认12次
	OnCheck    func(retry int, status *OrderStatusData) // 每次查询后的回调
	OnError    func(retry int, err error)              // 查询出错时的回调
}

// 订单状态常量
const (
	OrderStatusNotPaid     = 1   // 未支付
	OrderStatusPaidSuccess = 2   // 已支付
	OrderStatusClosed      = 201 // 已关闭
)

// 支付方式常量
const (
	PayWayAlipay = "alipay" // 支付宝
	PayWayWechat = "wechat" // 微信支付
	PayWayPaypal = "paypal" // PayPal
)

// GetOrderStatusText 获取订单状态文本
func GetOrderStatusText(status int) string {
	switch status {
	case OrderStatusNotPaid:
		return "未支付"
	case OrderStatusPaidSuccess:
		return "已支付"
	case OrderStatusClosed:
		return "已关闭"
	default:
		return "未知状态"
	}
}

// IsPaymentSuccess 判断订单是否支付成功
func (o *OrderStatusData) IsPaymentSuccess() bool {
	return o.Status == OrderStatusPaidSuccess
}

// IsClosed 判断订单是否已关闭
func (o *OrderStatusData) IsClosed() bool {
	return o.Status == OrderStatusClosed
}

// IsPending 判断订单是否待支付
func (o *OrderStatusData) IsPending() bool {
	return o.Status == OrderStatusNotPaid
}
