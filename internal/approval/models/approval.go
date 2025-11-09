package models

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

// -------------------- Request 表 --------------------

// 记录用户提交的 K8S 资源申请

type Request struct {
    ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    RequestID    string    `gorm:"size:64;uniqueIndex;not null" json:"request_id"`
    CreatedBy    string    `gorm:"size:128;not null" json:"created_by"`
    BusinessLine string    `gorm:"size:128;not null" json:"business_line"`
    ServiceName  string    `gorm:"size:128;not null" json:"service_name"`
    Image        string    `gorm:"size:512;not null" json:"image"`
    Replicas     int       `gorm:"not null" json:"replicas"`
    TemplateName string    `gorm:"size:128" json:"template_name"`
    Purpose      string    `gorm:"type:text" json:"purpose"`
    Status       string    `gorm:"size:32;not null" json:"status"`
    Operation    string    `gorm:"default:CREATE" json:"operation"`
    LastUpdated  time.Time `gorm:"autoUpdateTime" json:"last_updated"`
    CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// 在创建时自动生成 RequestID

func (r *Request) BeforeCreate(tx *gorm.DB) (err error) {
    if r.RequestID == "" {
        r.RequestID = uuid.New().String()
    }
    return
}

// -------------------- Approval 表 --------------------
// 记录每一步的审批动作

type Approval struct {
    ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    RequestID    string    `gorm:"column:request_id;index;size:64;not null" json:"request_id"`
    ApproverRole string    `gorm:"column:approver_role;type:varchar(64);not null" json:"approver_role"` // OPS / SRE / K8S
    ApproverName string    `gorm:"column:approver_name;type:varchar(128);not null" json:"approver_name"`
    Decision     string    `gorm:"column:decision;type:varchar(32);not null" json:"decision"` // APPROVE / REJECT
    Comment      string    `gorm:"column:comment;type:text" json:"comment"`
    CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// -------------------- RequestHistory 表 --------------------
// 记录 Request 状态变更历史

type RequestHistory struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    RequestID string    `gorm:"column:request_id;index;size:64;not null" json:"request_id"`
    Status    string    `gorm:"column:status;type:varchar(32);not null" json:"status"`
    ChangedBy string    `gorm:"column:changed_by;type:varchar(128);not null" json:"changed_by"`
    Note      string    `gorm:"column:note;type:text" json:"note"`
    CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

