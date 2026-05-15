package collector

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PlayGroup 播放组
type PlayGroup struct {
	GroupName string   `json:"group_name"`
	Links     []string `json:"links"`
}

// Parser 播放组解析器
type Parser struct{}

// NewParser 创建解析器
func NewParser() *Parser {
	return &Parser{}
}

// ParsePlayGroups 解析 MacCMS 播放组格式
// 输入格式: "播放组1$$$播放组2$$$播放组3"
// 每个播放组的链接格式: "链接1#链接2#链接3"
func (p *Parser) ParsePlayGroups(playFrom, playURL string) []PlayGroup {
	if playFrom == "" || playURL == "" {
		return nil
	}

	// 播放组名称，用 $$$ 分隔
	groupNames := strings.Split(playFrom, "$$$")
	// 播放链接，用 $$$ 分隔
	groupURLs := strings.Split(playURL, "$$$")

	var groups []PlayGroup

	for i := 0; i < len(groupNames) && i < len(groupURLs); i++ {
		name := strings.TrimSpace(groupNames[i])
		urls := strings.TrimSpace(groupURLs[i])

		if name == "" || urls == "" {
			continue
		}

		// 单个播放组内的链接用 # 分隔
		links := strings.Split(urls, "#")
		validLinks := make([]string, 0, len(links))

		for _, link := range links {
			link = strings.TrimSpace(link)
			if link != "" {
				validLinks = append(validLinks, link)
			}
		}

		if len(validLinks) > 0 {
			groups = append(groups, PlayGroup{
				GroupName: name,
				Links:     validLinks,
			})
		}
	}

	return groups
}

// FilterM3U8 过滤非 .m3u8 链接
func (p *Parser) FilterM3U8(groups []PlayGroup) []PlayGroup {
	var filtered []PlayGroup

	for _, group := range groups {
		var m3u8Links []string
		for _, link := range group.Links {
			if strings.HasSuffix(strings.ToLower(link), ".m3u8") {
				m3u8Links = append(m3u8Links, link)
			}
		}
		if len(m3u8Links) > 0 {
			filtered = append(filtered, PlayGroup{
				GroupName: group.GroupName,
				Links:     m3u8Links,
			})
		}
	}

	return filtered
}

// ToJSONBArray 将播放组转为 JSONB 数组格式
func (p *Parser) ToJSONBArray(groups []PlayGroup) []string {
	result := make([]string, 0, len(groups))
	for _, group := range groups {
		data, err := json.Marshal(group)
		if err != nil {
			continue
		}
		result = append(result, string(data))
	}
	return result
}

// ParsePlayURLFromJSONB 从 JSONB 数组解析播放链接
func (p *Parser) ParsePlayURLFromJSONB(jsonbArray []string) []PlayGroup {
	var groups []PlayGroup
	for _, item := range jsonbArray {
		var group PlayGroup
		if err := json.Unmarshal([]byte(item), &group); err != nil {
			continue
		}
		groups = append(groups, group)
	}
	return groups
}

// ExtractEpisodeName 从链接中提取集名
// 格式: "第1集$url" 或 "01$url"
func (p *Parser) ExtractEpisodeName(link string) string {
	parts := strings.Split(link, "$")
	if len(parts) >= 2 {
		return parts[0]
	}
	return ""
}

// ExtractURL 从链接中提取 URL
// 格式: "第1集$url"
func (p *Parser) ExtractURL(link string) string {
	parts := strings.Split(link, "$")
	if len(parts) >= 2 {
		return parts[1]
	}
	return link
}

// ParseEpisodeLinks 解析单集链接列表
// 格式: "第1集$url1#第2集$url2"
func (p *Parser) ParseEpisodeLinks(episodeStr string) []PlayGroup {
	if episodeStr == "" {
		return nil
	}

	links := strings.Split(episodeStr, "#")
	group := PlayGroup{
		GroupName: "默认",
		Links:     make([]string, 0, len(links)),
	}

	for _, link := range links {
		link = strings.TrimSpace(link)
		if link != "" {
			group.Links = append(group.Links, link)
		}
	}

	if len(group.Links) > 0 {
		return []PlayGroup{group}
	}

	return nil
}

// FormatPlayLinks 格式化播放链接为前端需要的格式
func (p *Parser) FormatPlayLinks(groups []PlayGroup) map[string]interface{} {
	result := make(map[string]interface{})
	for _, group := range groups {
		episodes := make([]map[string]string, 0, len(group.Links))
		for i, link := range group.Links {
			name := p.ExtractEpisodeName(link)
			url := p.ExtractURL(link)
			if name == "" {
				name = fmt.Sprintf("第%d集", i+1)
			}
			episodes = append(episodes, map[string]string{
				"name": name,
				"url":  url,
			})
		}
		result[group.GroupName] = episodes
	}
	return result
}
