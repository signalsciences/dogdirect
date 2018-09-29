package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/signalsciences/dogdirect"
)

var (
	client    *dogdirect.Client
	namespace string
)

func gauge(args []string) ([]string, error) {
	name := namespace + args[0]
	log.Printf("gauge %s", name)
	val, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return nil, err
	}
	client.Gauge(name, val, nil)
	return args[2:], nil
}
func count(args []string) ([]string, error) {
	name := namespace + args[0]
	log.Printf("count %s", name)
	val, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return nil, err
	}
	client.Count(name, val, nil)
	return args[2:], nil
}

func incr(args []string) ([]string, error) {
	name := namespace + args[0]
	log.Printf("incr %s", name)
	client.Incr(name, nil)
	return args[1:], nil
}

func decr(args []string) ([]string, error) {
	name := namespace + args[0]
	log.Printf("decr %s", name)
	client.Decr(name, nil)
	return args[1:], nil
}

func sleep(args []string) ([]string, error) {
	log.Printf("sleep %s", args[0])
	d, err := time.ParseDuration(args[0])
	if err != nil {
		return nil, err
	}
	time.Sleep(d)
	return args[1:], nil
}

func flush(args []string) ([]string, error) {
	log.Printf("flush")
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
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("unable to get hostname: %s", err)
	}
	api := dogdirect.NewAPI(os.Getenv("DD_API_KEY"), os.Getenv("DD_APP_KEY"), 5*time.Second)

	client = dogdirect.New(hostname, api)
	defer client.Close()

	flagNS := flag.String("namespace", "", "sets global namespace")
	flagTags := flag.String("tags", "", "CSV of tags applied to every metric")
	flag.Parse()

	if *flagTags != "" {
		tags := strings.Split(*flagTags, ",")
		for i, t := range tags {
			tags[i] = strings.TrimSpace(t)
		}
		log.Printf("setting tags to %v", tags)
		if err := api.AddHostTags(hostname, "", tags); err != nil {
			log.Fatalf("Unable to set tags: %s", err)
		}

	}
	if *flagNS != "" {
		namespace = *flagNS
		if namespace[len(namespace)-1] != '.' {
			namespace += "."
		}
		log.Printf("setting namespace to %q", namespace)
	}

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
