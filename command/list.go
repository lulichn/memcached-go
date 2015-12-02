package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/lulichn/memcached-go/memcache"
)

func CmdList(c *cli.Context) {
	client, err := memcache.Conn(c.GlobalString("host"), c.GlobalInt("port"))
	if err != nil {
		fmt.Println(err)
		return
	}

	items, err := client.DumpItems()
	if err != nil {
		fmt.Println(err)
		return
	}

	for key, meta := range items {
		fmt.Println(key)
		fmt.Println(meta)

	}
}
