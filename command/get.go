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
	key := []byte(c.Args().Get(0))

	client, err := memcache.Conn(c.GlobalString("host"), c.GlobalInt("port"))
	if err != nil {
		fmt.Println(err)
		return
	}

	if key, err := client.Get(key); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Get Success: " + key)
	}
}
