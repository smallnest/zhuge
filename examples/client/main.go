package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/smallnest/zhuge"
	"github.com/smallnest/zhuge/examples/btree"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	go initPyroscope()

	tick := time.NewTicker(10 * time.Second)

	for range tick.C {
		btree.Run(16)
	}
}

func initPyroscope() {
	sc := zhuge.Config{
		ApplicationName: "zhuge.test_client",
		ServerAddress:   "http://rpcx.io:4040",
		Tags:            map[string]string{"host": "192.168.3.1"},

		Logger: zhuge.StandardLogger,

		AuthToken: os.Getenv("PYROSCOPE_AUTH_TOKEN"),

		ProfileTypes: []zhuge.ProfileType{
			zhuge.ProfileCPU,
			zhuge.ProfileAllocObjects,
			zhuge.ProfileAllocSpace,
			zhuge.ProfileInuseObjects,
			zhuge.ProfileInuseSpace,
		},
	}
	profiler, err := zhuge.NewSampleProfiler(sc)
	if err != nil {
		panic(err)
	}

	for {
		go profiler.SampleNow()

		d := time.Duration(rand.Intn(10)+10) * time.Second
		time.Sleep(time.Duration(d))
	}

}
