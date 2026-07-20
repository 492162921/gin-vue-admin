package inspection

import (
	"context"

	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const initOrderMenu = system.InitOrderExternal + 2

type initMenu struct{}

func init() {
	system.RegisterInit(initOrderMenu, &initMenu{})
}

func (i *initMenu) InitializerName() string { return "inspection_menus" }

func (i *initMenu) MigrateTable(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	return ctx, db.AutoMigrate(&sysModel.SysBaseMenu{}, &sysModel.SysAuthority{})
}

func (i *initMenu) TableCreated(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	return ok && db.Migrator().HasTable(&sysModel.SysBaseMenu{})
}

func (i *initMenu) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	parent := sysModel.SysBaseMenu{
		MenuLevel: 0, Path: "inspection", Name: "inspection", Component: "view/routerHolder.vue", Sort: 9,
		Meta: sysModel.Meta{Title: "K8s 巡检", Icon: "monitor-gva"},
	}
	if err := db.Where("name = ?", parent.Name).FirstOrCreate(&parent).Error; err != nil {
		return ctx, errors.Wrap(err, "初始化巡检父菜单失败")
	}
	children := []sysModel.SysBaseMenu{
		{MenuLevel: 1, ParentId: parent.ID, Path: "cluster", Name: "inspectionCluster", Component: "view/inspection/cluster.vue", Sort: 1, Meta: sysModel.Meta{Title: "集群管理", Icon: "connection"}},
		{MenuLevel: 1, ParentId: parent.ID, Path: "rule", Name: "inspectionRule", Component: "view/inspection/rule.vue", Sort: 2, Meta: sysModel.Meta{Title: "巡检规则", Icon: "set-up"}},
		{MenuLevel: 1, ParentId: parent.ID, Path: "task", Name: "inspectionTask", Component: "view/inspection/task.vue", Sort: 3, Meta: sysModel.Meta{Title: "巡检任务", Icon: "calendar"}},
		{MenuLevel: 1, ParentId: parent.ID, Path: "alert", Name: "inspectionAlert", Component: "view/inspection/alert.vue", Sort: 4, Meta: sysModel.Meta{Title: "巡检告警", Icon: "warning"}},
	}
	for _, menu := range children {
		if err := db.Where("name = ?", menu.Name).FirstOrCreate(&menu).Error; err != nil {
			return ctx, errors.Wrap(err, "初始化巡检子菜单失败")
		}
	}
	var admin sysModel.SysAuthority
	if err := db.Where("authority_id = ?", 888).First(&admin).Error; err != nil {
		return ctx, errors.Wrap(err, "查询超级管理员失败")
	}
	menus := append([]sysModel.SysBaseMenu{parent}, children...)
	if err := db.Model(&admin).Association("SysBaseMenus").Append(menus); err != nil {
		return ctx, errors.Wrap(err, "授予巡检菜单权限失败")
	}
	return ctx, nil
}

func (i *initMenu) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	return db.Where("name = ?", "inspectionAlert").First(&sysModel.SysBaseMenu{}).Error == nil
}
