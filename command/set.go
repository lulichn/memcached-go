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
	key := []byte(c.Args().Get(0))
	value := []byte(c.Args().Get(1))

	client, err := memcache.Conn(c.GlobalString("host"), c.GlobalInt("port"))
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := client.Set(key, value, 0, 0); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Set Success")
	}
}
