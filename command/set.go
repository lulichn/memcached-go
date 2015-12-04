package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/lulichn/memcached-go/memcache"
)

func CmdSet(c *cli.Context) {
	if len(c.Args()) <= 1 {
		return
	}
	key   := c.Args().Get(0)
	value := []byte(c.Args().Get(1))

	client := memcache.New(c.GlobalStringSlice("host"))

	if err := client.Set(key, value, 0, 0); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Set Success")
	}
}
