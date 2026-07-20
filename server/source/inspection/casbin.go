package inspection

import (
	"context"

	adapter "github.com/casbin/gorm-adapter/v3"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const initOrderCasbin = system.InitOrderExternal + 4

type initCasbin struct{}

func init() {
	system.RegisterInit(initOrderCasbin, &initCasbin{})
}

func (i *initCasbin) InitializerName() string { return "inspection_casbin" }

func (i *initCasbin) MigrateTable(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	return ctx, db.AutoMigrate(&adapter.CasbinRule{})
}

func (i *initCasbin) TableCreated(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	return ok && db.Migrator().HasTable(&adapter.CasbinRule{})
}

func (i *initCasbin) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	for _, api := range inspectionAPIs() {
		rule := adapter.CasbinRule{Ptype: "p", V0: "888", V1: api.Path, V2: api.Method}
		if err := db.Where(adapter.CasbinRule{Ptype: rule.Ptype, V0: rule.V0, V1: rule.V1, V2: rule.V2}).FirstOrCreate(&rule).Error; err != nil {
			return ctx, errors.Wrap(err, "初始化巡检 Casbin 权限失败")
		}
	}
	return ctx, nil
}

func (i *initCasbin) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	var count int64
	return db.Model(&adapter.CasbinRule{}).Where("ptype = ? AND v0 = ? AND v1 LIKE ?", "p", "888", "/inspection/%").Count(&count).Error == nil && count == int64(len(inspectionAPIs()))
}
