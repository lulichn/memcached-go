package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/lulichn/memcached-go/memcache"
)

func CmdDelete(c *cli.Context) {
	if len(c.Args()) <= 0 {
		return
	}
	key := []byte(c.Args().Get(0))

	client, err := memcache.Conn(c.GlobalString("host"), c.GlobalInt("port"))
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := client.Delete(key); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Delete Success")
	}
}
