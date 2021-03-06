package memcache

import (
	"net"
	"bufio"
	"time"
	"sync"
)

const DefaultTimeout = 100 * time.Millisecond

type Nodes struct {
	addrs []net.Addr
	lock  sync.RWMutex
}

type ClusterNodes struct {
	cnf   net.Addr
	addrs []net.Addr
	lock  sync.RWMutex
}

type Client struct {
	serverSelector ServerSelector
	configuration  ClientConfiguration
}

type ClientConfiguration struct {
	Timeout       time.Duration
	HashAlgorithm HashAlgorithm
}

type conn struct {
	nc   net.Conn
	rw   *bufio.ReadWriter
	addr net.Addr
}

func New(servers []string) *Client {
	return NewWithConfiguration(servers, ClientConfiguration{})
}

func NewWithConfiguration(servers []string, configuration ClientConfiguration) *Client {
	nodes := new(Nodes)
	nodes.SetNodes(servers)
	return NewFromSelector(nodes, configuration)
}

func NewCluster(cfg string) *Client {
	return NewClusterWithConfiguration(cfg, ClientConfiguration{})
}

func NewClusterWithConfiguration(cfg string, configuration ClientConfiguration) *Client {
	nodes := new(ClusterNodes)
	nodes.SetConfigurationNode(cfg)
	return NewFromSelector(nodes, configuration)
}

func NewFromSelector(ss ServerSelector, configuration ClientConfiguration) *Client {
	return &Client {
		serverSelector: ss,
		configuration: configuration,
	}
}

func (c *Client) pickServer(key string) (net.Addr, error) {
	addr, err := c.serverSelector.PickServer(key, c.configuration.HashAlgorithm)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func (c *Client) getConn(addr net.Addr) (*conn, error) {
	return getConn(addr, c.configuration.getTimeOut())
}

func getConn(addr net.Addr, timeout time.Duration) (*conn, error) {
	nc, err := dial(addr, timeout)
	if err != nil {
		return nil, err
	}
	cn := &conn {
		nc:   nc,
		addr: addr,
		rw:   bufio.NewReadWriter(bufio.NewReader(nc), bufio.NewWriter(nc)),
	}
	cn.nc.SetDeadline(time.Now().Add(timeout))

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
	nodes, err := c.serverSelector.Servers()
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return nil, error_no_available_server
	}

	conn, err := c.getConn(nodes[0])
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func dial(addr net.Addr, timeout time.Duration) (net.Conn, error) {
	type connError struct {
		cn  net.Conn
		err error
	}

	nc, err := net.DialTimeout(addr.Network(), addr.String(), timeout)
	if err == nil {
		return nc, nil
	}

	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		return nil, error_time_out
	}

	return nil, err
}

func (c *ClientConfiguration) getTimeOut() time.Duration {
	if c.Timeout != 0 {
		return c.Timeout
	}
	return DefaultTimeout
}
