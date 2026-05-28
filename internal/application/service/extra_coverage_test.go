package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQAService_GetAIResponse_NilDialogueMgr(t *testing.T) {
	svc := &QAService{dialogueMgr: nil}
	result, err := svc.getAIResponse(context.Background(), "session-1", "test question")
	assert.NoError(t, err)
	assert.Contains(t, result, "AI服务暂未配置")
}
