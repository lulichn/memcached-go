package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/lulichn/memcached-go/memcache"
)

func CmdGet(c *cli.Context) {
	if len(c.Args()) <= 0 {
		return
	}
	key := []byte(c.Args().Get(0))

	conn, err := memcache.Conn(c)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer conn.Close()

	if key, err := memcache.Get(conn, key); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Get Success: " + key)
	}
}
