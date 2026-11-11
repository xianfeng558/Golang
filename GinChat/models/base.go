package models

import (
	"time"

	"gorm.io/gorm"
)

// Base 代替 gorm.Model，作用相同，但创建、更新、软删除三列会带上数据库注释。
// 主键注释由启动时的补注释语句写入。GORM 1.24 不会给主键列同步 comment。
type Base struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time      `gorm:"comment:创建时间"`
	UpdatedAt time.Time      `gorm:"comment:更新时间"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:软删除时间，有值时普通查询看不到"`
}
