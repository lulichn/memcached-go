package memcache

import (
	"net"
	"bufio"
	"errors"
	"time"
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

func (nodes *Nodes) Servers() []net.Addr {
	return nodes.addrs
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
