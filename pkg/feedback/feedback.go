package feedback

import (
	"context"
	"fmt"
	"time"
)

// FeedbackManager 用户反馈管理器
type FeedbackManager struct {
	feedbacks      map[string]Feedback
	feedbackStats  map[string]int
}

// Feedback 用户反馈
type Feedback struct {
	ID          string
	UserID      string
	Username    string
	Email       string
	Type        string // bug, feature, suggestion, question
	Title       string
	Content     string
	Status      string // pending, in_progress, resolved, closed
	Priority    string // low, medium, high, critical
	Assignee    string
	Tags        []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Comments    []Comment
}

// Comment 反馈评论
type Comment struct {
	ID          string
	UserID      string
	Username    string
	Content     string
	CreatedAt   time.Time
}

// FeedbackStats 反馈统计
type FeedbackStats struct {
	Total       int
	ByType      map[string]int
	ByStatus    map[string]int
	ByPriority  map[string]int
}

// NewFeedbackManager 创建反馈管理器
func NewFeedbackManager() *FeedbackManager {
	return &FeedbackManager{
		feedbacks:     make(map[string]Feedback),
		feedbackStats: make(map[string]int),
	}
}

// CreateFeedback 创建反馈
func (fm *FeedbackManager) CreateFeedback(ctx context.Context, feedback Feedback) error {
	feedback.ID = fmt.Sprintf("feedback-%d", time.Now().UnixNano())
	feedback.Status = "pending"
	feedback.CreatedAt = time.Now()
	feedback.UpdatedAt = time.Now()
	fm.feedbacks[feedback.ID] = feedback

	// 更新统计
	fm.feedbackStats[feedback.Type]++

	fmt.Printf("Feedback created: %s - %s\n", feedback.ID, feedback.Title)
	return nil
}

// GetFeedback 获取反馈
func (fm *FeedbackManager) GetFeedback(ctx context.Context, id string) (Feedback, error) {
	feedback, exists := fm.feedbacks[id]
	if !exists {
		return Feedback{}, fmt.Errorf("feedback not found: %s", id)
	}
	return feedback, nil
}

// UpdateFeedback 更新反馈
func (fm *FeedbackManager) UpdateFeedback(ctx context.Context, feedback Feedback) error {
	existingFeedback, err := fm.GetFeedback(ctx, feedback.ID)
	if err != nil {
		return err
	}

	feedback.CreatedAt = existingFeedback.CreatedAt
	feedback.UpdatedAt = time.Now()
	fm.feedbacks[feedback.ID] = feedback

	fmt.Printf("Feedback updated: %s - %s\n", feedback.ID, feedback.Title)
	return nil
}

// ListFeedbacks 列出所有反馈
func (fm *FeedbackManager) ListFeedbacks(ctx context.Context, filter map[string]string) []Feedback {
	var result []Feedback

	for _, feedback := range fm.feedbacks {
		match := true
		for key, value := range filter {
			switch key {
			case "type":
				if feedback.Type != value {
					match = false
				}
			case "status":
				if feedback.Status != value {
					match = false
				}
			case "priority":
				if feedback.Priority != value {
					match = false
				}
			case "user_id":
				if feedback.UserID != value {
					match = false
				}
			}
		}

		if match {
			result = append(result, feedback)
		}
	}

	return result
}

// AddComment 添加评论
func (fm *FeedbackManager) AddComment(ctx context.Context, feedbackID string, comment Comment) error {
	feedback, err := fm.GetFeedback(ctx, feedbackID)
	if err != nil {
		return err
	}

	comment.ID = fmt.Sprintf("comment-%d", time.Now().UnixNano())
	comment.CreatedAt = time.Now()
	feedback.Comments = append(feedback.Comments, comment)
	feedback.UpdatedAt = time.Now()

	fm.feedbacks[feedbackID] = feedback

	fmt.Printf("Comment added to feedback: %s\n", feedbackID)
	return nil
}

