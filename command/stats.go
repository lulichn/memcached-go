package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/lulichn/memcached-go/memcache"
)

func CmdStats(c *cli.Context) {
	client := &memcache.Client{}
	if c.GlobalBool("ec") {
		client = memcache.NewCluster(c.GlobalStringSlice("host")[0])
	} else {
		client = memcache.New(c.GlobalStringSlice("host"))
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
}
