package entity

import "time"

type Region struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Code        string    `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name        string    `json:"name" gorm:"type:varchar(200);not null"`
	ParentID    *string   `json:"parent_id" gorm:"type:varchar(36);index"`
	Level       int       `json:"level" gorm:"default:1"`
	SortOrder   int       `json:"sort_order" gorm:"default:0"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	SubRegions  []*Region `json:"sub_regions" gorm:"foreignKey:ParentID"`
	Stations    []*Station `json:"stations" gorm:"foreignKey:SubRegionID"`
}

func (r *Region) TableName() string {
	return "regions"
}

func NewRegion(code, name string, parentID *string, level int) *Region {
	return &Region{
		Code:     code,
		Name:     name,
		ParentID: parentID,
		Level:    level,
	}
}

func (r *Region) IsRoot() bool {
	return r.ParentID == nil
}

type SubRegion struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Code        string    `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name        string    `json:"name" gorm:"type:varchar(200);not null"`
	RegionID    string    `json:"region_id" gorm:"type:varchar(36);index;not null"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *SubRegion) TableName() string {
	return "sub_regions"
}

func NewSubRegion(code, name, regionID string) *SubRegion {
	return &SubRegion{
		Code:     code,
		Name:     name,
		RegionID: regionID,
	}
}
