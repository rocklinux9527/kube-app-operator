package models

import (
	"encoding/json"
	"time"
)

type Template struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(64);unique;not null" json:"name"`
	Type        string    `gorm:"type:varchar(32);not null" json:"type"`
	Description string    `json:"description"`
	Content     json.RawMessage `gorm:"type:json" json:"content"`
	CreatedAt   time.Time `gorm:"<-:create" json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 可选: 便捷方法——将 JSON 内容转为 map

// ✅ 方便直接读取 JSON 对象

func (t *Template) GetContentMap() (map[string]interface{}, error) {
	var data map[string]interface{}
	if len(t.Content) == 0 {
		return data, nil
	}
	err := json.Unmarshal(t.Content, &data)
	return data, err
}

// ✅ 方便直接设置 JSON 内容

func (t *Template) SetContent(v interface{}) error {
	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	t.Content = bytes
	return nil
}
