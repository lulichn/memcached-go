package memcache

import (
	"bytes"
	"strconv"
	"regexp"
	"fmt"
	"net"
)

var (
	request_stats          = "stats\r\n"
	request_stats_items    = "stats items\r\n"
	request_stats_slabs    = "stats slabs\r\n"
	request_stats_settings = "stats settings\r\n"
	request_cache_dump     = "stats cachedump %d %d\r\n"
)

type ItemMeta struct {
	Key    string
	Size   string
	Expire int
}

func (cli *Client) Stats() ([]map[string]string, error) {
	return cli.clusterStats(request_stats)
}

func (cli *Client) StatsItems() ([]map[string]string, error) {
	return cli.clusterStats(request_stats_items)
}

func (cli *Client) StatsSlabs() ([]map[string]string, error) {
	return cli.clusterStats(request_stats_slabs)
}

func (cli *Client) StatsSettings() ([]map[string]string, error) {
	return cli.clusterStats(request_stats_settings)
}

func (cli *Client) dumpItems(addr net.Addr) (map[string]ItemMeta, error) {
	conn, err := cli.getConn(addr)
	if err != nil {
		return nil, err
	}

	statsItems, err := cli.stats(addr, request_stats_items)
	if err != nil {
		return nil, err
	}


	itemSize, err := getItemSize(statsItems)
	if err != nil {
		return nil, err
	}

	items := map[string]ItemMeta{}
	r := regexp.MustCompile(`ITEM ([\w\-]+) \[(\d+) b; (\w+) s\]`)

	for bucket, number := range itemSize {
		if _, err := fmt.Fprintf(conn.rw, request_cache_dump, bucket, number); err != nil {
			return nil, err
		}
		if err := conn.rw.Flush(); err != nil {
			return nil, err
		}

		for {
			data, err := conn.rw.ReadSlice('\n')
			if err != nil {
				return nil, err
			}
			if bytes.Equal(data, response_error) {
				return nil, error_response_error
			}
			if bytes.Equal(data, response_end) {
				break
			}
			subStr := r.FindStringSubmatch(string(data))
			if len(subStr) > 0 {
				key  := subStr[1]
				size := subStr[2]
				expire, err := strconv.Atoi(subStr[3])
				if err != nil {
					return nil, err
				}
				items[key] = ItemMeta{Key: key, Size: size, Expire: expire}
			}
		}
	}

	return items, nil
}

func (cli *Client) ClusterDumpItems() ([]map[string]ItemMeta, error) {
	servers, err := cli.serverSelector.Servers()
	if err != nil {
		return nil, err
	}

	mapVal := make([]map[string]ItemMeta, len(servers))

	for i, server := range servers {
		if items, err := cli.dumpItems(server); err != nil {
			return nil, err
		} else {
			mapVal[i] = items
		}
	}

	return mapVal, nil
}


func (cli *Client) stats(addr net.Addr, request string) (map[string]string, error) {
	conn, err := cli.getConn(addr)
	if err != nil {
		return nil, err
	}
	
	if _, err := fmt.Fprintf(conn.rw, request); err != nil {
		return nil, err
	}
	if err := conn.rw.Flush(); err != nil {
		return nil, err
	}

	mapVal := map[string]string{}
	for {
		data, err := conn.rw.ReadSlice('\n')
		if err != nil {
			return nil, err
		}
		if bytes.Equal(data, response_error) {
			return nil, error_response_error
		}
		if bytes.Equal(data, response_end) {
			break
		}

		split := bytes.SplitN(bytes.Trim(data, "\r\n"), []byte(" "), 3)
		mapVal[string(split[1])] = string(split[2])
	}

	return mapVal, nil
}

func (cli *Client) clusterStats(request string) ([]map[string]string, error) {
	servers, err := cli.serverSelector.Servers()
	if err != nil {
		return nil, err
	}

	mapVal := make([]map[string]string, len(servers))

	for i, server := range servers {
		if stat, err := cli.stats(server, request); err != nil {
			return nil, err
		} else {
			mapVal[i] = stat
		}
	}

	return mapVal, nil
}

func getItemSize(items map[string]string) (map[int]int, error) {
	r := regexp.MustCompile(`items:(\d+):number`)

	itemSize := map[int]int{}
	for key, value := range items {
		sub := r.FindStringSubmatch(key)
		if len(sub) > 0 {
			bucket, err := strconv.Atoi(sub[1])
			if err != nil {
				return nil, err
			}
			number, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			itemSize[bucket] = number

		}
	}

	return itemSize, nil
}

