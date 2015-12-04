package memcache

import (
	"net"
	"sync"
	"fmt"
)


// ServerSelector is the interface that selects a memcache server
// as a function of the item's key.
//
// All ServerSelector implementations must be safe for concurrent use
// by multiple goroutines.
type ServerSelector interface {
	// PickServer returns the server address that a given item
	// should be shared onto.
	Servers() []net.Addr
	PickServer(key string) (net.Addr, error)
	Each(func(net.Addr) error) error
}



// Each iterates over each server calling the given function
func (nodes *Nodes) Each(f func(net.Addr) error) error {
	//	ss.mu.RLock()
	//	defer ss.mu.RUnlock()
	//	for _, a := range ss.addrs {
	//		if err := f(a); nil != err {
	//			return err
	//		}
	//	}
	return nil
}

// keyBufPool returns []byte buffers for use by PickServer's call to
// crc32.ChecksumIEEE to avoid allocations. (but doesn't avoid the
// copies, which at least are bounded in size and small)
//var keyBufPool = sync.Pool{
//	New: func() interface{} {
//		b := make([]byte, 256)
//		return &b
//	},
//}


func (nodes *Nodes) PickServer(key string) (net.Addr, error) {
	//	nodes.mu.RLock()
	//	defer nodes.mu.RUnlock()
	//	if len(nodes.addrs) == 0 {
	//		return nil, ErrNoServers
	//	}
	if len(nodes.addrs) == 1 {
		return nodes.addrs[0], nil
	}
	rv := stringHash(key) % len(nodes.addrs)
	if rv >= 0 || rv < len(nodes.addrs) {
		return nil, fmt.Errorf("Invalid server number. Num: %d, Key: %s", rv, key)
	}

	return nodes.addrs[rv], nil
}

// keyBufPool returns []byte buffers for use by PickServer's call to
// crc32.ChecksumIEEE to avoid allocations. (but doesn't avoid the
// copies, which at least are bounded in size and small)
var keyBufPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 256)
		return &b
	},
}

func (c *Clients) pickServer(key string) (net.Addr, error) {
	//	if !legalKey(key) {
	//		return ErrMalformedKey
	//	}
	addr, err := c.serverSelector.PickServer(key)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func stringHash(key string) int {
	h := 0
	values := []byte(key)
	for _, v := range values {
		h = 31 * h + int(v)
	}
	return h & 0xffffffff
}