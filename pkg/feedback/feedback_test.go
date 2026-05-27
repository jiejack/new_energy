package feedback

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFeedbackManager(t *testing.T) {
	fm := NewFeedbackManager()
	require.NotNil(t, fm)
}

func TestFeedbackManager_CreateFeedback(t *testing.T) {
	fm := NewFeedbackManager()
	fb := Feedback{
		UserID:   "user1",
		Username: "Test User",
		Email:    "test@example.com",
		Type:     "bug",
		Title:    "Test Bug",
		Content:  "This is a test bug report",
		Priority: "high",
	}
	err := fm.CreateFeedback(context.Background(), fb)
	require.NoError(t, err)
}

func TestFeedbackManager_GetFeedback(t *testing.T) {
	fm := NewFeedbackManager()
	fb := Feedback{
		UserID: "user1",
		Type:   "bug",
		Title:  "Test",
	}
	fm.CreateFeedback(context.Background(), fb)

	feedbacks := fm.ListFeedbacks(context.Background(), nil)
	require.True(t, len(feedbacks) > 0)

	got, err := fm.GetFeedback(context.Background(), feedbacks[0].ID)
	require.NoError(t, err)
	assert.Equal(t, "Test", got.Title)
	assert.Equal(t, "pending", got.Status)

	_, err = fm.GetFeedback(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestFeedbackManager_UpdateFeedback(t *testing.T) {
	fm := NewFeedbackManager()
	fm.CreateFeedback(context.Background(), Feedback{UserID: "user1", Type: "bug", Title: "Original"})
	feedbacks := fm.ListFeedbacks(context.Background(), nil)

	fb := feedbacks[0]
	fb.Title = "Updated"
	err := fm.UpdateFeedback(context.Background(), fb)
	require.NoError(t, err)

	got, _ := fm.GetFeedback(context.Background(), fb.ID)
	assert.Equal(t, "Updated", got.Title)

	err = fm.UpdateFeedback(context.Background(), Feedback{ID: "nonexistent"})
	assert.Error(t, err)
}

func TestFeedbackManager_ListFeedbacks(t *testing.T) {
	fm := NewFeedbackManager()
	fm.CreateFeedback(context.Background(), Feedback{UserID: "user1", Type: "bug", Title: "Bug1"})
	fm.CreateFeedback(context.Background(), Feedback{UserID: "user2", Type: "feature", Title: "Feature1"})
	fm.CreateFeedback(context.Background(), Feedback{UserID: "user1", Type: "suggestion", Title: "Suggestion1"})

	all := fm.ListFeedbacks(context.Background(), nil)
	assert.Equal(t, 3, len(all))

	bugs := fm.ListFeedbacks(context.Background(), map[string]string{"type": "bug"})
	assert.Equal(t, 1, len(bugs))

	user1 := fm.ListFeedbacks(context.Background(), map[string]string{"user_id": "user1"})
	assert.Equal(t, 2, len(user1))
}

func TestFeedbackManager_AddComment(t *testing.T) {
	fm := NewFeedbackManager()
	fm.CreateFeedback(context.Background(), Feedback{UserID: "user1", Type: "bug", Title: "Test"})
	feedbacks := fm.ListFeedbacks(context.Background(), nil)

	comment := Comment{UserID: "user2", Username: "Commenter", Content: "Great find!"}
	err := fm.AddComment(context.Background(), feedbacks[0].ID, comment)
	require.NoError(t, err)

	got, _ := fm.GetFeedback(context.Background(), feedbacks[0].ID)
	assert.Equal(t, 1, len(got.Comments))
	assert.Equal(t, "Great find!", got.Comments[0].Content)

	err = fm.AddComment(context.Background(), "nonexistent", comment)
	assert.Error(t, err)
}

func TestFeedbackManager_UpdateStatus(t *testing.T) {
	fm := NewFeedbackManager()
	fm.CreateFeedback(context.Background(), Feedback{UserID: "user1", Type: "bug", Title: "Test"})
	feedbacks := fm.ListFeedbacks(context.Background(), nil)

	err := fm.UpdateStatus(context.Background(), feedbacks[0].ID, "in_progress")
	require.NoError(t, err)

	got, _ := fm.GetFeedback(context.Background(), feedbacks[0].ID)
	assert.Equal(t, "in_progress", got.Status)

	err = fm.UpdateStatus(context.Background(), "nonexistent", "closed")
	assert.Error(t, err)
}

func TestFeedbackManager_UpdatePriority(t *testing.T) {
	fm := NewFeedbackManager()
	fm.CreateFeedback(context.Background(), Feedback{UserID: "user1", Type: "bug", Title: "Test", Priority: "low"})
	feedbacks := fm.ListFeedbacks(context.Background(), nil)

	err := fm.UpdatePriority(context.Background(), feedbacks[0].ID, "critical")
	require.NoError(t, err)

	got, _ := fm.GetFeedback(context.Background(), feedbacks[0].ID)
	assert.Equal(t, "critical", got.Priority)

	err = fm.UpdatePriority(context.Background(), "nonexistent", "high")
	assert.Error(t, err)
}

func TestFeedbackManager_AssignFeedback(t *testing.T) {
	fm := NewFeedbackManager()
	fm.CreateFeedback(context.Background(), Feedback{UserID: "user1", Type: "bug", Title: "Test"})
	feedbacks := fm.ListFeedbacks(context.Background(), nil)

	err := fm.AssignFeedback(context.Background(), feedbacks[0].ID, "assignee1")
	require.NoError(t, err)

	got, _ := fm.GetFeedback(context.Background(), feedbacks[0].ID)
	assert.Equal(t, "assignee1", got.Assignee)

	err = fm.AssignFeedback(context.Background(), "nonexistent", "assignee1")
	assert.Error(t, err)
}

func TestFeedbackManager_AddTag(t *testing.T) {
	fm := NewFeedbackManager()
	fm.CreateFeedback(context.Background(), Feedback{UserID: "user1", Type: "bug", Title: "Test"})
	feedbacks := fm.ListFeedbacks(context.Background(), nil)

	err := fm.AddTag(context.Background(), feedbacks[0].ID, "urgent")
	require.NoError(t, err)

	got, _ := fm.GetFeedback(context.Background(), feedbacks[0].ID)
	assert.Contains(t, got.Tags, "urgent")

	err = fm.AddTag(context.Background(), feedbacks[0].ID, "urgent")
	require.NoError(t, err)

	got, _ = fm.GetFeedback(context.Background(), feedbacks[0].ID)
	count := 0
	for _, tag := range got.Tags {
		if tag == "urgent" {
			count++
		}
	}
	assert.Equal(t, 1, count)

	err = fm.AddTag(context.Background(), "nonexistent", "tag")
	assert.Error(t, err)
}

func TestFeedbackManager_RemoveTag(t *testing.T) {
	fm := NewFeedbackManager()
	fm.CreateFeedback(context.Background(), Feedback{UserID: "user1", Type: "bug", Title: "Test", Tags: []string{"urgent", "backend"}})
	feedbacks := fm.ListFeedbacks(context.Background(), nil)

	err := fm.RemoveTag(context.Background(), feedbacks[0].ID, "urgent")
	require.NoError(t, err)

	got, _ := fm.GetFeedback(context.Background(), feedbacks[0].ID)
	assert.NotContains(t, got.Tags, "urgent")
	assert.Contains(t, got.Tags, "backend")

	err = fm.RemoveTag(context.Background(), "nonexistent", "tag")
	assert.Error(t, err)
}

func TestFeedbackManager_GetStats(t *testing.T) {
	fm := NewFeedbackManager()
	fm.CreateFeedback(context.Background(), Feedback{UserID: "user1", Type: "bug", Title: "B1", Priority: "high"})
	fm.CreateFeedback(context.Background(), Feedback{UserID: "user2", Type: "feature", Title: "F1", Priority: "low"})
	fm.UpdateStatus(context.Background(), fm.ListFeedbacks(context.Background(), nil)[0].ID, "resolved")

	stats := fm.GetStats(context.Background())
	assert.Equal(t, 2, stats.Total)
	assert.Equal(t, 1, stats.ByType["bug"])
	assert.Equal(t, 1, stats.ByType["feature"])
	assert.Equal(t, 1, stats.ByStatus["resolved"])
	assert.Equal(t, 1, stats.ByPriority["high"])
}

func TestFeedbackManager_ExportFeedbacks(t *testing.T) {
	fm := NewFeedbackManager()

	jsonExport, err := fm.ExportFeedbacks(context.Background(), "json")
	require.NoError(t, err)
	assert.Contains(t, jsonExport, "feedbacks")

	csvExport, err := fm.ExportFeedbacks(context.Background(), "csv")
	require.NoError(t, err)
	assert.Contains(t, csvExport, "ID")

	_, err = fm.ExportFeedbacks(context.Background(), "xml")
	assert.Error(t, err)
}

func TestFeedbackManager_GetFeedbackTrends(t *testing.T) {
	fm := NewFeedbackManager()
	fm.CreateFeedback(context.Background(), Feedback{UserID: "user1", Type: "bug", Title: "B1"})
	fm.CreateFeedback(context.Background(), Feedback{UserID: "user2", Type: "feature", Title: "F1"})

	trends := fm.GetFeedbackTrends(context.Background(), 7)
	assert.NotNil(t, trends)
	assert.Equal(t, 7, len(trends["bug"]))
	assert.Equal(t, 7, len(trends["feature"]))
	assert.Equal(t, 7, len(trends["suggestion"]))
	assert.Equal(t, 7, len(trends["question"]))
}
