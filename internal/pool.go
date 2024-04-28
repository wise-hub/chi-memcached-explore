package internal

import "github.com/bradfitz/gomemcache/memcache"

type MemcachePool struct {
	pool chan *memcache.Client
}

func newMemcachePool(size int, server string) *MemcachePool {
	pool := make(chan *memcache.Client, size)
	for i := 0; i < size; i++ {
		client := memcache.New(server)
		pool <- client
	}
	return &MemcachePool{pool: pool}
}

func (p *MemcachePool) getClient() *memcache.Client {
	return <-p.pool
}

func (p *MemcachePool) releaseClient(client *memcache.Client) {
	select {
	case p.pool <- client:
	default:
		client.Close()
	}
}

var memcachePool *MemcachePool

func init() {
	memcachePool = newMemcachePool(16, "memcache:11211")
}
