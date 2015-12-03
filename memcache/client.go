package memcache

import (
	"fmt"
	"net"
	"bufio"
)

type Client struct {
	conn *net.TCPConn
	rw   *bufio.ReadWriter
}

type ClusterClient struct {
	cli   Client
	nodes []Client
}

type host struct {
	hostName string
	port     int
}

func ConnESCluster(configEndpoint string, port int) (ClusterClient, error) {
	clusterClient := ClusterClient{}

	client, err := Conn(configEndpoint, port)
	if err != nil {
		return clusterClient, err
	}

	hosts, err := client.clusterConfig()
	if err != nil {
		return clusterClient, err
	}

	nodes := make([]Client, 0)
	for idx := 0; idx < len(hosts); idx += 1 {
		host := hosts[idx]
		if node, err := Conn(host.hostName, host.port); err != nil {
			return clusterClient, err
		} else {
			nodes = append(nodes, node)
		}
	}

	clusterClient.cli = client
	clusterClient.nodes = nodes

	return clusterClient, nil
}

func Conn(host string, port int) (Client, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	tcpAddress, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return Client{}, err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddress)
	if err != nil {
		return Client{}, err
	}
	bufio.NewReader(conn)
	return Client{
		conn: conn,
		rw:   bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
	}, nil
}