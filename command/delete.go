package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/lulichn/memcached-go/memcache"
)

func CmdDelete(c *cli.Context) {
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

	if err := memcache.Delete(conn, key); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Delete Success")
	}
}