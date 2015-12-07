package memcache

import (
	"net"
	"sync"
	"fmt"
	"errors"
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
	PickServer(key string, algorithm HashAlgorithm) (net.Addr, error)
	Each(func(net.Addr) error) error
}

func (nodes *Nodes) SetNodes(servers []string) error {
	addrs := make([]net.Addr, len(servers))

	for i, server := range servers {
		if addr, err := net.ResolveTCPAddr("tcp", server); err != nil {
			return err
		} else {
			addrs[i] = addr
		}
	}

	nodes.addrs = addrs
	return nil
}

func (nodes *Nodes) Servers() []net.Addr {
	nodes.lock.RLock()
	defer nodes.lock.RUnlock()

	return nodes.addrs
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


func (nodes *Nodes) PickServer(key string, algorithm HashAlgorithm) (net.Addr, error) {
	nodes.lock.RLock()
	defer nodes.lock.RUnlock()

	if len(nodes.addrs) == 0 {
		return nil, errors.New("No Server")
	}

	if len(nodes.addrs) == 1 {
		return nodes.addrs[0], nil
	}

	h, err := hash(key, algorithm)
	if err != nil {
		return nil, err
	}
	rv := h % len(nodes.addrs)

	if rv < 0 || rv >= len(nodes.addrs) {
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
