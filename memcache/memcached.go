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

var (
	error_time_out             = errors.New("TimeOot")

	error_no_available_server  = errors.New("Not exist available server")

	error_not_select_algorithm = errors.New("Not selected Hash Algorithm")

	error_get_cache_miss       = errors.New("Get: cache miss")

	error_set_failed           = errors.New("Set: Failed")

	error_delete_failed        = errors.New("Delete: Failed")
	error_delete_key_not_found = errors.New("Delete: Key not found")

	error_response_error       = errors.New("Response error")

	error_cluster_config_response_error = errors.New("Cluster Config: Response error")
)

type Item struct {
	Key    string
	Flags  int
	Expire int
	Value  []byte
}

func (cli *Client) Get(key string) (Item, error) {
	getItem := Item{}

	conn, err := cli.getConnWithKey(key)
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
		return getItem, error_get_cache_miss
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
	conn, err := cli.getConnWithKey(key)
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

	return error_set_failed
}

func (cli *Client) Delete(key string) error {
	conn, err := cli.getConnWithKey(key)
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
		return error_delete_key_not_found
	}

	return error_delete_failed
}
