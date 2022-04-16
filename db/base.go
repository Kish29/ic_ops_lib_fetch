package db

import "time"

type BaseDBMod struct {
	Id         uint64    `gorm:"column:id;primary_key" json:"id"`
	UpdateTime time.Time `gorm:"column:create_time"`
	CreateTime time.Time `gorm:"column:update_time"`
}
