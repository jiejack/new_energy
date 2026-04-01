package entity

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type OperationLog struct {
	ID           string                 `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID       string                 `json:"user_id" gorm:"type:varchar(36);index"`
	Username     string                 `json:"username" gorm:"type:varchar(100)"`
	Action       string                 `json:"action" gorm:"type:varchar(100);not null;index"`
	ResourceType string                 `json:"resource_type" gorm:"type:varchar(50)"`
	ResourceID   string                 `json:"resource_id" gorm:"type:varchar(36)"`
	Details      Details                `json:"details" gorm:"type:jsonb"`
	IPAddress    string                 `json:"ip_address" gorm:"type:varchar(50)"`
	UserAgent    string                 `json:"user_agent" gorm:"type:varchar(500)"`
	CreatedAt    time.Time              `json:"created_at"`
}

type Details map[string]interface{}

func (d Details) Value() (driver.Value, error) {
	if d == nil {
		return nil, nil
	}
	return json.Marshal(d)
}

func (d *Details) Scan(value interface{}) error {
	if value == nil {
		*d = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, d)
}

func (o *OperationLog) TableName() string {
	return "operation_logs"
}

func NewOperationLog(userID, username, action string) *OperationLog {
	return &OperationLog{
		UserID:   userID,
		Username: username,
		Action:   action,
	}
}

func (o *OperationLog) SetResource(resourceType, resourceID string) {
	o.ResourceType = resourceType
	o.ResourceID = resourceID
}

func (o *OperationLog) SetDetails(details Details) {
	o.Details = details
}

func (o *OperationLog) SetRequestInfo(ipAddress, userAgent string) {
	o.IPAddress = ipAddress
	o.UserAgent = userAgent
}

const (
	ActionLogin         = "login"
	ActionLogout        = "logout"
	ActionCreateUser    = "create_user"
	ActionUpdateUser    = "update_user"
	ActionDeleteUser    = "delete_user"
	ActionChangePassword = "change_password"
	ActionAssignRole    = "assign_role"
	ActionRemoveRole    = "remove_role"
	ActionCreateRole    = "create_role"
	ActionUpdateRole    = "update_role"
	ActionDeleteRole    = "delete_role"
	ActionAssignPermission = "assign_permission"
	ActionRemovePermission = "remove_permission"
)

const (
	ResourceSystemConfig   = "system_config"
	ResourceAlarmRule      = "alarm_rule"
)
