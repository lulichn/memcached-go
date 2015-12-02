package memcache

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

var EOL = []byte("\r\n")

var (
	request_get              = []byte("get ")
	request_delete           = []byte("delete ")
)

var (
	response_error     = "ERROR"
	response_deleted   = "DELETED"
	response_not_found = "NOT_FOUND"
)

var (
	response_stored      = []byte("STORED")
	response_stored_eol  = []byte("STORED\r\n")
	response_end         = []byte("END")
	response_end_eol     = []byte("END\r\n")
)

type Client struct {
	conn *net.TCPConn
	rw   *bufio.ReadWriter
}

type ItemMeta struct {
	Key    string
	Size   string
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
	bufio.NewReader(conn)
	return Client{
		conn: conn,
		rw:   bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
	}, nil
}

func (cli *Client) Get(key []byte) (string, error) {
	conn := cli.conn

	buff := bytes.NewBuffer(request_get)
	buff.Write(key)
	buff.WriteString("\r\n")
	message := buff.Bytes()

	result := send(conn, message)
	fmt.Println(result)
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

func (cli *Client) Set(key, value []byte, flags uint16, expireTime int) error {
	length := len(value)
	if _, err := fmt.Fprintf(cli.rw, "set %x %d %d %d\r\n", key, flags, expireTime, length); err != nil {
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

func send_b(conn *net.TCPConn, message []byte) ([]byte, error) {
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	conn.Write(message)

	readBuff := make([]byte, 2048)  // FIXME
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	readLength, err := conn.Read(readBuff)
	if err != nil {
		return nil, err
	}

	return readBuff[:readLength], nil
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
