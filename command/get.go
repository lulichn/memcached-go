package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/lulichn/memcached-go/memcache"
)

func CmdGet(c *cli.Context) {
	if len(c.Args()) <= 0 {
		return
	}
	key := c.Args().Get(0)

	client := memcache.New(c.GlobalStringSlice("host")...)

	if item, err := client.Get(key); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Get Success: " + item.Key)
	}
}