// UpdateStatus 更新状态
func (fm *FeedbackManager) UpdateStatus(ctx context.Context, feedbackID string, status string) error {
	feedback, err := fm.GetFeedback(ctx, feedbackID)
	if err != nil {
		return err
	}

	feedback.Status = status
	feedback.UpdatedAt = time.Now()
	fm.feedbacks[feedbackID] = feedback

	fmt.Printf("Feedback status updated: %s - %s\n", feedbackID, status)
	return nil
}

// UpdatePriority 更新优先级
func (fm *FeedbackManager) UpdatePriority(ctx context.Context, feedbackID string, priority string) error {
	feedback, err := fm.GetFeedback(ctx, feedbackID)
	if err != nil {
		return err
	}

	feedback.Priority = priority
	feedback.UpdatedAt = time.Now()
	fm.feedbacks[feedbackID] = feedback

	fmt.Printf("Feedback priority updated: %s - %s\n", feedbackID, priority)
	return nil
}

// AssignFeedback 分配反馈
func (fm *FeedbackManager) AssignFeedback(ctx context.Context, feedbackID string, assignee string) error {
	feedback, err := fm.GetFeedback(ctx, feedbackID)
	if err != nil {
		return err
	}

	feedback.Assignee = assignee
	feedback.UpdatedAt = time.Now()
	fm.feedbacks[feedbackID] = feedback

	fmt.Printf("Feedback assigned: %s to %s\n", feedbackID, assignee)
	return nil
}

// AddTag 添加标签
func (fm *FeedbackManager) AddTag(ctx context.Context, feedbackID string, tag string) error {
	feedback, err := fm.GetFeedback(ctx, feedbackID)
	if err != nil {
		return err
	}

	// 检查标签是否已存在
	for _, existingTag := range feedback.Tags {
		if existingTag == tag {
			return nil
		}
	}

	feedback.Tags = append(feedback.Tags, tag)
	feedback.UpdatedAt = time.Now()
	fm.feedbacks[feedbackID] = feedback

	fmt.Printf("Tag added to feedback: %s - %s\n", feedbackID, tag)
	return nil
}

// RemoveTag 移除标签
func (fm *FeedbackManager) RemoveTag(ctx context.Context, feedbackID string, tag string) error {
	feedback, err := fm.GetFeedback(ctx, feedbackID)
	if err != nil {
		return err
	}

	newTags := []string{}
	for _, existingTag := range feedback.Tags {
		if existingTag != tag {
			newTags = append(newTags, existingTag)
		}
	}

	feedback.Tags = newTags
	feedback.UpdatedAt = time.Now()
	fm.feedbacks[feedbackID] = feedback

	fmt.Printf("Tag removed from feedback: %s - %s\n", feedbackID, tag)
	return nil
}

// GetStats 获取统计信息
func (fm *FeedbackManager) GetStats(ctx context.Context) FeedbackStats {
	stats := FeedbackStats{
		Total:      len(fm.feedbacks),
		ByType:     make(map[string]int),
		ByStatus:   make(map[string]int),
		ByPriority: make(map[string]int),
	}

	for _, feedback := range fm.feedbacks {
		stats.ByType[feedback.Type]++
		stats.ByStatus[feedback.Status]++
		stats.ByPriority[feedback.Priority]++
	}

	return stats
}

// ExportFeedbacks 导出反馈
func (fm *FeedbackManager) ExportFeedbacks(ctx context.Context, format string) (string, error) {
	switch format {
	case "json":
		// 实现JSON导出
		return "{\"feedbacks\": []}", nil
	case "csv":
		// 实现CSV导出
		return "ID,Type,Title,Status,Priority,CreatedAt\n", nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// GetFeedbackTrends 获取反馈趋势
func (fm *FeedbackManager) GetFeedbackTrends(ctx context.Context, days int) map[string][]int {
	trends := make(map[string][]int)

	// 初始化趋势数据
	trends["bug"] = make([]int, days)
	trends["feature"] = make([]int, days)
	trends["suggestion"] = make([]int, days)
	trends["question"] = make([]int, days)

	// 计算趋势
	now := time.Now()
	for _, feedback := range fm.feedbacks {
		daysSinceCreation := int(now.Sub(feedback.CreatedAt).Hours() / 24)
		if daysSinceCreation < days {
			trends[feedback.Type][daysSinceCreation]++
		}
	}

	return trends
}
