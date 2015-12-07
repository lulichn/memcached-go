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
	key := c.Args().Get(0)

	client := &memcache.Client{}
	if c.GlobalBool("ec") {
		client = memcache.NewCluster(c.GlobalStringSlice("host")[0])
	} else {
		client = memcache.New(c.GlobalStringSlice("host"))
	}

	if err := client.Delete(key); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Delete Success")
	}
}
