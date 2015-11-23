package memcache

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	request_get              = []byte("get ")
	request_set              = []byte("set ")
	request_delete           = []byte("delete ")
	request_stats_items      = []byte("stats items")
	request_stats_cache_dump = []byte("stats cachedump ")
)

var (
	response_error     = "ERROR"
	response_end       = "END"
	response_stored    = "STORED"
	response_deleted   = "DELETED"
	response_not_found = "NOT_FOUND"
)

type Client struct {
	conn *net.TCPConn
}

type ItemMeta struct {
	Key    string
	Size   int
	Expire int
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

	return Client{conn: conn}, nil
}

func (cli *Client) Get(key []byte) (string, error) {
	conn := cli.conn

	buff := bytes.NewBuffer(request_get)
	buff.Write(key)
	buff.WriteString("\r\n")
	message := buff.Bytes()

	result := send(conn, message)
	if len(result) == 0 {
		return "", errors.New("Delete Faild. Unkown Error")
	}

	r := regexp.MustCompile(`^VALUE\s+([\w\-]+)\s+(\d+)\s+(\d+)`)
	if !r.MatchString(result[0]) {
		return "", errors.New("Get Faild. Cache miss")
	}

	sub := r.FindStringSubmatch(r.FindString(result[0]))

	return sub[1], nil
}

func (cli *Client) Set(key, value []byte) (string, error) {
	conn := cli.conn

	length := len(value)

	buff := bytes.NewBuffer(request_set)
	buff.Write(key)
	buff.WriteString(" ")
	buff.WriteString("0") // Flag
	buff.WriteString(" ")
	buff.WriteString("0") // Expire
	buff.WriteString(" ")
	buff.WriteString(strconv.Itoa(length)) // Bytes length
	buff.WriteString("\r\n")
	buff.Write(value)
	buff.WriteString("\r\n")
	message := buff.Bytes()

	result := send(conn, message)
	if len(result) == 0 {
		return "", errors.New("Delete Faild. Unkown Error")
	}

	if result[0] != response_stored {
		return "", errors.New("Set Failed. " + result[0])
	}

	return string(key), nil
}

func (cli *Client) Delete(key []byte) error {
	conn := cli.conn

	buff := bytes.NewBuffer(request_delete)
	buff.Write(key)
	buff.WriteString("\r\n")
	message := buff.Bytes()

	result := send(conn, message)
	if len(result) == 0 {
		return errors.New("Delete Faild. Unkown Error")
	}

	if strings.Contains(result[0], response_not_found) {
		return errors.New("Delete Faild. Cache miss")
	}
	return nil
}

func (cli *Client)  DumpItems() ([]ItemMeta, error) {
	conn := cli.conn

	stats := send(conn, []byte(string(request_stats_items)+"\r\n"))
	if strings.Contains(stats[0], response_end) {
		return nil, errors.New("Empty")
	}

	size, err := getItemSize(stats)
	if err != nil {
		return nil, err
	}
	buff := bytes.NewBuffer(request_stats_cache_dump)
	buff.WriteString("1 ")
	buff.WriteString(strconv.Itoa(size))
	buff.WriteString("\r\n")
	message := buff.Bytes()

	lines := send(conn, message)

	r := regexp.MustCompile(`^ITEM ([\w\-]+) \[(\d+) b; (\d+) s\]$`)
	var items []ItemMeta

	for idx := 0; idx < len(lines); idx += 1 {
		line := lines[idx]
		if r.MatchString(line) {
			sub := r.FindStringSubmatch(r.FindString(line))
			size, err := strconv.Atoi(sub[2])
			if err != nil {
				panic(err)
			}
			expire, err := strconv.Atoi(sub[3])
			if err != nil {
				panic(err)
			}

			items = append(items, ItemMeta{Key: sub[1], Size: size, Expire: expire})
		}
	}

	return items, nil
}

func getItemSize(lines []string) (int, error) {
	r := regexp.MustCompile(`^STAT\s*items:(\d+):number\s*(\d+)`)

	for idx := 0; idx < len(lines); idx += 1 {
		if r.MatchString(lines[idx]) {
			sub := r.FindStringSubmatch(r.FindString(lines[idx]))

			return strconv.Atoi(sub[2])
		}
	}
	return 0, errors.New("Unknown")
}

func send(conn *net.TCPConn, message []byte) []string {
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	conn.Write(message)

	readBuff := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	readLength, err := conn.Read(readBuff)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(readBuff[:readLength]))
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return lines
}
