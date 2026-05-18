// Package collector MacCMS 采集器
package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"videosgo/internal/logger"
	"golang.org/x/time/rate"
)

// 全局速率限制器，每秒最多 10 个请求
var rateLimiter = rate.NewLimiter(rate.Every(100*time.Millisecond), 10)

// MacCMSClient MacCMS API 客户端
type MacCMSClient struct {
	client      *http.Client
	rateLimiter *rate.Limiter
}

// NewMacCMSClient 创建 MacCMS 客户端
func NewMacCMSClient(timeout time.Duration) *MacCMSClient {
	return &MacCMSClient{
		client: &http.Client{
			Timeout: timeout,
		},
		rateLimiter: rateLimiter,
	}
}

// MacCMSVideoListResponse MacCMS 视频列表 API 响应
type MacCMSVideoListResponse struct {
	Code      int               `json:"code"`
	Msg       string            `json:"msg"`
	Page      int               `json:"page"`
	PageCount int               `json:"pagecount"`
	Total     int               `json:"total"`
	List      []MacCMSVideoItem `json:"list"`
}

// MacCMSVideoItem MacCMS 视频条目
type MacCMSVideoItem struct {
	VodID       int    `json:"vod_id"`
	TypeID      int    `json:"type_id"`
	TypeName    string `json:"type_name"`
	VodName     string `json:"vod_name"`
	VodSub      string `json:"vod_sub"`
	VodEn       string `json:"vod_en"`
	VodStatus   int    `json:"vod_status"`
	VodPic      string `json:"vod_pic"`
	VodTags     string `json:"vod_tags"`
	VodClass    string `json:"vod_class"`
	VodRemark   string `json:"vod_remark"`
	VodYear     string `json:"vod_year"`
	VodArea     string `json:"vod_area"`
	VodLang     string `json:"vod_lang"`
	VodDirector string `json:"vod_director"`
	VodActor    string `json:"vod_actor"`
	VodContent  string `json:"vod_content"`
	VodPlayFrom string `json:"vod_play_from"`
	VodPlayUrl  string `json:"vod_play_url"`
	VodPlayNote string `json:"vod_play_note"`
	VodDownFrom string `json:"vod_down_from"`
	VodDownUrl  string `json:"vod_down_url"`
	VodTime     string `json:"vod_time"`
	VodTimeAdd  string `json:"vod_time_add"`
}

// IsValid 检查视频条目是否有效
func (item MacCMSVideoItem) IsValid() bool {
	return item.VodID > 0 && item.VodName != ""
}

// HasPlayLinks 检查是否有播放链接
func (item MacCMSVideoItem) HasPlayLinks() bool {
	return item.VodPlayFrom != "" && item.VodPlayUrl != ""
}

// FetchVideoList 获取视频列表
// incremental: true 增量采集（最近24小时），false 全量采集
func (c *MacCMSClient) FetchVideoList(ctx context.Context, apiURL, apiKey string, incremental bool, page int) (*MacCMSVideoListResponse, error) {
	// 等待速率限制
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("速率限制等待失败: %w", err)
	}

	// 构建请求 URL
	reqURL := fmt.Sprintf("%s?ac=videolist", strings.TrimSuffix(apiURL, "/"))
	if apiKey != "" {
		reqURL += fmt.Sprintf("&acode=%s", apiKey)
	}
	if incremental {
		reqURL += "&h=24"
	}
	if page > 0 {
		reqURL += fmt.Sprintf("&pg=%d", page)
	}

	logger.Debugf("[MacCMS] 请求 URL: %s", reqURL)

	// 创建带超时的请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")

	// 发送 HTTP 请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 MacCMS API 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MacCMS API 返回状态码: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查响应体是否为空
	if len(body) == 0 {
		return nil, fmt.Errorf("API 返回空响应")
	}

	// 尝试 JSON 解析
	var result MacCMSVideoListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		// 尝试解析错误响应
		var errResp map[string]interface{}
		if json.Unmarshal(body, &errResp) == nil {
			if msg, ok := errResp["msg"].(string); ok && msg != "" {
				return nil, fmt.Errorf("API 返回错误: %s", msg)
			}
		}
		return nil, fmt.Errorf("解析 JSON 失败: %w, 响应内容: %s", err, string(body[:min(len(body), 200)]))
	}

	// 检查响应码
	if result.Code != 1 && result.Code != 200 {
		return nil, fmt.Errorf("MacCMS API 错误: %s (code: %d)", result.Msg, result.Code)
	}

	// 过滤无效数据
	validItems := make([]MacCMSVideoItem, 0, len(result.List))
	for _, item := range result.List {
		if item.IsValid() && item.HasPlayLinks() {
			validItems = append(validItems, item)
		} else {
			logger.Warnf("[MacCMS] 跳过无效视频条目: ID=%d, Name=%s", item.VodID, item.VodName)
		}
	}
	result.List = validItems

	logger.Debugf("[MacCMS] 获取到 %d 条有效视频数据 (总计: %d)", len(validItems), result.Total)

	return &result, nil
}

// FetchAllPages 获取所有页的视频列表
func (c *MacCMSClient) FetchAllPages(apiURL, apiKey string, incremental bool) ([]MacCMSVideoItem, error) {
	ctx := context.Background()
	var allItems []MacCMSVideoItem
	page := 1
	maxPages := 1000 // 防止无限循环

	for {
		if page > maxPages {
			logger.Warnf("[MacCMS] 达到最大页数限制 (%d)，停止采集", maxPages)
			break
		}

		result, err := c.FetchVideoList(ctx, apiURL, apiKey, incremental, page)
		if err != nil {
			return nil, fmt.Errorf("获取第 %d 页失败: %w", page, err)
		}

		allItems = append(allItems, result.List...)

		logger.Infof("[MacCMS] 已获取第 %d/%d 页，共 %d 条数据", page, result.PageCount, len(result.List))

		// 检查是否还有更多页
		if page >= result.PageCount || result.PageCount == 0 {
			break
		}
		page++

		// 防止请求过快，添加额外延迟
		time.Sleep(500 * time.Millisecond)
	}

	logger.Infof("[MacCMS] 采集完成，共获取 %d 条视频数据", len(allItems))

	return allItems, nil
}

// FetchVideoDetail 获取视频详情
func (c *MacCMSClient) FetchVideoDetail(ctx context.Context, apiURL, apiKey string, vodID int) (*MacCMSVideoItem, error) {
	// 等待速率限制
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("速率限制等待失败: %w", err)
	}

	reqURL := fmt.Sprintf("%s?ac=detail&ids=%d", strings.TrimSuffix(apiURL, "/"), vodID)
	if apiKey != "" {
		reqURL += fmt.Sprintf("&acode=%s", apiKey)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求详情失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 返回状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		Code int               `json:"code"`
		Msg  string            `json:"msg"`
		List []MacCMSVideoItem `json:"list"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	if result.Code != 1 && result.Code != 200 {
		return nil, fmt.Errorf("API 错误: %s", result.Msg)
	}

	if len(result.List) == 0 {
		return nil, fmt.Errorf("视频不存在")
	}

	return &result.List[0], nil
}

// min 返回较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
