package model

import "time"

// SysUploadFile 上传文件记录表。文件上传到存储平台后写入本表。
// UploaderID 可空：匿名/未登录场景上传时为 NULL。表：sys_upload_file
type SysUploadFile struct {
	ID           uint      `gorm:"primaryKey;comment:主键ID" json:"id"`
	CreatedAt    time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time `gorm:"comment:更新时间" json:"updated_at"`
	FileName     string    `gorm:"size:255;not null;comment:原始文件名" json:"file_name"`
	HashName     string    `gorm:"size:128;not null;index;comment:平台哈希文件名" json:"hash_name"`
	Size         int64     `gorm:"not null;default:0;comment:文件大小字节" json:"size"`
	MimeType     string    `gorm:"size:128;default:'';comment:MIME类型" json:"mime_type"`
	Ext          string    `gorm:"size:32;default:'';comment:文件扩展名" json:"ext"`
	Client       string    `gorm:"size:16;default:'';comment:客户端类型" json:"client"`
	RelativePath string    `gorm:"size:512;default:'';comment:文件相对路径" json:"relative_path"`
	AbsolutePath string    `gorm:"size:1024;default:'';comment:上传时完整访问地址" json:"absolute_path"`
	UploaderID   *uint     `gorm:"index;comment:上传人ID" json:"uploader_id"`
}

func (SysUploadFile) TableName() string { return "sys_upload_file" }
