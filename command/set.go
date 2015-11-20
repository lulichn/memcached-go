package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/lulichn/memcached-go/memcache"
)

func CmdSet(c *cli.Context) {
	if len(c.Args()) <= 1 {
		return
	}
	key := []byte(c.Args().Get(0))
	value := []byte(c.Args().Get(1))

	conn, err := memcache.Conn(c)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer conn.Close()

	if key, err := memcache.Set(conn, key, value); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Set Success: " + key)
	}
}
