package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/lulichn/memcached-go/memcache"
	"time"
)

func CmdStats(c *cli.Context) {
	config := memcache.ClientConfiguration {
		Timeout: 100 * time.Millisecond,
		HashAlgorithm: memcache.NATIVE_HASH,
	}
	client := memcache.NewWithConfiguration(c.GlobalStringSlice("host"), config)

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
