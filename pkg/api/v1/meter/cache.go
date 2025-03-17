package meter

import "sync"

type Cache struct {
	data *Data
	sync.RWMutex
}

func (c *Cache) Get() *Data {
	c.RLock()
	defer c.RUnlock()
	return c.data
}
func (c *Cache) Set(d *Data) {
	c.Lock()
	c.data = d
	c.Unlock()
}
