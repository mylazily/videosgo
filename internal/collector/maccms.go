// Package collector MacCMS 采集器
package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// MacCMSClient MacCMS API 客户端
type MacCMSClient struct {
	client *http.Client
}

// NewMacCMSClient 创建 MacCMS 客户端
func NewMacCMSClient(timeout time.Duration) *MacCMSClient {
	return &MacCMSClient{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// MacCMSVideoListResponse MacCMS 视频列表 API 响应
type MacCMSVideoListResponse struct {
	Code    int                `json:"code"`
	Msg     string             `json:"msg"`
	Page    int                `json:"page"`
	PageCount int              `json:"pagecount"`
	Total   int                `json:"total"`
	List    []MacCMSVideoItem  `json:"list"`
}

// MacCMSVideoItem MacCMS 视频条目
type MacCMSVideoItem struct {
	VodID     int    `json:"vod_id"`
	TypeID    int    `json:"type_id"`
	TypeName  string `json:"type_name"`
	VodName   string `json:"vod_name"`
	VodSub    string `json:"vod_sub"`
	VodEn     string `json:"vod_en"`
	VodStatus int    `json:"vod_status"`
	VodPic    string `json:"vod_pic"`
	VodTags   string `json:"vod_tags"`
	VodClass  string `json:"vod_class"`
	VodRemark string `json:"vod_remark"`
	VodYear   string `json:"vod_year"`
	VodArea   string `json:"vod_area"`
	VodLang   string `json:"vod_lang"`
	VodDirector string `json:"vod_director"`
	VodActor  string `json:"vod_actor"`
	VodContent string `json:"vod_content"`
	VodPlayFrom string `json:"vod_play_from"`
	VodPlayUrl  string `json:"vod_play_url"`
	VodPlayNote string `json:"vod_play_note"`
	VodDownFrom string `json:"vod_down_from"`
	VodDownUrl  string `json:"vod_down_url"`
	VodTime    string `json:"vod_time"`
	VodTimeAdd string `json:"vod_time_add"`
}

// FetchVideoList 获取视频列表
// incremental: true 增量采集（最近24小时），false 全量采集
func (c *MacCMSClient) FetchVideoList(apiURL, apiKey string, incremental bool, page int) (*MacCMSVideoListResponse, error) {
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

	// 发送 HTTP 请求
	resp, err := c.client.Get(reqURL)
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

	// 尝试 JSON 解析
	var result MacCMSVideoListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	if result.Code != 1 && result.Code != 200 {
		return nil, fmt.Errorf("MacCMS API 错误: %s", result.Msg)
	}

	return &result, nil
}

// FetchAllPages 获取所有页的视频列表
func (c *MacCMSClient) FetchAllPages(apiURL, apiKey string, incremental bool) ([]MacCMSVideoItem, error) {
	var allItems []MacCMSVideoItem
	page := 1

	for {
		result, err := c.FetchVideoList(apiURL, apiKey, incremental, page)
		if err != nil {
			return nil, err
		}

		allItems = append(allItems, result.List...)

		// 检查是否还有更多页
		if page >= result.PageCount || result.PageCount == 0 {
			break
		}
		page++

		// 防止请求过快
		time.Sleep(500 * time.Millisecond)
	}

	return allItems, nil
}
