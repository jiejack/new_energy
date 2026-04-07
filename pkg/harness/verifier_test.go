package harness

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

// mockVerifier 是一个用于测试的 Verifier 实现
type mockVerifier struct {
	shouldFail     bool
	snapshotResult []byte
	verifyResult   bool
	verifyError    error
	snapshotError  error
}

func (m *mockVerifier) Verify(ctx context.Context, expected, actual interface{}) (bool, error) {
	if m.verifyError != nil {
		return false, m.verifyError
	}
	if m.shouldFail {
		return false, errors.New("verification failed")
	}
	return m.verifyResult, nil
}

func (m *mockVerifier) Snapshot(ctx context.Context, target interface{}) ([]byte, error) {
	if m.snapshotError != nil {
		return nil, m.snapshotError
	}
	if m.shouldFail {
		return nil, errors.New("snapshot failed")
	}
	if m.snapshotResult != nil {
		return m.snapshotResult, nil
	}
	// 默认行为：序列化为 JSON
	return json.Marshal(target)
}

func TestVerifier_Verify(t *testing.T) {
	tests := []struct {
		name       string
		verifier   Verifier
		expected   interface{}
		actual     interface{}
		wantResult bool
		wantErr    bool
		errMessage string
	}{
		{
			name:       "验证成功",
			verifier:   &mockVerifier{verifyResult: true},
			expected:   map[string]int{"value": 100},
			actual:     map[string]int{"value": 100},
			wantResult: true,
			wantErr:    false,
		},
		{
			name:       "验证失败",
			verifier:   &mockVerifier{verifyResult: false},
			expected:   map[string]int{"value": 100},
			actual:     map[string]int{"value": 200},
			wantResult: false,
			wantErr:    false,
		},
		{
			name:       "验证错误",
			verifier:   &mockVerifier{verifyError: errors.New("verification error")},
			expected:   nil,
			actual:     nil,
			wantResult: false,
			wantErr:    true,
			errMessage: "verification error",
		},
		{
			name:       "nil输入验证",
			verifier:   &mockVerifier{verifyResult: true},
			expected:   nil,
			actual:     nil,
			wantResult: true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := tt.verifier.Verify(ctx, tt.expected, tt.actual)

			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != tt.wantResult {
				t.Errorf("Verify() result = %v, want %v", result, tt.wantResult)
			}

			if tt.wantErr && err.Error() != tt.errMessage {
				t.Errorf("Verify() error message = %v, want %v", err.Error(), tt.errMessage)
			}
		})
	}
}

func TestVerifier_Snapshot(t *testing.T) {
	tests := []struct {
		name       string
		verifier   Verifier
		target     interface{}
		wantErr    bool
		errMessage string
	}{
		{
			name:     "快照成功",
			verifier: &mockVerifier{},
			target: struct {
				Name  string
				Value int
			}{
				Name:  "test",
				Value: 100,
			},
			wantErr: false,
		},
		{
			name:     "快照nil目标",
			verifier: &mockVerifier{},
			target:   nil,
			wantErr:  false,
		},
		{
			name:       "快照失败",
			verifier:   &mockVerifier{shouldFail: true},
			target:     "test",
			wantErr:    true,
			errMessage: "snapshot failed",
		},
		{
			name:       "快照错误",
			verifier:   &mockVerifier{snapshotError: errors.New("snapshot error")},
			target:     "test",
			wantErr:    true,
			errMessage: "snapshot error",
		},
		{
			name:       "自定义快照结果",
			verifier:   &mockVerifier{snapshotResult: []byte("custom snapshot")},
			target:     "test",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := tt.verifier.Snapshot(ctx, tt.target)

			if (err != nil) != tt.wantErr {
				t.Errorf("Snapshot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err.Error() != tt.errMessage {
				t.Errorf("Snapshot() error message = %v, want %v", err.Error(), tt.errMessage)
				return
			}

			if !tt.wantErr {
				if len(result) == 0 {
					t.Error("Snapshot() result is empty")
				}
			}
		})
	}
}

func TestVerifier_Interface_Compliance(t *testing.T) {
	// 确保 mockVerifier 实现了 Verifier 接口
	var _ Verifier = (*mockVerifier)(nil)
}

func TestVerifier_Context_Usage(t *testing.T) {
	verifier := &mockVerifier{verifyResult: true}

	t.Run("Verify with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// 即使上下文已取消，也应该能返回结果
		result, err := verifier.Verify(ctx, "expected", "actual")
		if err != nil {
			t.Errorf("Verify() with cancelled context returned error: %v", err)
		}
		if !result {
			t.Error("Verify() should return true")
		}
	})

	t.Run("Snapshot with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// 即使上下文已取消，也应该能返回结果
		_, err := verifier.Snapshot(ctx, "target")
		if err != nil {
			t.Errorf("Snapshot() with cancelled context returned error: %v", err)
		}
	})
}

