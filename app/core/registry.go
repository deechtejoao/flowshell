package core

import "sync"

var (
	registryMu sync.RWMutex
	registry   = make(map[string]NodeActionMeta)
)

func RegisterNodeAction(tag string, alloc func() NodeAction) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[tag] = NodeActionMeta{Tag: tag, Alloc: alloc}
}

// GetNodeActionMeta retrieves the metadata for a node action tag.
// It replaces the previous hardcoded lookup.
func GetNodeActionMeta(tag string) NodeActionMeta {
	registryMu.RLock()
	defer registryMu.RUnlock()
	if meta, ok := registry[tag]; ok {
		return meta
	}
	panic("unknown node action type: " + tag)
}
