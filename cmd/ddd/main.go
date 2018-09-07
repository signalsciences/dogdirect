package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/signalsciences/dogdirect"
)

var client *dogdirect.Client

func gauge(args []string) ([]string, error) {
	val, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return nil, err
	}
	client.Gauge(args[0], val)
	return args[2:], nil
}
func count(args []string) ([]string, error) {
	val, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return nil, err
	}
	client.Count(args[0], val)
	return args[2:], nil
}

func incr(args []string) ([]string, error) {
	client.Incr(args[0])
	return args[1:], nil
}

func decr(args []string) ([]string, error) {
	client.Decr(args[0])
	return args[1:], nil
}

func sleep(args []string) ([]string, error) {
	d, err := time.ParseDuration(args[0])
	if err != nil {
		return nil, err
	}
	time.Sleep(d)
	return args[1:], nil
}

func flush(args []string) ([]string, error) {
	err := client.Flush()
	return args, err
}

type cmdfn func([]string) ([]string, error)

var cmdmap = map[string]cmdfn{
	"g":     gauge,
	"gauge": gauge,
	"i":     incr,
	"incr":  incr,
	"d":     decr,
	"decr":  decr,
	"c":     count,
	"count": count,
	"s":     sleep,
	"sleep": sleep,
	"f":     flush,
	"flush": flush,
}

func main() {

	var err error
	name, err := os.Hostname()
	if err != nil {
		log.Fatalf("unable to get hostname: %s", err)
	}
	client, err = ddd.New(name, os.Getenv("DD_API_KEY"))
	if err != nil {
		log.Fatalf("unable to create: %s", err)
	}

	flag.Parse()
	args := flag.Args()

	for len(args) > 0 {
		cmd := args[0]
		args = args[1:]
		fn := cmdmap[cmd]
		if fn == nil {
			log.Fatalf("unknown command: %q", cmd)
		}
		args, err = fn(args)
		if err != nil {
			log.Fatalf("cmd %q failed: %s", cmd, err)
		}
	}
}
