// Package paymentsdk 提供统一支付平台的 Golang SDK
package paymentsdk

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Client 支付客户端
type Client struct {
	baseURL    string
	appID      string
	appSecret  string
	httpClient *http.Client
}

// Config 客户端配置
type Config struct {
	BaseURL    string        // API基础URL，例如：http://localhost:8080
	AppID      string        // 应用ID
	AppSecret  string        // 应用密钥
	Timeout    time.Duration // 请求超时时间，默认30秒
	HTTPClient *http.Client  // 自定义HTTP客户端（可选）
}

// NewClient 创建支付客户端
func NewClient(config *Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: config.Timeout,
		}
	}

	return &Client{
		baseURL:    strings.TrimRight(config.BaseURL, "/"),
		appID:      config.AppID,
		appSecret:  config.AppSecret,
		httpClient: httpClient,
	}
}

// CreatePayment 创建支付订单
func (c *Client) CreatePayment(req *PaymentRequest) (*PaymentData, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := fmt.Sprintf("%s/api/v2/payment/pay", c.baseURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeaders(httpReq, bodyBytes)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var paymentResp PaymentResponse
	if err := json.Unmarshal(respBody, &paymentResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, body: %s", err, string(respBody))
	}

	if paymentResp.Code != 200 {
		return nil, fmt.Errorf("创建支付订单失败 (code=%d): %s", paymentResp.Code, paymentResp.Message)
	}

	return &paymentResp.Data, nil
}

// QueryOrder 查询订单状态
func (c *Client) QueryOrder(orderNo string) (*OrderStatusData, error) {
	url := fmt.Sprintf("%s/api/v2/payment/query/%s", c.baseURL, orderNo)
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	c.setAuthHeaders(httpReq, nil)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var queryResp OrderQueryResponse
	if err := json.Unmarshal(respBody, &queryResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, body: %s", err, string(respBody))
	}

	if queryResp.Code != 200 {
		return nil, fmt.Errorf("查询订单失败 (code=%d): %s", queryResp.Code, queryResp.Message)
	}

	return &queryResp.Data, nil
}

// GetOrderList 获取订单列表
func (c *Client) GetOrderList(req *OrderListRequest) (*OrderListResponse, error) {
	url := fmt.Sprintf("%s/api/v2/payment/orders", c.baseURL)
	
	// 构建查询参数
	params := url + "?"
	if req.UserID != "" {
		params += fmt.Sprintf("userId=%s&", req.UserID)
	}
	if req.Status != "" {
		params += fmt.Sprintf("status=%s&", req.Status)
	}
	if req.PayWay != "" {
		params += fmt.Sprintf("payWay=%s&", req.PayWay)
	}
	if req.Page > 0 {
		params += fmt.Sprintf("page=%d&", req.Page)
	}
	if req.PageSize > 0 {
		params += fmt.Sprintf("pageSize=%d&", req.PageSize)
	}
	params = strings.TrimSuffix(params, "&")

	httpReq, err := http.NewRequest("GET", params, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	c.setAuthHeaders(httpReq, nil)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var listResp struct {
		Code    int                `json:"code"`
		Message string             `json:"message"`
		Data    OrderListResponse  `json:"data"`
	}
	
	if err := json.Unmarshal(respBody, &listResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, body: %s", err, string(respBody))
	}

	if listResp.Code != 200 {
		return nil, fmt.Errorf("获取订单列表失败 (code=%d): %s", listResp.Code, listResp.Message)
	}

	return &listResp.Data, nil
}

// CancelOrder 取消订单
func (c *Client) CancelOrder(orderNo, reason string) error {
	reqBody := map[string]string{}
	if reason != "" {
		reqBody["cancelReason"] = reason
	}
	
	bodyBytes, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/api/v2/payment/cancel/%s", c.baseURL, orderNo)
	
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeaders(httpReq, bodyBytes)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 200 {
		return fmt.Errorf("取消订单失败 (code=%d): %s", result.Code, result.Message)
	}

	return nil
}

// RefundOrder 申请退款
func (c *Client) RefundOrder(req *RefundRequest) (*RefundResponse, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := fmt.Sprintf("%s/api/v2/payment/refund", c.baseURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeaders(httpReq, bodyBytes)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		Code    int            `json:"code"`
		Message string         `json:"message"`
		Data    RefundResponse `json:"data"`
	}
	
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("申请退款失败 (code=%d): %s", result.Code, result.Message)
	}

	return &result.Data, nil
}

// setAuthHeaders 设置认证请求头
func (c *Client) setAuthHeaders(req *http.Request, body []byte) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := generateNonce(16)

	params := map[string]string{
		"appId":     c.appID,
		"timestamp": timestamp,
		"nonce":     nonce,
	}

	sign := c.generateSign(params)

	req.Header.Set("X-App-Id", c.appID)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Nonce", nonce)
	req.Header.Set("X-Sign", sign)
}

// generateSign 生成API签名
func (c *Client) generateSign(params map[string]string) string {
	paramsCopy := make(map[string]string)
	for k, v := range params {
		if k != "sign" && v != "" {
			paramsCopy[k] = v
		}
	}

	var keys []string
	for k := range paramsCopy {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, paramsCopy[k]))
	}
	signString := strings.Join(parts, "&")

	h := hmac.New(sha256.New, []byte(c.appSecret))
	h.Write([]byte(signString))
	return hex.EncodeToString(h.Sum(nil))
}

// generateNonce 生成随机字符串
func generateNonce(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// PollOrderStatus 轮询查询订单状态，直到支付成功或超时
func (c *Client) PollOrderStatus(orderNo string, opts *PollOptions) (*OrderStatusData, error) {
	if opts == nil {
		opts = &PollOptions{
			Interval:   5 * time.Second,
			MaxRetries: 12,
		}
	}

	if opts.Interval == 0 {
		opts.Interval = 5 * time.Second
	}
	if opts.MaxRetries == 0 {
		opts.MaxRetries = 12
	}

	for i := 0; i < opts.MaxRetries; i++ {
		time.Sleep(opts.Interval)

		status, err := c.QueryOrder(orderNo)
		if err != nil {
			if opts.OnError != nil {
				opts.OnError(i+1, err)
			}
			continue
		}

		if opts.OnCheck != nil {
			opts.OnCheck(i+1, status)
		}

		// 支付成功
		if status.Status == OrderStatusPaidSuccess {
			return status, nil
		}

		// 订单已关闭
		if status.Status == OrderStatusClosed {
			return status, fmt.Errorf("订单已关闭")
		}
	}

	return nil, fmt.Errorf("查询超时，已达到最大重试次数 %d", opts.MaxRetries)
}
