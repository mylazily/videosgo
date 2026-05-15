package collector

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/database"
	"github.com/mylazily/videosgo/internal/model"
	"gorm.io/gorm"
)

// CollectResult 采集结果统计
type CollectResult struct {
	Total   int
	New     int
	Updated int
	Failed  int
}

// Worker 采集 Worker 池
type Worker struct {
	client   *MacCMSClient
	parser   *Parser
	probe    *Probe
	db       *gorm.DB
	workers  int
	retryMax int
}

// NewWorker 创建 Worker 池
func NewWorker(client *MacCMSClient, parser *Parser, probe *Probe, db *gorm.DB, workers, retryMax int) *Worker {
	return &Worker{
		client:   client,
		parser:   parser,
		probe:    probe,
		db:       db,
		workers:  workers,
		retryMax: retryMax,
	}
}

// Collect 执行采集任务
func (w *Worker) Collect(source *model.CollectSource, incremental bool) (*CollectResult, error) {
	log.Printf("[采集] 开始采集源: %s (类型: %s)", source.Name, map[bool]string{true: "增量", false: "全量"}[incremental])

	// 获取所有视频数据
	items, err := w.client.FetchAllPages(source.APIURL, source.APIKey, incremental)
	if err != nil {
		return nil, fmt.Errorf("获取视频列表失败: %w", err)
	}

	log.Printf("[采集] 获取到 %d 条视频数据", len(items))

	result := &CollectResult{}
	result.Total = len(items)

	// 使用 Worker 池并发处理
	taskCh := make(chan MacCMSVideoItem, len(items))
	var wg sync.WaitGroup

	// 启动 Worker
	for i := 0; i < w.workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for item := range taskCh {
				w.processItem(item, source, result, workerID)
			}
		}(i)
	}

	// 分发任务
	for _, item := range items {
		taskCh <- item
	}
	close(taskCh)

	// 等待所有 Worker 完成
	wg.Wait()

	log.Printf("[采集] 采集完成: 总计=%d, 新增=%d, 更新=%d, 失败=%d",
		result.Total, result.New, result.Updated, result.Failed)

	return result, nil
}

// processItem 处理单个视频条目（带重试）
func (w *Worker) processItem(item MacCMSVideoItem, source *model.CollectSource, result *CollectResult, workerID int) {
	var err error
	for retry := 0; retry <= w.retryMax; retry++ {
		err = w.doProcessItem(item, source)
		if err == nil {
			return
		}
		if retry < w.retryMax {
			// 指数退避
			backoff := time.Duration(1<<uint(retry)) * time.Second
			log.Printf("[Worker-%d] 重试 %s (第%d次, 等待%v): %v", workerID, item.VodName, retry+1, backoff, err)
			time.Sleep(backoff)
		}
	}

	// 重试耗尽
	result.Failed++
	log.Printf("[Worker-%d] 处理失败 %s: %v", workerID, item.VodName, err)
}

// doProcessItem 实际处理单个视频条目
func (w *Worker) doProcessItem(item MacCMSVideoItem, source *model.CollectSource) error {
	// 归一化标题
	normalTitle := normalizeTitle(item.VodName)

	// 检查是否已存在
	var existing model.Video
	err := w.db.Where("title = ?", normalTitle).First(&existing).Error
	isNew := err == gorm.ErrRecordNotFound
	if err != nil && !isNew {
		return fmt.Errorf("查询数据库失败: %w", err)
	}

	// 解析播放链接
	playLinks := w.parser.ParsePlayGroups(item.VodPlayFrom, item.VodPlayUrl)

	// 过滤非 m3u8 链接
	playLinks = w.parser.FilterM3U8(playLinks)

	// m3u8 存活探针
	if w.probe != nil {
		playLinks = w.probe.FilterAliveLinks(playLinks)
	}

	// 将播放链接转为 JSONB 数组格式
	playLinksJSON := w.parser.ToJSONBArray(playLinks)

	// 解析标签
	tags := parseTags(item.VodTags)

	if isNew {
		// 新增视频
		video := &model.Video{
			Title:       normalTitle,
			SubTitle:    item.VodSub,
			Cover:       item.VodPic,
			Description: stripHTML(item.VodContent),
			CategoryID:  item.TypeID,
			Category:    item.TypeName,
			Year:        item.VodYear,
			Area:        item.VodArea,
			Director:    item.VodDirector,
			Actors:      item.VodActor,
			Tags:        tags,
			Remarks:     item.VodRemark,
			PlayLinks:   playLinksJSON,
			Status:      "active",
			SourceID:    source.ID,
		}

		if err := w.db.Create(video).Error; err != nil {
			return fmt.Errorf("创建视频失败: %w", err)
		}

		// 创建剧集
		w.createEpisodes(video.ID, playLinks, source.ID)

		// 记录到 Redis 热搜
		recordSearchHot(normalTitle)

		return nil
	}

	// 更新已有视频
	updates := map[string]interface{}{
		"cover":       item.VodPic,
		"description": stripHTML(item.VodContent),
		"category":    item.TypeName,
		"year":        item.VodYear,
		"area":        item.VodArea,
		"director":    item.VodDirector,
		"actors":      item.VodActor,
		"tags":        tags,
		"remarks":     item.VodRemark,
		"play_links":  playLinksJSON,
		"source_id":   source.ID,
	}

	if err := w.db.Model(&existing).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新视频失败: %w", err)
	}

	return nil
}

// createEpisodes 创建剧集
func (w *Worker) createEpisodes(videoID uuid.UUID, playLinks []PlayGroup, sourceID uuid.UUID) {
	for _, pg := range playLinks {
		for i, link := range pg.Links {
			episode := &model.Episode{
				VideoID:  videoID,
				Name:     fmt.Sprintf("%s 第%d集", pg.GroupName, i+1),
				EpIndex:  i + 1,
				URL:      link,
				URLType:  "m3u8",
				SourceID: sourceID,
			}
			if err := w.db.Create(episode).Error; err != nil {
				log.Printf("[采集] 创建剧集失败: %v", err)
			}
		}
	}
}

// recordSearchHot 记录热搜到 Redis
func recordSearchHot(keyword string) {
	if database.RDB == nil {
		return
	}
	ctx := context.Background()
	database.RDB.ZIncrBy(ctx, "search:hot", 1, keyword)
}

// normalizeTitle 归一化标题（去除多余空格等）
func normalizeTitle(title string) string {
	result := strings.TrimSpace(title)
	result = strings.Join(strings.Fields(result), " ")
	return result
}

// parseTags 解析标签字符串
func parseTags(tagsStr string) []string {
	if tagsStr == "" {
		return []string{}
	}
	tags := strings.Split(tagsStr, ",")
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			result = append(result, tag)
		}
	}
	return result
}

// stripHTML 去除 HTML 标签
func stripHTML(s string) string {
	// 简单实现：去除 < 和 > 之间的内容
	result := ""
	inTag := false
	for _, c := range s {
		if c == '<' {
			inTag = true
			continue
		}
		if c == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result += string(c)
		}
	}
	return strings.TrimSpace(result)
}
