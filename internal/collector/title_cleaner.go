package collector

import (
	"regexp"
	"strings"
	"unicode"
)

// TitleCleaner 标题清洗器
type TitleCleaner struct {
	suffixPatterns []*regexp.Regexp // 后缀匹配模式
	prefixPatterns []*regexp.Regexp // 前缀匹配模式
	yearRegex      *regexp.Regexp   // 年份匹配
	seasonRegex    *regexp.Regexp   // 季数匹配
	episodeRegex   *regexp.Regexp   // 集数匹配
}

// NewTitleCleaner 创建标题清洗器
func NewTitleCleaner() *TitleCleaner {
	// 常见后缀模式：画质、语言、版本、状态等
	suffixStrs := []string{
		// 画质相关
		`HD`, `hd`, `Hd`,
		`1080[Pp]`, `720[Pp]`, `480[Pp]`, `4[Kk]`, `2160[Pp]`,
		`蓝光`, `蓝光原盘`, `BD`, `bd`, `Bd`,
		`TC版`, `TS版`, `HDTC`, `HDT[Cc]`, `CAM`,
		`60帧`, `60fps`, `120帧`, `120fps`,
		`高清`, `超清`, `标清`, `极速`,
		`HDR`, `hdr`, `Dolby`, `dolby`, `杜比`, `IMAX`, `imax`,
		// 语言相关
		`国语`, `中字`, `中英双字`, `双语`, `双字幕`,
		`日语`, `英语`, `韩语`, `粤语`, `川话`, `东北话`,
		`台语`, `闽南语`, `泰语`, `法语`, `德语`, `俄语`,
		`日语中字`, `英语中字`, `韩语中字`,
		// 版本相关
		`无删减`, `加长版`, `导演剪辑版`, `未删减`, `完整版`,
		`院线版`, `网络版`, `电视版`, `DVD版`, `Blu-ray`,
		`修复版`, `重制版`, `珍藏版`, `终极版`,
		// 状态相关
		`完结`, `更新至.*`, `连载中`,
		`正片`, `预告片`, `花絮`, `片段`, `特辑`,
		// 集数相关（后缀）
		`第[零一二三四五六七八九十百千万\d]+集`,
		`第[零一二三四五六七八九十百千万\d]+期`,
		`EP?\d+`, `ep?\d+`,
		// 其他常见后缀
		`全集`, `合集`, `OVA`, `ova`,
		`电影版`, `剧场版`, `番外篇`, `前传`, `后传`,
		`会员版`, `付费版`, `抢先看`,
	}

	suffixPatterns := make([]*regexp.Regexp, 0, len(suffixStrs))
	for _, s := range suffixStrs {
		re := regexp.MustCompile(s)
		suffixPatterns = append(suffixPatterns, re)
	}

	// 常见前缀模式
	prefixStrs := []string{
		`^【.*?】`,       // 【xxx】
		`^《.*?》`,       // 《xxx》
		`^\[.*?\]`,      // [xxx]
		`^正片`,          // 正片
		`^HD`,           // HD
		`^[（(].*?[）)]`, // 全角/半角括号包裹
	}

	prefixPatterns := make([]*regexp.Regexp, 0, len(prefixStrs))
	for _, s := range prefixStrs {
		re := regexp.MustCompile(s)
		prefixPatterns = append(prefixPatterns, re)
	}

	return &TitleCleaner{
		suffixPatterns: suffixPatterns,
		prefixPatterns: prefixPatterns,
		yearRegex:      regexp.MustCompile(`(?:\((\d{4})\)|(\d{4})年|[\s\-_](\d{4})[\s\-_])`),
		seasonRegex:    regexp.MustCompile(`(?i)第([零一二三四五六七八九十\d]+)季|S(\d+)|Season\s*(\d+)`),
		episodeRegex:   regexp.MustCompile(`(?i)第([零一二三四五六七八九十百千万\d]+)[集期]|EP?(\d+)|第([零一二三四五六七八九十百千万\d]+)话`),
	}
}

