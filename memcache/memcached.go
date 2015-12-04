package memcache

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	request_get    = "get %s\r\n"
	request_set    = "set %s %d %d %d\r\n"
	request_delete = "delete %s\r\n"
	request_config = "config get cluster\r\n"
)

var (
	response_stored    = []byte("STORED\r\n")
	response_end       = []byte("END\r\n")
	response_deleted   = []byte("DELETED\r\n")
	response_not_found = []byte("NOT_FOUND\r\n")
	response_error     = []byte("ERROR\r\n")
)

type ItemMeta struct {
	Key    string
	Size   string
	Expire int
}

type Item struct {
	Key   string
	Flags int
	Expire int
	Value []byte
}

func (cli *Clients) Get(key string) (Item, error) {
	getItem := Item{}

	addr, err := cli.pickServer(key)
	if err != nil {
		return getItem, err
	}
	conn, err := cli.getConn(addr)
	if err != nil {
		return getItem, err
	}

	if _, err := fmt.Fprintf(conn.rw, request_get, key); err != nil {
		return getItem, err
	}
	if err := conn.rw.Flush(); err != nil {
		return getItem, err
	}

	r := regexp.MustCompile(`^VALUE\s+([\w\-]+)\s+(\d+)\s+(\d+)`)
	meta, err := conn.rw.ReadSlice('\n')
	if err != nil {
		return getItem, err
	}
	if bytes.Equal(meta, response_end) {
		return getItem, errors.New("Cache Miss")
	}
	metaSub := r.FindStringSubmatch(string(meta))

	flags, err := strconv.Atoi(metaSub[2])
	if err != nil {
		return getItem, err
	}
	size, err := strconv.Atoi(metaSub[3])
	if err != nil {
		return getItem, err
	}

	buffer := make([]byte, size)
	readSize, err := conn.rw.Read(buffer)
	if err != nil {
		return getItem, err
	}

	getItem.Key   = metaSub[1]
	getItem.Flags = flags
	getItem.Value = buffer[:readSize]

	return getItem, nil
}

func (cli *Clients) Set(key string, value []byte, flags uint16, expireTime int) error {
	addr, err := cli.pickServer(key)
	if err != nil {
		return err
	}
	conn, err := cli.getConn(addr)
	if err != nil {
		return err
	}

	length := len(value)
	if _, err := fmt.Fprintf(conn.rw, request_set, key, flags, expireTime, length); err != nil {
		return err
	}
	if _, err := conn.rw.Write(value); err != nil {
		return err
	}
	if _, err := conn.rw.Write([]byte("\r\n")); err != nil {
		return err
	}
	if err := conn.rw.Flush(); err != nil {
		return err
	}

	result, err := conn.rw.ReadSlice('\n')
	if err != nil {
		return err
	}

	if bytes.Equal(result, response_stored) {
		return nil
	}
	return errors.New("Set Faild")
}

func (cli *Clients) Delete(key string) error {
	addr, err := cli.pickServer(key)
	if err != nil {
		return err
	}
	conn, err := cli.getConn(addr)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(conn.rw, request_delete, key); err != nil {
		return err
	}
	if err := conn.rw.Flush(); err != nil {
		return err
	}

	meta, err := conn.rw.ReadSlice('\n')
	if err != nil {
		return err
	}
	if bytes.Equal(meta, response_deleted) {
		return nil
	}
	if bytes.Equal(meta, response_not_found) {
		return errors.New("Delete failed: Key Not Found. (Key: " + key + ")")
	}

	return errors.New("Delete failed: Unknown")
}

func (cli * Clients) clusterConfig() ([]host, error) {
	addr, err := cli.pickServer("")
	if err != nil {
		return nil, err
	}
	conn, err := cli.getConn(addr)
	if err != nil {
		return nil, err
	}

	if _, err := fmt.Fprintf(conn.rw, request_config); err != nil {
		return nil, err
	}
	if err := conn.rw.Flush(); err != nil {
		return nil, err
	}

	_, err = conn.rw.ReadSlice('\n')
	if err != nil {
		return nil, err
	}

	_, err = conn.rw.ReadSlice('\n')
	if err != nil {
		return nil, err
	}

	info, err := conn.rw.ReadSlice('\n')
	if err != nil {
		return nil, err
	}

	clusters := strings.Split(strings.TrimRight(string(info), "\n"), " ")

	hosts := make([]host, 0)
	for idx := 0; idx < len(clusters); idx += 1 {
		cluster := clusters[idx]
		data := strings.Split(cluster, "|")

		portNum, err := strconv.Atoi(data[2])
		if err != nil {
			return nil, err
		}

		hosts = append( hosts, host {
			hostName: data[0],
			port: portNum,
		})
	}
	return hosts, nil
}
