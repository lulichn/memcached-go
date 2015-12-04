package memcache

import (
	"net"
	"bufio"
	"errors"
	"time"
	"hash/crc32"
	"sync"
)

const DefaultTimeout = 100 * time.Millisecond

type Nodes struct {
	addrs []net.Addr
}

type Clients struct {
	serverSelector ServerSelector
}

type conn struct {
	nc   net.Conn
	rw   *bufio.ReadWriter
	addr net.Addr
	c    *Clients
}

type host struct {
	hostName string
	port     int
}

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

func New(servers ...string) *Clients {
	nodes := new(Nodes)
	nodes.SetNodes(servers...)
	return NewFromSelector(nodes)
}

func NewFromSelector(ss ServerSelector) *Clients {
	return &Clients{serverSelector: ss}
}

func (nodes *Nodes) SetNodes(servers ...string) error {
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

func (nodes *Nodes) Servers() []net.Addr {
	return nodes.addrs
}

func (nodes *Nodes) PickServer(key string) (net.Addr, error) {
//	nodes.mu.RLock()
//	defer nodes.mu.RUnlock()
//	if len(nodes.addrs) == 0 {
//		return nil, ErrNoServers
//	}
	if len(nodes.addrs) == 1 {
		return nodes.addrs[0], nil
	}
	bufp := keyBufPool.Get().(*[]byte)
	n := copy(*bufp, key)
	cs := crc32.ChecksumIEEE((*bufp)[:n])
	keyBufPool.Put(bufp)

	return nodes.addrs[cs%uint32(len(nodes.addrs))], nil
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

func (c *Clients) dial(addr net.Addr) (net.Conn, error) {
	type connError struct {
		cn  net.Conn
		err error
	}

	nc, err := net.DialTimeout(addr.Network(), addr.String(), DefaultTimeout)
	if err == nil {
		return nc, nil
	}

	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		return nil, errors.New("Timeout Error") //&ConnectTimeoutError{addr}
	}

	return nil, err
}

func (c *Clients) getConn(addr net.Addr) (*conn, error) {
	nc, err := c.dial(addr)
	if err != nil {
		return nil, err
	}
	cn := &conn{
		nc:   nc,
		addr: addr,
		rw:   bufio.NewReadWriter(bufio.NewReader(nc), bufio.NewWriter(nc)),
		c:    c,
	}
	cn.nc.SetDeadline(time.Now().Add(DefaultTimeout))

	return cn, nil
}