// TestJSONVerifier 是一个使用 JSON 比较的验证器实现测试
type jsonVerifier struct{}

func (j *jsonVerifier) Verify(ctx context.Context, expected, actual interface{}) (bool, error) {
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		return false, err
	}
	actualJSON, err := json.Marshal(actual)
	if err != nil {
		return false, err
	}
	return string(expectedJSON) == string(actualJSON), nil
}

func (j *jsonVerifier) Snapshot(ctx context.Context, target interface{}) ([]byte, error) {
	return json.Marshal(target)
}

func TestJSONVerifier_Verify(t *testing.T) {
	verifier := &jsonVerifier{}

	tests := []struct {
		name     string
		expected interface{}
		actual   interface{}
		want     bool
		wantErr  bool
	}{
		{
			name:     "相同的map",
			expected: map[string]interface{}{"key": "value"},
			actual:   map[string]interface{}{"key": "value"},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "不同的map",
			expected: map[string]interface{}{"key": "value1"},
			actual:   map[string]interface{}{"key": "value2"},
			want:     false,
			wantErr:  false,
		},
		{
			name:     "相同的slice",
			expected: []int{1, 2, 3},
			actual:   []int{1, 2, 3},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "不同的slice",
			expected: []int{1, 2, 3},
			actual:   []int{1, 2, 4},
			want:     false,
			wantErr:  false,
		},
		{
			name:     "nil值",
			expected: nil,
			actual:   nil,
			want:     true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := verifier.Verify(ctx, tt.expected, tt.actual)

			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != tt.want {
				t.Errorf("Verify() result = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestJSONVerifier_Snapshot(t *testing.T) {
	verifier := &jsonVerifier{}

	tests := []struct {
		name    string
		target  interface{}
		wantErr bool
	}{
		{
			name:    "结构体快照",
			target:  struct{ Name string }{Name: "test"},
			wantErr: false,
		},
		{
			name:    "map快照",
			target:  map[string]int{"count": 100},
			wantErr: false,
		},
		{
			name:    "slice快照",
			target:  []string{"a", "b", "c"},
			wantErr: false,
		},
		{
			name:    "nil快照",
			target:  nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := verifier.Snapshot(ctx, tt.target)

			if (err != nil) != tt.wantErr {
				t.Errorf("Snapshot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result) == 0 {
					t.Error("Snapshot() result is empty")
				}

				// 验证结果是有效的 JSON
				var unmarshaled interface{}
				if err := json.Unmarshal(result, &unmarshaled); err != nil {
					t.Errorf("Snapshot() result is not valid JSON: %v", err)
				}
			}
		})
	}
}

// TestDeepEqualVerifier 是一个使用 reflect.DeepEqual 的验证器实现测试
type deepEqualVerifier struct{}

func (d *deepEqualVerifier) Verify(ctx context.Context, expected, actual interface{}) (bool, error) {
	return reflect.DeepEqual(expected, actual), nil
}

func (d *deepEqualVerifier) Snapshot(ctx context.Context, target interface{}) ([]byte, error) {
	return json.Marshal(target)
}

func TestDeepEqualVerifier_Verify(t *testing.T) {
	verifier := &deepEqualVerifier{}

	tests := []struct {
		name     string
		expected interface{}
		actual   interface{}
		want     bool
	}{
		{
			name:     "相同的结构体",
			expected: struct{ Name string }{Name: "test"},
			actual:   struct{ Name string }{Name: "test"},
			want:     true,
		},
		{
			name:     "不同的结构体",
			expected: struct{ Name string }{Name: "test1"},
			actual:   struct{ Name string }{Name: "test2"},
			want:     false,
		},
		{
			name:     "相同的slice",
			expected: []int{1, 2, 3},
			actual:   []int{1, 2, 3},
			want:     true,
		},
		{
			name:     "不同的slice",
			expected: []int{1, 2, 3},
			actual:   []int{3, 2, 1},
			want:     false,
		},
		{
			name:     "nil值",
			expected: nil,
			actual:   nil,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := verifier.Verify(ctx, tt.expected, tt.actual)

			if err != nil {
				t.Errorf("Verify() returned unexpected error: %v", err)
				return
			}

			if result != tt.want {
				t.Errorf("Verify() result = %v, want %v", result, tt.want)
			}
		})
	}
}
