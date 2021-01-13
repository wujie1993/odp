package operators

import (
	"context"
	"sync"

	"github.com/wujie1993/waves/pkg/orm/v2"
)

type HealthChecks struct {
	items map[string]HealthCheckItem
	mutex sync.RWMutex
}

type HealthCheckItem struct {
	v2.LivenessProbe
	Cancel context.CancelFunc
}

func (hc *HealthChecks) Set(name string, item HealthCheckItem) {
	hc.mutex.Lock()
	hc.items[name] = item
	hc.mutex.Unlock()
}

func (hc *HealthChecks) Unset(name string) {
	hc.mutex.Lock()
	delete(hc.items, name)
	hc.mutex.Unlock()
}

func (hc *HealthChecks) Get(name string) (HealthCheckItem, bool) {
	hc.mutex.RLock()
	item, ok := hc.items[name]
	hc.mutex.RUnlock()
	return item, ok
}

func NewHealthChecks() *HealthChecks {
	return &HealthChecks{
		items: make(map[string]HealthCheckItem),
	}
}
