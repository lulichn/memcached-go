package memcache

import (
	"net"
	"bufio"
	"errors"
	"time"
	"sync"
)

const DefaultTimeout = 100 * time.Millisecond

type Nodes struct {
	addrs []net.Addr
	lock  sync.RWMutex
}

type ClientConfiguration struct {
	Timeout       time.Duration
	HashAlgorithm HashAlgorithm
}

type Client struct {
	serverSelector ServerSelector
	configuration  ClientConfiguration
}

type conn struct {
	nc   net.Conn
	rw   *bufio.ReadWriter
	addr net.Addr
	c    *Client
}


func New(servers []string) *Client {
	return NewWithConfiguration(servers, ClientConfiguration{})
}

func NewWithConfiguration(servers []string, configuration ClientConfiguration) *Client {
	nodes := new(Nodes)
	nodes.SetNodes(servers)
	return NewFromSelector(nodes, configuration)
}

func NewFromSelector(ss ServerSelector, configuration ClientConfiguration) *Client {
	return &Client {
		serverSelector: ss,
		configuration: configuration,
	}
}

func (c *Client) pickServer(key string) (net.Addr, error) {
	addr, err := c.serverSelector.PickServer(key, c.getHashAlgorithm())
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func (c *Client) dial(addr net.Addr) (net.Conn, error) {
	type connError struct {
		cn  net.Conn
		err error
	}

	nc, err := net.DialTimeout(addr.Network(), addr.String(), c.getTimeOut())
	if err == nil {
		return nc, nil
	}

	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		return nil, errors.New("Timeout Error")
	}

	return nil, err
}

func (c *Client) getConn(addr net.Addr) (*conn, error) {
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
	cn.nc.SetDeadline(time.Now().Add(c.getTimeOut()))

	return cn, nil
}

func (c *Client) getConnWithKey(key string) (*conn, error) {
	addr, err := c.pickServer(key)
	if err != nil {
		return nil, err
	}

	conn, err := c.getConn(addr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *Client) getConnConfigNode() (*conn, error) {
	nodes := c.serverSelector.Servers()
	if len(nodes) == 0 {
		return nil, errors.New("No Server")
	}

	conn, err := c.getConn(nodes[0])
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *Client) getTimeOut() time.Duration {
	if c.configuration.Timeout != 0 {
		return c.configuration.Timeout
	}
	return DefaultTimeout
}

func (c *Client) getHashAlgorithm() HashAlgorithm {
	return c.configuration.HashAlgorithm
}
