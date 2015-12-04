package memcache

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
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

func (cli *Client) Get(key string) (Item, error) {
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

func (cli *Client) Set(key string, value []byte, flags uint16, expireTime int) error {
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

func (cli *Client) Delete(key string) error {
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

func (cli * Client) ClusterConfig() ([]string, error) {
	addrs := cli.serverSelector.Servers()
	conn, err := cli.getConn(addrs[0])
	if err != nil {
		return nil, err
	}

	if _, err := fmt.Fprintf(conn.rw, request_config); err != nil {
		return nil, err
	}
	if err := conn.rw.Flush(); err != nil {
		return nil, err
	}

	versionNum := 0
	hosts := make([]string, 0)
	for idx := 0; ; idx += 1 {
		data, err := conn.rw.ReadSlice('\n')
		if err != nil {
			return nil, err
		}
		if bytes.Equal(data, response_error) {
			return nil, errors.New("ERROR")
		}
		if bytes.Equal(data, response_end) {
			break
		}
		switch idx {
		case 1:
			if num, err := strconv.Atoi(string(data)); err != nil {
				return nil, err
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
	fmt.Println(versionNum)
	fmt.Println(hosts)
	return hosts, nil
}
