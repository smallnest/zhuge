package main

import (
	"os"
	"time"

	"github.com/pyroscope-io/client/pyroscope"
	"github.com/smallnest/zhuge/examples/btree"
)

func main() {
	initPyroscope()

	tick := time.NewTicker(10 * time.Second)

	for range tick.C {
		btree.Run(16)
	}
}

func initPyroscope() {
	sc := pyroscope.Config{
		ApplicationName: "zhuge.test_app",
		ServerAddress:   "http://rpcx.io:4040",
		Tags:            map[string]string{"host": "192.168.3.1"},

		// you can disable logging by setting this to nil
		Logger: pyroscope.StandardLogger,

		AuthToken: os.Getenv("PYROSCOPE_AUTH_TOKEN"),

		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
		},
	}
	pyroscope.Start(sc)

}
