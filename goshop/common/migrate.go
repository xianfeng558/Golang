package common

import "gorm.io/gorm"

type tableCommenter interface {
	TableComment() string
}

// AutoMigrateCommented 一次迁一张表。
// 模型实现了 TableComment 时，新建表会带上这句表注释。
// 已经存在的表不会被这句改掉，已有表的注释要另执行 ALTER TABLE。
func AutoMigrateCommented(db *gorm.DB, models ...interface{}) error {
	for _, model := range models {
		session := db
		if commented, ok := model.(tableCommenter); ok {
			if comment := commented.TableComment(); comment != "" {
				session = db.Set("gorm:table_options", "COMMENT='"+comment+"'")
			}
		}
		if err := session.AutoMigrate(model); err != nil {
			return err
		}
	}
	return nil
}