// Clean 清洗标题，返回纯净的核心标题
// 处理逻辑：
//  1. 去除常见后缀：HD、国语、中字、1080P、蓝光、无删减等
//  2. 去除常见前缀：【】、《》、[]、正片、HD
//  3. 去除特殊符号：-、_、|、.
//  4. 去除多余空格
//  5. 统一全角半角（全角数字/字母转半角）
//  6. 去除年份后缀（如 (2024)、2024年）
//  7. 返回小写化结果用于比较
func (tc *TitleCleaner) Clean(title string) string {
	if title == "" {
		return ""
	}

	result := title

	// 步骤 1: 统一全角半角（全角数字/字母转半角）
	result = toHalfWidth(result)

	// 步骤 2: 去除常见前缀
	for _, re := range tc.prefixPatterns {
		result = re.ReplaceAllString(result, "")
	}

	// 步骤 3: 去除常见后缀
	for _, re := range tc.suffixPatterns {
		result = re.ReplaceAllString(result, "")
	}

	// 步骤 4: 去除年份后缀
	result = tc.yearRegex.ReplaceAllString(result, "")

	// 步骤 5: 去除特殊符号（标题中间的 -、_、|、.）
	result = strings.ReplaceAll(result, "-", " ")
	result = strings.ReplaceAll(result, "_", " ")
	result = strings.ReplaceAll(result, "|", " ")
	result = strings.ReplaceAll(result, "·", " ")
	// 去除标题中间的连续点号（保留正常标点）
	result = regexp.MustCompile(`\.{2,}`).ReplaceAllString(result, " ")
	// 去除单独的点号（但不在文件扩展名上下文中）
	result = regexp.MustCompile(`\s+\.\s+`).ReplaceAllString(result, " ")

	// 步骤 6: 去除多余空格
	result = strings.Join(strings.Fields(result), " ")

	// 步骤 7: 去除首尾空格
	result = strings.TrimSpace(result)

	// 步骤 8: 返回小写化结果用于比较
	result = strings.ToLower(result)

	return result
}

// ExtractYear 从标题中提取年份
func (tc *TitleCleaner) ExtractYear(title string) int {
	matches := tc.yearRegex.FindStringSubmatch(title)
	for _, m := range matches[1:] {
		if m != "" {
			year := 0
			for _, c := range m {
				year = year*10 + int(c-'0')
			}
			if year >= 1900 && year <= 2100 {
				return year
			}
		}
	}
	return 0
}

// ExtractSeason 从标题中提取季数
func (tc *TitleCleaner) ExtractSeason(title string) int {
	matches := tc.seasonRegex.FindStringSubmatch(title)
	for _, m := range matches[1:] {
		if m != "" {
			return chineseNumToInt(m)
		}
	}
	return 0
}

// ExtractEpisode 从标题中提取集数
func (tc *TitleCleaner) ExtractEpisode(title string) int {
	matches := tc.episodeRegex.FindStringSubmatch(title)
	for _, m := range matches[1:] {
		if m != "" {
			return chineseNumToInt(m)
		}
	}
	return 0
}

// IsSameVideo 判断两个标题是否是同一个视频
// 使用编辑距离(Levenshtein) + Jaccard 相似度双重判定
// 相似度阈值 > 0.7 认为是同一个视频
func (tc *TitleCleaner) IsSameVideo(title1, title2 string) bool {
	return tc.Similarity(title1, title2) > 0.7
}

// Similarity 计算两个标题的相似度 (0.0 ~ 1.0)
// 使用编辑距离和 Jaccard 相似度的加权平均
func (tc *TitleCleaner) Similarity(title1, title2 string) float64 {
	// 清洗两个标题
	clean1 := tc.Clean(title1)
	clean2 := tc.Clean(title2)

	if clean1 == "" || clean2 == "" {
		return 0.0
	}

	// 完全相同
	if clean1 == clean2 {
		return 1.0
	}

	// 计算编辑距离相似度
	editDist := levenshteinDistance(clean1, clean2)
	maxLen := max(len(clean1), len(clean2))
	if maxLen == 0 {
		return 1.0
	}
	editSimilarity := 1.0 - float64(editDist)/float64(maxLen)

	// 计算 Jaccard 相似度
	jaccardSim := jaccardSimilarity(clean1, clean2)

	// 加权平均：编辑距离权重 0.4，Jaccard 权重 0.6
	// Jaccard 对中文字符 bigram 更友好，给予更高权重
	return editSimilarity*0.4 + jaccardSim*0.6
}

