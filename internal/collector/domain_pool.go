package collector

import (
	"net/url"
	"strings"
)

// PlayLink 播放链接结构
type PlayLink struct {
	SourceName string // 来源名称（如"最大资源"、"非凡资源"）
	M3U8URL    string // m3u8 播放地址
}

// DomainPoolExtractor 域名池提取器
type DomainPoolExtractor struct{}

// NewDomainPoolExtractor 创建域名池提取器
func NewDomainPoolExtractor() *DomainPoolExtractor {
	return &DomainPoolExtractor{}
}

// ExtractFromM3U8 从 m3u8 URL 提取域名和路径
// 输入: https://cdn1.ziyuan.com/2024/movie/hack4/index.m3u8
// 输出: domain="cdn1.ziyuan.com", path="/2024/movie/hack4/index.m3u8"
func (e *DomainPoolExtractor) ExtractFromM3U8(m3u8URL string) (domain, path string) {
	if m3u8URL == "" {
		return "", ""
	}

	u, err := url.Parse(m3u8URL)
	if err != nil {
		// 解析失败时尝试简单分割
		if idx := strings.Index(m3u8URL, "://"); idx > 0 {
			rest := m3u8URL[idx+3:]
			slashIdx := strings.Index(rest, "/")
			if slashIdx > 0 {
				return rest[:slashIdx], rest[slashIdx:]
			}
			return rest, ""
		}
		return "", ""
	}

	domain = u.Host
	path = u.Path

	// 去除端口号（如果有）
	if colonIdx := strings.LastIndex(domain, ":"); colonIdx > 0 {
		// 确保不是 IPv6 地址的一部分
		if !strings.Contains(domain[colonIdx:], "]") {
			domain = domain[:colonIdx]
		}
	}

	return domain, path
}

// ExtractFromPlayLinks 从多个播放链接中提取域名池
// 如果多个链接的路径相同（只是域名不同），返回域名池 + 共享路径
// 如果路径不同，返回 nil（不能混用）
func (e *DomainPoolExtractor) ExtractFromPlayLinks(links []PlayLink) (domainPool []string, sharedPath string) {
	if len(links) == 0 {
		return nil, ""
	}

	// 解析每个链接的域名和路径
	type domainPath struct {
		domain string
		path   string
	}

	parsed := make([]domainPath, 0, len(links))
	for _, link := range links {
		if link.M3U8URL == "" {
			continue
		}
		d, p := e.ExtractFromM3U8(link.M3U8URL)
		if d != "" {
			parsed = append(parsed, domainPath{domain: d, path: p})
		}
	}

	if len(parsed) == 0 {
		return nil, ""
	}

	// 只有一个链接，不需要域名池
	if len(parsed) == 1 {
		return []string{parsed[0].domain}, parsed[0].path
	}

	// 检查所有路径是否相同
	firstPath := parsed[0].path
	allSamePath := true
	for _, pp := range parsed[1:] {
		if pp.path != firstPath {
			allSamePath = false
			break
		}
	}

	if !allSamePath {
		// 路径不同，不能混用域名池
		return nil, ""
	}

	// 路径相同，提取去重后的域名池
	domainSet := make(map[string]bool)
	for _, pp := range parsed {
		domainSet[pp.domain] = true
	}

	pool := make([]string, 0, len(domainSet))
	for d := range domainSet {
		pool = append(pool, d)
	}

	return pool, firstPath
}

// ExtractDomainsFromPlayGroups 从播放组中提取所有 m3u8 链接的域名池
// 适用于从 Parser 解析出的 PlayGroup 结构
func (e *DomainPoolExtractor) ExtractDomainsFromPlayGroups(groups []PlayGroup) (domainPool []string, sharedPath string) {
	var links []PlayLink

	for _, group := range groups {
		for _, linkURL := range group.Links {
			// 处理 "第1集$url" 格式
			actualURL := linkURL
			if idx := strings.Index(linkURL, "$"); idx >= 0 {
				actualURL = linkURL[idx+1:]
			}
			links = append(links, PlayLink{
				SourceName: group.GroupName,
				M3U8URL:    actualURL,
			})
		}
	}

	return e.ExtractFromPlayLinks(links)
}

// BuildAlternateURLs 根据域名池和共享路径构建备用 URL 列表
// 输入: domainPool=["cdn1.com","cdn2.com"], sharedPath="/2024/movie.m3u8", scheme="https"
// 输出: ["https://cdn1.com/2024/movie.m3u8", "https://cdn2.com/2024/movie.m3u8"]
func (e *DomainPoolExtractor) BuildAlternateURLs(domainPool []string, sharedPath string, scheme string) []string {
	if len(domainPool) == 0 || sharedPath == "" {
		return nil
	}

	if scheme == "" {
		scheme = "https"
	}

	urls := make([]string, 0, len(domainPool))
	for _, domain := range domainPool {
		urls = append(urls, scheme+"://"+domain+sharedPath)
	}

	return urls
}
