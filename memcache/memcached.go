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
	EOL                  = []byte("\r\n")
	response_stored      = []byte("STORED")
	response_stored_eol  = []byte("STORED\r\n")
	response_end         = []byte("END")
	response_end_eol     = []byte("END\r\n")

	response_deleted     = []byte("DELETED\r\n")
	response_not_found   = []byte("NOT_FOUND\r\n")

	response_error       = []byte("ERROR\r\n")
)

type ItemMeta struct {
	Key    string
	Size   string
	Expire int
}

type GetItem struct {
	Key   string
	Flags int
	Value []byte
}

func (cli *Client) Get(key string) (GetItem, error) {
	getItem := GetItem{}

	if _, err := fmt.Fprintf(cli.rw, "get %s\r\n", key); err != nil {
		return getItem, err
	}
	if err := cli.rw.Flush(); err != nil {
		return getItem, err
	}

	r := regexp.MustCompile(`^VALUE\s+([\w\-]+)\s+(\d+)\s+(\d+)`)
	meta, err := cli.rw.ReadSlice('\n')
	if err != nil {
		return getItem, err
	}
	if bytes.Equal(meta, response_end_eol) {
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
	readSize, err := cli.rw.Read(buffer)
	if err != nil {
		return getItem, err
	}

	getItem.Key   = metaSub[1]
	getItem.Flags = flags
	getItem.Value = buffer[:readSize]

	return getItem, nil
}

func (cli *Client) Set(key string, value []byte, flags uint16, expireTime int) error {
	length := len(value)
	if _, err := fmt.Fprintf(cli.rw, "set %s %d %d %d\r\n", key, flags, expireTime, length); err != nil {
		return err
	}
	if _, err := cli.rw.Write(value); err != nil {
		return err
	}
	if _, err := cli.rw.Write([]byte("\r\n")); err != nil {
		return err
	}
	if err := cli.rw.Flush(); err != nil {
		return err
	}

	result, err := cli.rw.ReadSlice('\n')
	if err != nil {
		return err
	}

	if bytes.Equal(result, response_stored_eol) {
		return nil
	}
	return errors.New("Set Faild")
}

func (cli *Client) Delete(key string) error {

	if _, err := fmt.Fprintf(cli.rw, "delete %s\r\n", key); err != nil {
		return err
	}
	if err := cli.rw.Flush(); err != nil {
		return err
	}

	meta, err := cli.rw.ReadSlice('\n')
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

func (cli * Client) clusterConfig() ([]host, error) {
	if _, err := fmt.Fprintf(cli.rw, "config get cluster\r\n"); err != nil {
		return nil, err
	}
	if err := cli.rw.Flush(); err != nil {
		return nil, err
	}

	_, err := cli.rw.ReadSlice('\n')
	if err != nil {
		return nil, err
	}

	_, err = cli.rw.ReadSlice('\n')
	if err != nil {
		return nil, err
	}

	info, err := cli.rw.ReadSlice('\n')
	if err != nil {
		return nil, err
	}

	clusterInfo := string(info)
	clusters := strings.Split(clusterInfo, " ")

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

func send_bb(cli *Client, request []byte) (*bytes.Buffer, error) {
	if _, err := cli.rw.Write(request); err != nil {
		return nil, err
	}
	if err := cli.rw.Flush(); err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(make([]byte, 0, 100))
	for {
		readBuff := make([]byte, 1024)
		size, err := cli.rw.Read(readBuff)
		if err != nil {
			return nil, err
		}
		buffer.Write(readBuff[:size])
		if size < 1024 {
			break
		}
	}
	return buffer, nil
}
