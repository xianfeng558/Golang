// Package data 启动时连接计价库，并写入三行可重复执行的示例规则。
package data

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"valuation/internal/biz"
	"valuation/internal/conf"
)

// ProviderSet 是 data 层交给 wire 的构造函数。
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewPriceRuleInterface)

// Data 持有计价服务的 MySQL。
type Data struct {
	Mdb *gorm.DB
}

// NewData 打开 MySQL，建表 price_rules，并写入城市 1 的三段示例规则。
// 三行的主键固定是 1、2、3。每次启动用 OnConflict 按主键覆盖，所以改了示例数字后重新运行就会生效，不会越插越多。
// 金额单位是分。7 点到 23 点前起步 300，23 点到 24 点前起步 350，0 点到 7 点前起步 400。里程费和时长费都是 35 和 10。
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	db, err := openMySQL(c.GetDatabase().GetSource())
	if err != nil {
		return nil, nil, err
	}
	if err := autoMigrate(db, &biz.PriceRule{}); err != nil {
		return nil, nil, fmt.Errorf("migrate price_rule: %w", err)
	}
	rules := []biz.PriceRule{
		{ID: 1, PriceRuleWork: biz.PriceRuleWork{CityID: 1, StartFee: 300, DistanceFee: 35, DurationFee: 10, StartAt: 7, EndAt: 23}},
		{ID: 2, PriceRuleWork: biz.PriceRuleWork{CityID: 1, StartFee: 350, DistanceFee: 35, DurationFee: 10, StartAt: 23, EndAt: 24}},
		{ID: 3, PriceRuleWork: biz.PriceRuleWork{CityID: 1, StartFee: 400, DistanceFee: 35, DurationFee: 10, StartAt: 0, EndAt: 7}},
	}
	if err := db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&rules).Error; err != nil {
		return nil, nil, fmt.Errorf("seed price_rule: %w", err)
	}
	sqlDB, _ := db.DB()
	cleanup := func() {
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{Mdb: db}, cleanup, nil
}

func autoMigrate(db *gorm.DB, models ...interface{}) error {
	for _, model := range models {
		session := db
		if commented, ok := model.(interface{ TableComment() string }); ok {
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

// openMySQL 先建库 laomadj_valuation，再打开它。
// 连接最多 10 条、空闲 2 条、寿命 30 秒，用来避开本机 MySQL 掐掉空闲连接后出现的 invalid connection。
func openMySQL(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("mysql dsn is empty")
	}
	cfg, err := mysqldriver.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	name := cfg.DBName
	if name == "" || strings.ContainsAny(name, "`;'\"") {
		return nil, fmt.Errorf("invalid mysql database name")
	}
	cfg.DBName = ""
	admin, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	defer admin.Close()
	if err := admin.Ping(); err != nil {
		return nil, err
	}
	if _, err := admin.Exec("CREATE DATABASE IF NOT EXISTS `" + name + "` DEFAULT CHARACTER SET utf8mb4"); err != nil {
		return nil, err
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(30 * time.Second)
	return db, nil
}
