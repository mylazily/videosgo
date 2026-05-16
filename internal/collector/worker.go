package collector

import (
	"context"
	"encoding/json"
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
	client        *MacCMSClient
	parser        *Parser
	probe         *Probe
	db            *gorm.DB
	workers       int
	retryMax      int
	titleCleaner  *TitleCleaner  // 标题清洗器
	domainExtractor *DomainPoolExtractor // 域名池提取器
}

// NewWorker 创建 Worker 池
func NewWorker(client *MacCMSClient, parser *Parser, probe *Probe, db *gorm.DB, workers, retryMax int) *Worker {
	return &Worker{
		client:          client,
		parser:          parser,
		probe:           probe,
		db:              db,
		workers:         workers,
		retryMax:        retryMax,
		titleCleaner:    NewTitleCleaner(),
		domainExtractor: NewDomainPoolExtractor(),
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
// 增强逻辑：使用标题清洗 + 模糊匹配去重，实现一剧多线聚合
func (w *Worker) doProcessItem(item MacCMSVideoItem, source *model.CollectSource) error {
	// 归一化标题（保留原始标题用于展示）
	normalTitle := normalizeTitle(item.VodName)

	// 使用标题清洗器获取纯净标题（用于去重匹配）
	cleanTitle := w.titleCleaner.Clean(item.VodName)

	// 解析播放链接
	playLinks := w.parser.ParsePlayGroups(item.VodPlayFrom, item.VodPlayUrl)

	// 过滤非 m3u8 链接
	playLinks = w.parser.FilterM3U8(playLinks)

	// m3u8 存活探针
	if w.probe != nil {
		playLinks = w.probe.FilterAliveLinks(playLinks)
	}

	// 将播放链接转为 JSONB 数组格式（兼容旧字段）
	playLinksJSON := w.parser.ToJSONBArray(playLinks)

	// 构建聚合播放线路（PlayLineJSON）
	playLines := w.buildPlayLines(playLinks, source.Name)

	// 解析标签
	tags := parseTags(item.VodTags)

	// ====== 增强去重逻辑 ======
	// 步骤 1: 精确匹配 clean_title
	var existing model.Video
	err := w.db.Where("clean_title = ? AND clean_title != ''", cleanTitle).First(&existing).Error
	isNew := err == gorm.ErrRecordNotFound
	if err != nil && !isNew {
		return fmt.Errorf("查询数据库失败: %w", err)
	}

	// 步骤 2: 如果精确匹配未命中，尝试模糊匹配
	if isNew {
		matched := w.findSimilarVideo(cleanTitle)
		if matched != nil {
			existing = *matched
			isNew = false
			log.Printf("[采集] 模糊匹配命中: \"%s\" -> 已有视频 ID=%s (相似度>0.7)",
				item.VodName, existing.ID)
		}
	}

	if isNew {
		// ====== 新增视频 ======
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
			// 新增字段
			CleanTitle:  cleanTitle,
			PlayLines:   playLines,
			SourceCount: 1,
		}

		// 尝试提取域名池（单源时也提取，为后续聚合做准备）
		domainPool, sharedPath := w.domainExtractor.ExtractDomainsFromPlayGroups(playLinks)
		if len(domainPool) > 1 && sharedPath != "" {
			video.DomainPool = domainPool
			video.SharedPath = sharedPath
		}

		if err := w.db.Create(video).Error; err != nil {
			return fmt.Errorf("创建视频失败: %w", err)
		}

		// 创建剧集
		w.createEpisodes(video.ID, playLinks, source.ID)

		// 记录到 Redis 热搜
		recordSearchHot(normalTitle)

		log.Printf("[采集] 新增视频: %s (clean: %s)", normalTitle, cleanTitle)
		return nil
	}

	// ====== 更新已有视频（一剧多线聚合） ======
	// 更新基础信息
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

	// 如果已有视频没有 clean_title，补充设置
	if existing.CleanTitle == "" {
		updates["clean_title"] = cleanTitle
	}

	if err := w.db.Model(&existing).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新视频失败: %w", err)
	}

	// 追加新的播放线路（去重）
	if err := w.appendPlayLines(existing.ID, playLines); err != nil {
		log.Printf("[采集] 追加播放线路失败 (视频ID=%s): %v", existing.ID, err)
	}

	// 增加资源站计数
	w.db.Model(&model.Video{}).Where("id = ?", existing.ID).
		UpdateColumn("source_count", gorm.Expr("source_count + 1"))

	// 尝试更新域名池
	w.tryUpdateDomainPool(existing.ID, playLinks)

	log.Printf("[采集] 聚合更新: %s -> 已有视频 ID=%s (来源: %s)",
		item.VodName, existing.ID, source.Name)

	return nil
}

// findSimilarVideo 在数据库中查找与给定 cleanTitle 相似的视频
// 使用前缀匹配缩小候选范围，再在应用层计算相似度
func (w *Worker) findSimilarVideo(cleanTitle string) *model.Video {
	if cleanTitle == "" || len(cleanTitle) < 2 {
		return nil
	}

	// 取前缀的前 4 个字符作为候选范围（中文字符约 2 个字）
	prefixLen := 4
	if len(cleanTitle) < prefixLen {
		prefixLen = len(cleanTitle)
	}
	prefix := cleanTitle[:prefixLen]

	// 查询候选视频
	var candidates []model.Video
	err := w.db.Where("clean_title LIKE ? AND clean_title != '' AND status = ?",
		prefix+"%", "active").
		Limit(50).
		Find(&candidates).Error
	if err != nil {
		return nil
	}

	// 在候选中查找相似度最高的视频
	var bestMatch *model.Video
	bestSimilarity := 0.7 // 阈值

	for i := range candidates {
		sim := w.titleCleaner.Similarity(cleanTitle, candidates[i].CleanTitle)
		if sim > bestSimilarity {
			bestSimilarity = sim
			bestMatch = &candidates[i]
		}
	}

	return bestMatch
}

