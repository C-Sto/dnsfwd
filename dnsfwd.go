package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

var version string

var domain string
var upstream string

var domainsplits []string
var verbose bool
var versionflag bool
var localbind string
var transport string
var upstreamtransport string
var outfile string
var logfile bool
var fullquery bool
var timeout int

func main() {

	flag.StringVar(&domain, "d", "", "highest level domain you'd like to filter on (can specify multiple, split on commas)")
	flag.StringVar(&upstream, "u", "127.0.0.1:5353", "Upstream server to send requests to. Requires port!!")
	flag.StringVar(&localbind, "l", "0.0.0.0:53", "Local address to listen on. Defaults to all interfaces on 53.")
	flag.StringVar(&transport, "t", "tcp,udp", "Transport to use. Options are the Net value for a DNS Server (udp, udp4, udp6tcp, tcp4, tcp6, tcp-tls, tcp4-tls, tcp6-tls). Multiple can be supplied - comma separate")
	flag.StringVar(&upstreamtransport, "ut", "udp", "Transport to use for upstream. Defaults to UDP.")
	flag.StringVar(&outfile, "of", "dnsfwd.log", "Path of log file location (defaults to local dir)")
	flag.IntVar(&timeout, "timeout", 2, "default timeout value for read/write/dial")
	flag.BoolVar(&logfile, "o", false, "Log output to file (there will probably be a lot of junk here if verbose, and full queries are turned on)")
	flag.BoolVar(&verbose, "v", false, "enable verbose")
	flag.BoolVar(&fullquery, "full", false, "log full dns queries and responses")
	flag.BoolVar(&versionflag, "version", false, "show version and exit")
	flag.Parse()

	if versionflag {
		if version == "" {
			fmt.Println("dnsfwd UNTAGGED LOCAL BUILD")
			return
		}
		fmt.Println("dnsfwd " + version)
		return
	}
	if logfile {
		f, err := os.OpenFile(outfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("error opening file: %v", err)
		} else {
			defer f.Close()
			w := io.MultiWriter(os.Stdout, f)
			log.SetOutput(w)
		}
	}

	//split up the monitored domains if provided on the cli
	domainsplits = strings.Split(domain, ",")

	transportSplits := strings.Split(transport, ",")

	x := sync.WaitGroup{}
	for _, transp := range transportSplits {

		x.Add(1)
		go startServer(transp)
		time.Sleep(time.Millisecond * 200)
	}

	x.Wait()
}

func startServer(transport string) {
	//listen via udp on localhost
	s := dns.Server{Addr: localbind, Net: transport}
	//handling all
	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) { checkQuery(w, r, transport) })
	for {
		if verbose {
			log.Printf("[%s] Listening for domains: %v", transport, domainsplits)
			log.Printf("[%s] Sending to %s", transport, upstream)
		}
		e := s.ListenAndServe()
		log.Printf("[%s] error: %s", transport, e)
		log.Printf("[%s] Sleeping for 5 seconds before retrying...", transport)
		time.Sleep(time.Second * 5)
	}
}

func checkQuery(w dns.ResponseWriter, r *dns.Msg, transport string) {
	for _, x := range r.Question {
		if len(domainsplits) > 0 {
			onematch := false
			for _, y := range domainsplits {
				if strings.HasSuffix(strings.ToLower(x.Name), strings.ToLower(y+".")) {
					onematch = true
					break
				}
			}
			if !onematch {
				if verbose {
					log.Printf("[%s] Rejected query for %s from %s", transport, x.Name, w.RemoteAddr().String())
				}
				return
			}
		}
		if verbose {
			log.Printf("[%s] Query for %s from %s", transport, x.Name, w.RemoteAddr().String())
		}
	}
	m := new(dns.Msg)
	m.Question = r.Question
	m.Compress = false
	m.Authoritative = true
	c := dns.Client{}
	if timeout != 2 {
		c.Timeout = time.Second * time.Duration(timeout)
	}
	c.Net = upstreamtransport
	c.UDPSize = 0xffff
	r2, _, err := c.Exchange(r, upstream)
	if err != nil {
		if verbose {
			log.Printf("[%s] Error communicating to upstream: %s", transport, err)
		}
		return
	}
	if fullquery {
		log.Printf("[%s] Response:\n%s", transport, r2)
	}
	w.WriteMsg(r2)
}
