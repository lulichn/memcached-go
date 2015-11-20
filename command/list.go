package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/lulichn/memcached-go/memcache"
)

func CmdList(c *cli.Context) {
	conn, err := memcache.Conn(c)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer conn.Close()

	items, err := memcache.DumpItems(conn)
	if err != nil {
		fmt.Println(err)
		return
	}

	for idx := 0; idx < len(items); idx += 1 {
		item := items[idx]
		fmt.Println(item.Key)
	}
}
