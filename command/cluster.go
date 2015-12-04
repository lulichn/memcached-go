package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/lulichn/memcached-go/memcache"
)

func CmdCluster(c *cli.Context) {
	client := memcache.New(c.GlobalStringSlice("host"))

	if nodes, err := client.ClusterConfig(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Get Cluster list success.")
		for _, node := range nodes {
			fmt.Println(node)
		}
	}
}
