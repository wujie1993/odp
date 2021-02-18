package operators

import (
	"context"
	"sync"

	"github.com/wujie1993/waves/pkg/orm/v2"
)

// HealthChecks 协程安全的健康检查记录器
type HealthChecks struct {
	items map[string]HealthCheckItem
	mutex sync.RWMutex
}

// HealthCheckItem 记录健康检查项的参数和中断方法
type HealthCheckItem struct {
	v2.LivenessProbe
	Cancel context.CancelFunc
}

// Set 记录健康检查记录，当记录发生覆盖时，返回true
func (hc *HealthChecks) Set(name string, item HealthCheckItem) bool {
	hc.mutex.Lock()
	_, ok := hc.items[name]
	hc.items[name] = item
	hc.mutex.Unlock()
	return ok
}

// Unset 移除健康检查记录
func (hc *HealthChecks) Unset(name string) {
	hc.mutex.Lock()
	delete(hc.items, name)
	hc.mutex.Unlock()
}

// Get 获取健康检查记录
func (hc *HealthChecks) Get(name string) (HealthCheckItem, bool) {
	hc.mutex.RLock()
	item, ok := hc.items[name]
	hc.mutex.RUnlock()
	return item, ok
}

// NewHealthChecks 创建一个新的健康检查记录器
func NewHealthChecks() *HealthChecks {
	return &HealthChecks{
		items: make(map[string]HealthCheckItem),
	}
}
