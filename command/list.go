package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/lulichn/memcached-go/memcache"
)

func CmdList(c *cli.Context) {
	client := &memcache.Client{}
	if c.GlobalBool("ec") {
		client = memcache.NewCluster(c.GlobalStringSlice("host")[0])
	} else {
		client = memcache.New(c.GlobalStringSlice("host"))
	}

	if itemsList, err := client.ClusterDumpItems(); err != nil {
		fmt.Println(err)
		return
	} else {
		for idx := 0; idx < len(itemsList); idx += 1 {
			items := itemsList[idx]
			for key, meta := range items {
				fmt.Printf("%s,%s,%d\n", key, meta.Size, meta.Expire)
			}
		}
	}
}
