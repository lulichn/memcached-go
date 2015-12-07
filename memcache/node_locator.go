package memcache

import (
	"net"
	"fmt"
	"errors"
	"bytes"
	"strconv"
)

type ServerSelector interface {
	Servers() ([]net.Addr, error)
	PickServer(key string, algorithm HashAlgorithm) (net.Addr, error)
}

type ECClusterConfig struct {
	Version int
	Hosts   []string
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

	nodes.lock.Lock()
	defer nodes.lock.Unlock()
	nodes.addrs = addrs

	return nil
}

func (nodes *Nodes) Servers() ([]net.Addr, error) {
	nodes.lock.RLock()
	defer nodes.lock.RUnlock()

	return nodes.addrs, nil
}

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

func (nodes *ClusterNodes) SetConfigurationNode(server string) error {
	addr, err := net.ResolveTCPAddr("tcp", server)
	if err != nil {
		return err
	}

	nodes.lock.Lock()
	defer nodes.lock.Unlock()
	nodes.cnf = addr

	return nil
}

func (nodes *ClusterNodes) SetNodes(servers []string) error {
	addrs := make([]net.Addr, len(servers))

	for i, server := range servers {
		if addr, err := net.ResolveTCPAddr("tcp", server); err != nil {
			return err
		} else {
			addrs[i] = addr
		}
	}

	nodes.lock.Lock()
	defer nodes.lock.Unlock()
	nodes.addrs = addrs

	return nil
}

func (nodes *ClusterNodes) Servers() ([]net.Addr, error) {
	conn, err := getConn(nodes.cnf, DefaultTimeout)
	if err != nil {
		return nil, err
	}
	config, err := conn.clusterConfig()
	if err != nil {
		return nil, err
	}
	nodes.SetNodes(config.Hosts)

	nodes.lock.RLock()
	defer nodes.lock.RUnlock()

	return nodes.addrs, nil
}

func (nodes *ClusterNodes) PickServer(key string, algorithm HashAlgorithm) (net.Addr, error) {
	conn, err := getConn(nodes.cnf, DefaultTimeout)
	if err != nil {
		return nil, err
	}
	config, err := conn.clusterConfig()
	if err != nil {
		return nil, err
	}
	nodes.SetNodes(config.Hosts)

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

func (conn *conn) clusterConfig() (ECClusterConfig, error) {
	config := ECClusterConfig{}

	if _, err := fmt.Fprintf(conn.rw, request_config); err != nil {
		return config, err
	}
	if err := conn.rw.Flush(); err != nil {
		return config, err
	}

	versionNum := 0
	hosts := make([]string, 0)
	for idx := 0; ; idx += 1 {
		data, err := conn.rw.ReadSlice('\n')
		if err != nil {
			return config, err
		}
		if bytes.Equal(data, response_error) {
			return config, errors.New("ERROR")
		}
		if bytes.Equal(data, response_end) {
			break
		}
		switch idx {
		case 1:
			if num, err := strconv.Atoi(string(bytes.Trim(data, "\r\n"))); err != nil {
				return config, err
			} else {
				versionNum = num
			}
		case 2:
			nodes := bytes.Split(bytes.Trim(data, "\r\n"), []byte(" "))
			for _, node := range nodes {
				sub := bytes.Split(node, []byte("|"))
				hosts = append(hosts, string(sub[0]) + ":" + string(sub[2]))
			}
		}
	}

	config.Version = versionNum
	config.Hosts = hosts

	return config, nil
}

