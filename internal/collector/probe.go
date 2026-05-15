package collector

import (
	"log"
	"net/http"
	"sync"
	"time"
)

// Probe m3u8 存活探针
type Probe struct {
	client  *http.Client
	workers int
}

// NewProbe 创建探针
func NewProbe(timeout time.Duration, workers int) *Probe {
	return &Probe{
		client: &http.Client{
			Timeout: timeout,
			// 不跟随重定向
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		workers: workers,
	}
}

// FilterAliveLinks 过滤存活的 m3u8 链接
func (p *Probe) FilterAliveLinks(groups []PlayGroup) []PlayGroup {
	var aliveGroups []PlayGroup

	for _, group := range groups {
		aliveLinks := p.probeLinks(group.Links)
		if len(aliveLinks) > 0 {
			aliveGroups = append(aliveGroups, PlayGroup{
				GroupName: group.GroupName,
				Links:     aliveLinks,
			})
		}
	}

	return aliveGroups
}

// probeLinks 并发探活链接列表
func (p *Probe) probeLinks(links []string) []string {
	if len(links) == 0 {
		return nil
	}

	// 少量链接直接串行探测
	if len(links) <= 3 {
		var alive []string
		for _, link := range links {
			if p.isAlive(link) {
				alive = append(alive, link)
			}
		}
		return alive
	}

	// 大量链接使用 goroutine pool 并发探测
	taskCh := make(chan string, len(links))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var alive []string

	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for link := range taskCh {
				if p.isAlive(link) {
					mu.Lock()
					alive = append(alive, link)
					mu.Unlock()
				}
			}
		}()
	}

	for _, link := range links {
		taskCh <- link
	}
	close(taskCh)

	wg.Wait()

	return alive
}

// isAlive 检查链接是否存活（HTTP HEAD 请求）
func (p *Probe) isAlive(url string) bool {
	resp, err := p.client.Head(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// 状态码 200 才算存活
	return resp.StatusCode == http.StatusOK
}

// ProbeSingle 单个链接探活（供外部调用）
func (p *Probe) ProbeSingle(url string) bool {
	return p.isAlive(url)
}

// ProbeStats 探活统计
type ProbeStats struct {
	Total int
	Alive int
	Dead  int
}

// ProbeWithStats 带统计的探活
func (p *Probe) ProbeWithStats(groups []PlayGroup) ([]PlayGroup, ProbeStats) {
	var totalLinks int
	for _, group := range groups {
		totalLinks += len(group.Links)
	}

	aliveGroups := p.FilterAliveLinks(groups)

	var aliveLinks int
	for _, group := range aliveGroups {
		aliveLinks += len(group.Links)
	}

	stats := ProbeStats{
		Total: totalLinks,
		Alive: aliveLinks,
		Dead:  totalLinks - aliveLinks,
	}

	log.Printf("[探活] 总计=%d, 存活=%d, 失效=%d", stats.Total, stats.Alive, stats.Dead)

	return aliveGroups, stats
}
