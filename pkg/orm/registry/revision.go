package registry

import (
	"context"

	"github.com/wujie1993/waves/pkg/orm/core"
)

// Revisioner 修订版本记录器接口，实现了该接口的对象可被注入到被存储器中，使存储器支持版本记录和回退
type Revisioner interface {
	// 列举目标资源的所有的修订版本
	ListRevisions(ctx context.Context, namespace string, name string) (core.ApiObjectList, error)

	// 为目标资源添加一条修订版本
	SetRevision(ctx context.Context, obj core.ApiObject) error

	// 获取最新的修订版本
	GetLastRevision(ctx context.Context, namespace string, name string) (core.ApiObject, error)

	// 获取指定编号的修订版本
	GetRevision(ctx context.Context, namespace string, name string, revision int) (core.ApiObject, error)

	// 将目标资源回退到指定的修订版本
	RevertRevision(ctx context.Context, namespace string, name string, revision int) (core.ApiObject, error)

	// 删除目标资源的修订版本
	DeleteRevision(ctx context.Context, namespace string, name string, revision int) (core.ApiObject, error)

	// 删除目标资源的所有修订版本
	DeleteAllRevisions(ctx context.Context, namespace string, name string) error
}
