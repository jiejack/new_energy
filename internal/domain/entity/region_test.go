package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRegion(t *testing.T) {
	parentID := "parent001"
	tests := []struct {
		name     string
		code     string
		regionName string
		parentID *string
		level    int
		want     *Region
	}{
		{
			name:       "创建根区域",
			code:       "REGION001",
			regionName: "华东区",
			parentID:   nil,
			level:      1,
			want: &Region{
				Code:  "REGION001",
				Name:  "华东区",
				Level: 1,
			},
		},
		{
			name:       "创建子区域",
			code:       "REGION002",
			regionName: "江苏省",
			parentID:   &parentID,
			level:      2,
			want: &Region{
				Code:     "REGION002",
				Name:     "江苏省",
				ParentID: &parentID,
				Level:    2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRegion(tt.code, tt.regionName, tt.parentID, tt.level)
			assert.Equal(t, tt.want.Code, got.Code)
			assert.Equal(t, tt.want.Name, got.Name)
			assert.Equal(t, tt.want.Level, got.Level)
			if tt.parentID != nil {
				assert.Equal(t, *tt.want.ParentID, *got.ParentID)
			} else {
				assert.Nil(t, got.ParentID)
			}
		})
	}
}

func TestRegion_IsRoot(t *testing.T) {
	parentID := "parent001"
	
	// 根区域
	rootRegion := NewRegion("ROOT", "根区域", nil, 1)
	assert.True(t, rootRegion.IsRoot())

	// 子区域
	subRegion := NewRegion("SUB", "子区域", &parentID, 2)
	assert.False(t, subRegion.IsRoot())
}

func TestRegion_TableName(t *testing.T) {
	region := Region{}
	assert.Equal(t, "regions", region.TableName())
}

func TestNewSubRegion(t *testing.T) {
	subRegion := NewSubRegion("SUB001", "南京分区", "region001")

	assert.Equal(t, "SUB001", subRegion.Code)
	assert.Equal(t, "南京分区", subRegion.Name)
	assert.Equal(t, "region001", subRegion.RegionID)
}

func TestSubRegion_TableName(t *testing.T) {
	subRegion := SubRegion{}
	assert.Equal(t, "sub_regions", subRegion.TableName())
}
