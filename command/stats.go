package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/lulichn/memcached-go/memcache"
)

func CmdStats(c *cli.Context) {
	client, err := memcache.Conn(c.GlobalString("host"), c.GlobalInt("port"))
	if err != nil {
		fmt.Println(err)
		return
	}

	items, err := client.StatsItems()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Stats Success:")

	for k := range items {
		fmt.Println(k + " : " + items[k])
	}
}
