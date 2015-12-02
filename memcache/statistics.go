package memcache

import (
	"bytes"
	"strconv"
	"regexp"
	"fmt"
)

var (
	request_stats          = []byte("stats\r\n")
	request_stats_items    = []byte("stats items\r\n")
	request_stats_slabs    = []byte("stats slabs\r\n")
	request_stats_settings = []byte("stats settings\r\n")
)

func (cli *Client) Stats() (map[string]string, error) {
	return stats(cli, request_stats)
}

func (cli *Client) StatsItems() (map[string]string, error) {
	return stats(cli, request_stats_items)
}

func (cli *Client) StatsSlabs() (map[string]string, error) {
	return stats(cli, request_stats_slabs)
}

func (cli *Client) StatsSettings() (map[string]string, error) {
	return stats(cli, request_stats_settings)
}

func (cli *Client) DumpItems() (map[string]ItemMeta, error) {
	statsItems, err := cli.StatsItems()
	if err != nil {
		return nil, err
	}


	itemSize, err := getItemSize(statsItems)
	if err != nil {
		return nil, err
	}

	items := map[string]ItemMeta{}
	r := regexp.MustCompile(`\[(\d+) b; (\w+) s\]`)

	for bucket, number := range itemSize {
		request := fmt.Sprintf("stats cachedump %d %d\r\n", bucket, number)
		byteArray, err := send_b(cli.conn, []byte(request))
		if err != nil {
			return nil, err
		}

		lines := toMap(byteArray)

		fmt.Println(lines)

		for key, value := range lines {
			sub := r.FindStringSubmatch(value)
			if len(sub) > 0 {
				expire, err := strconv.Atoi(sub[2])
				if err != nil {
					return nil ,err
				}

				items[key] = ItemMeta{Key: key, Size: sub[1], Expire: expire}
			}
		}
	}

	return items, nil
}


func stats(cli *Client, request []byte) (map[string]string, error) {
	if _, err := cli.rw.Write(request); err != nil {
		return nil, err
	}
	if err := cli.rw.Flush(); err != nil {
		return nil, err
	}

	byteArray, err := send_b(cli.conn, request)
	if err != nil {
		return nil, err
	}

	items := toMap(byteArray)

	return items, nil
}

func toMap(message []byte) map[string]string {
	mapVal := map[string]string{}

	subSlice := bytes.Split(message, EOL)
	for idx := 0; idx < len(subSlice); idx += 1 {
		line := subSlice[idx]

		if bytes.Equal(line, response_end) {
			break
		}

		split := bytes.SplitN(line, []byte(" "), 3)
		mapVal[string(split[1])] = string(split[2])
	}

	return mapVal
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

