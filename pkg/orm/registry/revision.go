package registry

import (
	"context"

	"github.com/wujie1993/waves/pkg/orm/core"
)

type Revisioner interface {
	ListRevisions(ctx context.Context, namespace string, name string) (core.ApiObjectList, error)
	SetRevision(ctx context.Context, obj core.ApiObject) error
	GetLastRevision(ctx context.Context, namespace string, name string) (core.ApiObject, error)
	GetRevision(ctx context.Context, namespace string, name string, revision int) (core.ApiObject, error)
	RevertRevision(ctx context.Context, namespace string, name string, revision int) (core.ApiObject, error)
	DeleteRevision(ctx context.Context, namespace string, name string, revision int) (core.ApiObject, error)
	DeleteAllRevisions(ctx context.Context, namespace string, name string) error
}