// levenshteinDistance 计算编辑距离
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// 使用两行滚动数组优化空间复杂度
	prev := make([]int, len(s2)+1)
	curr := make([]int, len(s2)+1)

	for j := 0; j <= len(s2); j++ {
		prev[j] = j
	}

	for i := 1; i <= len(s1); i++ {
		curr[0] = i
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			curr[j] = min(
					min(prev[j]+1, curr[j-1]+1), // 删除, 插入
					prev[j-1]+cost,              // 替换
				)
		}
		prev, curr = curr, prev
	}

	return prev[len(s2)]
}

// jaccardSimilarity 计算 Jaccard 相似度（基于字符 bigram）
func jaccardSimilarity(s1, s2 string) float64 {
	bigrams1 := extractBigrams(s1)
	bigrams2 := extractBigrams(s2)

	if len(bigrams1) == 0 && len(bigrams2) == 0 {
		return 1.0
	}
	if len(bigrams1) == 0 || len(bigrams2) == 0 {
		return 0.0
	}

	// 计算交集
	intersection := 0
	set2 := make(map[string]bool)
	for _, b := range bigrams2 {
		set2[b] = true
	}
	for _, b := range bigrams1 {
		if set2[b] {
			intersection++
			delete(set2, b) // 避免重复计数
		}
	}

	// 计算并集
	union := len(bigrams1) + len(bigrams2) - intersection

	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// extractBigrams 提取字符 bigram 集合
func extractBigrams(s string) []string {
	runes := []rune(s)
	if len(runes) < 2 {
		return nil
	}

	bigrams := make([]string, 0, len(runes)-1)
	for i := 0; i < len(runes)-1; i++ {
		bigrams = append(bigrams, string(runes[i:i+2]))
	}
	return bigrams
}

// toHalfWidth 将全角数字和字母转换为半角
func toHalfWidth(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	for _, r := range s {
		// 全角空格
		if r == '\u3000' {
			result.WriteRune(' ')
			continue
		}
		// 全角数字 ０-９ (0xFF10 - 0xFF19) -> 0-9
		if r >= 0xFF10 && r <= 0xFF19 {
			result.WriteRune(r - 0xFF10 + '0')
			continue
		}
		// 全角大写字母 Ａ-Ｚ (0xFF21 - 0xFF3A) -> A-Z
		if r >= 0xFF21 && r <= 0xFF3A {
			result.WriteRune(r - 0xFF21 + 'A')
			continue
		}
		// 全角小写字母 ａ-ｚ (0xFF41 - 0xFF5A) -> a-z
		if r >= 0xFF41 && r <= 0xFF5A {
			result.WriteRune(r - 0xFF41 + 'a')
			continue
		}
		result.WriteRune(r)
	}

	return result.String()
}

// chineseNumToInt 中文数字转整数
func chineseNumToInt(s string) int {
	// 纯数字
	result := 0
	allDigit := true
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		} else {
			allDigit = false
			break
		}
	}
	if allDigit {
		return result
	}

	// 中文数字映射
	chineseMap := map[rune]int{
		'零': 0, '一': 1, '二': 2, '三': 3, '四': 4,
		'五': 5, '六': 6, '七': 7, '八': 8, '九': 9,
		'十': 10, '百': 100, '千': 1000, '万': 10000,
	}

	runes := []rune(s)
	n := len(runes)
	if n == 0 {
		return 0
	}

	// 处理简单情况："十"开头
	if runes[0] == '十' {
		val := 10
		if n > 1 {
			if v, ok := chineseMap[runes[1]]; ok && v < 10 {
				val += v
			}
		}
		return val
	}

	// 通用中文数字解析
	result = 0
	unit := 0
	for _, r := range runes {
		v, ok := chineseMap[r]
		if !ok {
			continue
		}
		if v >= 10 {
			// 单位：十、百、千、万
			if v == 10000 {
				result = (result + unit) * v
				unit = 0
			} else {
				unit = v
			}
		} else {
			// 数字
			if unit == 0 {
				unit = 1
			}
			result += v * unit
			unit = 0
		}
	}
	result += unit

	return result
}

// isChinese 判断字符是否是中文
func isChinese(r rune) bool {
	return unicode.Is(unicode.Han, r)
}
