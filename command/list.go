package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/lulichn/memcached-go/memcache"
)

func CmdList(c *cli.Context) {
	if c.GlobalBool("es") {
		client, err := memcache.ConnESCluster(c.GlobalString("host"), c.GlobalInt("port"))
		if err != nil {
			fmt.Println(err)
			return
		}
		itemsList, err := client.DumpItems()
		if err != nil {
			fmt.Println(err)
			return
		}
		for idx := 0; idx < len(itemsList); idx += 1 {
			items := itemsList[idx]
			for key, meta := range items {
				fmt.Println(key)
				fmt.Println(meta)
			}
		}
	} else {
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
}
