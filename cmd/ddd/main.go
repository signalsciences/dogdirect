package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/signalsciences/dogdirect"
	"github.com/signalsciences/dogdirect/hostmetrics"
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
	var hostname string
	flagNS := flag.String("namespace", "", "sets global namespace")
	flagHostname := flag.String("hostname", "", "hostname, if empty use OS")
	flagHostTags := flag.String("hosttags", "", "CSV of host tags")
	flagSystem := flag.Bool("system", false, "emit system host metrics")

	flag.Parse()

	if *flagNS != "" {
		namespace = *flagNS
		if namespace[len(namespace)-1] != '.' {
			namespace += "."
		}
		log.Printf("setting namespace to %q", namespace)
	}

	// set hostname
	if *flagHostname != "" {
		hostname = *flagHostname
	}
	if hostname == "" {
		h, err := os.Hostname()
		if err != nil {
			log.Fatalf("unable to get hostname: %s", err)
		}
		hostname = h
	}
	log.Printf("setting hostname to %q", hostname)

	// todo - set timeout via flag
	api := dogdirect.NewAPI(os.Getenv("DD_API_KEY"), os.Getenv("DD_APP_KEY"), 5*time.Second)

	// create main metrics
	client = dogdirect.New(hostname, api)
	tasks := dogdirect.MultiTask{
		dogdirect.NewPeriodic(client, time.Second*15),
	}
	defer tasks.Close()

	// send system metrics?
	if *flagSystem {
		t, err := hostmetrics.NewFlusher(client, nil)
		if err != nil {
			log.Fatalf("unable create system metrics: %v", err)
		}
		log.Printf("turning on system metrics")
		tasks = append(tasks, dogdirect.NewPeriodic(t, time.Second*15))
	}

	// set host tags?
	if *flagHostTags != "" {
		tags := strings.Split(*flagHostTags, ",")
		for i, t := range tags {
			tags[i] = strings.TrimSpace(t)
		}
		log.Printf("setting host tags to %v", tags)
		t := dogdirect.NewHostTagger(api, hostname, tags)
		tasks = append(tasks, dogdirect.NewPeriodic(t, time.Second*30))
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
