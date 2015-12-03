package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/lulichn/memcached-go/memcache"
)

func CmdStats(c *cli.Context) {
	if c.GlobalBool("es") {
		client, err := memcache.ConnESCluster(c.GlobalString("host"), c.GlobalInt("port"))
		if err != nil {
			fmt.Println(err)
			return
		}

		statsList, err := client.Stats()
		if err != nil {
			fmt.Println(err)
			return
		}

		for idx := 0; idx < len(statsList); idx += 1 {
			stats := statsList[idx]

			for k := range stats {
				fmt.Println(k + " : " + stats[k])
			}
		}
	} else {
		client, err := memcache.Conn(c.GlobalString("host"), c.GlobalInt("port"))
		if err != nil {
			fmt.Println(err)
			return
		}

		stats, err := client.Stats()
		if err != nil {
			fmt.Println(err)
			return
		}

		for k := range stats {
			fmt.Println(k + " : " + stats[k])
		}
	}
}
