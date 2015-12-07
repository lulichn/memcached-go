package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/lulichn/memcached-go/memcache"
)

func CmdCluster(c *cli.Context) {
	client := memcache.New(c.GlobalStringSlice("host"))

	if config, err := client.ClusterConfig(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Get Cluster list success.")
		for _, node := range config.Hosts {
			fmt.Println(node)
		}
	}
}