// buildPlayLines 从播放组构建聚合播放线路
func (w *Worker) buildPlayLines(playLinks []PlayGroup, sourceName string) model.PlayLinesJSON {
	lines := make(model.PlayLinesJSON, 0)

	for _, group := range playLinks {
		for _, linkURL := range group.Links {
			// 处理 "第1集$url" 格式，提取实际 URL
			actualURL := linkURL
			if idx := strings.Index(linkURL, "$"); idx >= 0 {
				actualURL = linkURL[idx+1:]
			}

			// 提取域名和路径
			domain, path := w.domainExtractor.ExtractFromM3U8(actualURL)

			// 检测画质
			quality := detectQuality(linkURL)

			// 检测语言
			language := detectLanguage(linkURL)

			line := model.PlayLineJSON{
				SourceName: group.GroupName,
				M3U8URL:    actualURL,
				Domain:     domain,
				Path:       path,
				Format:     "m3u8",
				Quality:    quality,
				Language:   language,
			}
			lines = append(lines, line)
		}
	}

	return lines
}

// appendPlayLines 向已有视频追加播放线路（去重）
func (w *Worker) appendPlayLines(videoID uuid.UUID, newLines model.PlayLinesJSON) error {
	if len(newLines) == 0 {
		return nil
	}

	// 获取当前播放线路
	var video model.Video
	if err := w.db.Select("id, play_lines").First(&video, "id = ?", videoID).Error; err != nil {
		return err
	}

	// 构建已有 URL 集合用于去重
	existingURLs := make(map[string]bool)
	for _, line := range video.PlayLines {
		existingURLs[line.M3U8URL] = true
	}

	// 追加不重复的线路
	appended := false
	for _, line := range newLines {
		if !existingURLs[line.M3U8URL] {
			video.PlayLines = append(video.PlayLines, line)
			existingURLs[line.M3U8URL] = true
			appended = true
		}
	}

	if appended {
		return w.db.Model(&model.Video{}).Where("id = ?", videoID).
			Update("play_lines", video.PlayLines).Error
	}

	return nil
}

// tryUpdateDomainPool 尝试更新域名池
// 当视频的所有播放线路共享同一路径时，提取域名池
func (w *Worker) tryUpdateDomainPool(videoID uuid.UUID, playLinks []PlayGroup) {
	// 获取当前视频的播放线路
	var video model.Video
	if err := w.db.Select("id, play_lines, domain_pool, shared_path").First(&video, "id = ?", videoID).Error; err != nil {
		return
	}

	// 构建所有播放链接
	var allLinks []PlayLink
	for _, line := range video.PlayLines {
		allLinks = append(allLinks, PlayLink{
			SourceName: line.SourceName,
			M3U8URL:    line.M3U8URL,
		})
	}

	// 提取域名池
	domainPool, sharedPath := w.domainExtractor.ExtractFromPlayLinks(allLinks)
	if len(domainPool) > 1 && sharedPath != "" {
		// 有多个域名且共享路径，更新域名池
		poolJSON, err := json.Marshal(domainPool)
		if err != nil {
			return
		}
		w.db.Model(&model.Video{}).Where("id = ?", videoID).
			Updates(map[string]interface{}{
				"domain_pool": string(poolJSON),
				"shared_path": sharedPath,
			})
	}
}

// detectQuality 从链接文本中检测画质信息
func detectQuality(linkText string) string {
	text := strings.ToLower(linkText)
	switch {
	case strings.Contains(text, "4k") || strings.Contains(text, "2160"):
		return "4K"
	case strings.Contains(text, "1080"):
		return "1080P"
	case strings.Contains(text, "720"):
		return "720P"
	case strings.Contains(text, "480"):
		return "480P"
	default:
		return ""
	}
}

// detectLanguage 从链接文本中检测语言信息
func detectLanguage(linkText string) string {
	text := strings.ToLower(linkText)
	switch {
	case strings.Contains(text, "国语") || strings.Contains(text, "中字"):
		return "国语"
	case strings.Contains(text, "粤语"):
		return "粤语"
	case strings.Contains(text, "英语") || strings.Contains(text, "英文"):
		return "英语"
	case strings.Contains(text, "日语") || strings.Contains(text, "日文"):
		return "日语"
	case strings.Contains(text, "韩语") || strings.Contains(text, "韩文"):
		return "韩语"
	default:
		return ""
	}
}

// createEpisodes 创建剧集
func (w *Worker) createEpisodes(videoID uuid.UUID, playLinks []PlayGroup, sourceID uuid.UUID) {
	for _, pg := range playLinks {
		for i, link := range pg.Links {
			// Extract actual URL from link format "第1集$url"
			actualURL := link
			if idx := strings.Index(link, "$"); idx >= 0 {
				actualURL = link[idx+1:]
			}
			episode := &model.Episode{
				VideoID:  videoID,
				Name:     fmt.Sprintf("%s 第%d集", pg.GroupName, i+1),
				EpIndex:  i + 1,
				URL:      actualURL,
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
	// 使用 strings.Builder 提高性能
	var result strings.Builder
	result.Grow(len(s))
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
			result.WriteRune(c)
		}
	}
	return strings.TrimSpace(result.String())
}
