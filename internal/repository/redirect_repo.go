package repository

import (
	"time"

	"github.com/google/uuid"
	"videosgo/internal/model"
	"gorm.io/gorm"
)

// RedirectRepo 重定向规则数据仓库
type RedirectRepo struct {
	db *gorm.DB
}

// NewRedirectRepo 创建重定向规则仓库
func NewRedirectRepo(db *gorm.DB) *RedirectRepo {
	return &RedirectRepo{db: db}
}

// Create 创建重定向规则
func (r *RedirectRepo) Create(rule *model.RedirectRule) error {
	return r.db.Create(rule).Error
}

// Update 更新重定向规则
func (r *RedirectRepo) Update(rule *model.RedirectRule) error {
	return r.db.Save(rule).Error
}

// Delete 删除重定向规则
func (r *RedirectRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.RedirectRule{}, "id = ?", id).Error
}

// GetByID 根据 ID 获取规则
func (r *RedirectRepo) GetByID(id uuid.UUID) (*model.RedirectRule, error) {
	var rule model.RedirectRule
	err := r.db.First(&rule, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// List 获取规则列表（分页）
func (r *RedirectRepo) List(page, pageSize int, activeOnly bool) ([]model.RedirectRule, int64, error) {
	var rules []model.RedirectRule
	var total int64

	db := r.db.Model(&model.RedirectRule{})
	if activeOnly {
		db = db.Where("is_active = ?", true)
	}

	db.Count(&total)

	err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("priority DESC, created_at DESC").
		Find(&rules).Error
	return rules, total, err
}

// MatchRule 匹配最优规则
func (r *RedirectRepo) MatchRule(domain, path, ua, ip string) (*model.RedirectRule, error) {
	var rules []model.RedirectRule

	now := time.Now()

	// 查询匹配域名且启用、未过期的规则，按优先级降序
	err := r.db.Where(
		"source_domain = ? AND is_active = ? AND (expires_at IS NULL OR expires_at > ?)",
		domain, true, now,
	).Order("priority DESC").
		Find(&rules).Error
	if err != nil {
		return nil, err
	}

	// 按优先级匹配路径
	for i := range rules {
		if matchPath(rules[i].RuleType, rules[i].SourcePath, path) {
			// 检查附加条件（如果有）
			condMap, _ := rules[i].Conditions.ToMap()
			if matchConditions(condMap, ua, ip) {
				return &rules[i], nil
			}
		}
	}

	return nil, gorm.ErrRecordNotFound
}

// matchPath 根据规则类型匹配路径
func matchPath(ruleType, pattern, path string) bool {
	switch ruleType {
	case "exact":
		return pattern == path
	case "prefix":
		return len(path) >= len(pattern) && path[:len(pattern)] == pattern
	case "wildcard":
		// 简单通配符匹配：支持 * 作为通配符
		return wildcardMatch(pattern, path)
	default:
		return pattern == path
	}
}

// wildcardMatch 简单通配符匹配
func wildcardMatch(pattern, str string) bool {
	if pattern == "*" {
		return true
	}
	if pattern == "" {
		return str == ""
	}
	// 简单实现：将 * 替换为 .*
	// 这里使用简单的逐字符匹配
	return simpleWildcard(pattern, str)
}

// simpleWildcard 简单通配符匹配实现
func simpleWildcard(pattern, str string) bool {
	px, sx := 0, 0
	for px < len(pattern) && sx < len(str) {
		if pattern[px] == '*' {
			// 跳过连续的 *
			for px < len(pattern) && pattern[px] == '*' {
				px++
			}
			if px >= len(pattern) {
				return true
			}
			// 找到下一个匹配位置
			for sx < len(str) && str[sx] != pattern[px] {
				sx++
			}
		} else if pattern[px] == str[sx] {
			px++
			sx++
		} else {
			return false
		}
	}
	// 处理尾部通配符
	for px < len(pattern) && pattern[px] == '*' {
		px++
	}
	return px == len(pattern) && sx == len(str)
}

// matchConditions 检查附加条件
func matchConditions(conditions map[string]interface{}, ua, ip string) bool {
	if conditions == nil {
		return true
	}
	// 检查 UA 条件
	if uaPattern, ok := conditions["ua_pattern"].(string); ok && uaPattern != "" {
		// 简单包含匹配
		if ua == "" || !containsStr(ua, uaPattern) {
			return false
		}
	}
	// 检查 IP 条件
	if ipPattern, ok := conditions["ip_pattern"].(string); ok && ipPattern != "" {
		if ip == "" || !containsStr(ip, ipPattern) {
			return false
		}
	}
	return true
}

// containsStr 字符串包含检查
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && containsSubstr(s, substr)
}

// containsSubstr 子串检查
func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// IncrementHitCount 增加命中次数
func (r *RedirectRepo) IncrementHitCount(ruleID uuid.UUID) error {
	return r.db.Model(&model.RedirectRule{}).Where("id = ?", ruleID).
		UpdateColumn("hit_count", gorm.Expr("hit_count + 1")).Error
}

// LogHit 记录命中日志
func (r *RedirectRepo) LogHit(log *model.RedirectHitLog) error {
	return r.db.Create(log).Error
}

// ListHitLogs 获取命中日志
func (r *RedirectRepo) ListHitLogs(ruleID uuid.UUID, limit int) ([]model.RedirectHitLog, error) {
	var logs []model.RedirectHitLog
	err := r.db.Where("rule_id = ?", ruleID).
		Order("hit_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// ToggleRule 切换规则启用状态
func (r *RedirectRepo) ToggleRule(id uuid.UUID, active bool) error {
	return r.db.Model(&model.RedirectRule{}).Where("id = ?", id).
		Update("is_active", active).Error
}
