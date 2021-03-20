package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
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
var outfile string
var logfile bool

func main() {

	flag.StringVar(&domain, "d", "example.com,google.com", "highest level domain you'd like to filter on (can specify multiple, split on commas)")
	flag.StringVar(&upstream, "u", "127.0.0.1:5353", "Upstream server to send requests to. Requires port!!")
	flag.StringVar(&localbind, "l", "0.0.0.0:53", "Local address to listen on. Defaults to all interfaces on 53.")
	flag.StringVar(&transport, "t", "udp", "Transport to use. Options are the Net value for a DNS Server (udp, udp4, udp6tcp, tcp4, tcp6, tcp-tls, tcp4-tls, tcp6-tls)")
	flag.StringVar(&outfile, "of", "dnsfwd.log", "Path of log file location (defaults to local dir)")
	flag.BoolVar(&logfile, "o", false, "Log output to file (there will probably be a lot of junk here if verbose is turned on)")
	flag.BoolVar(&verbose, "v", false, "enable verbose")
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

	//listen via udp on localhost
	s := dns.Server{Addr: localbind, Net: transport}

	dns.HandleFunc(domain+".", func(w dns.ResponseWriter, r *dns.Msg) { checkQuery(w, r) })
	for {
		if verbose {
			log.Printf("Listening for domains: %v", domainsplits)
			log.Printf("Sending to %s", upstream)
		}
		e := s.ListenAndServe()
		log.Println(e)
		log.Println("Sleeping for 5 seconds before retrying...")
		time.Sleep(time.Second * 5)
	}
}

func checkQuery(w dns.ResponseWriter, r *dns.Msg) {
	for _, x := range r.Question {
		onematch := false
		for _, y := range domainsplits {
			if strings.HasSuffix(x.Name, y+".") {
				onematch = true
				break
			}
		}
		if !onematch {
			if verbose {
				log.Println("Rejected query for", x.Name, "from", w.RemoteAddr().String())
			}
			return
		}
		if verbose {
			log.Println("Query for", x.Name, "from", w.RemoteAddr().String())
		}
	}
	var m *dns.Msg
	m = new(dns.Msg)
	m.SetReply(r)
	m.Compress = false
	m.Authoritative = true

	c := dns.Client{}
	c.UDPSize = 0xffff

	r2, _, err := c.Exchange(m, upstream)
	if err != nil {
		return
	}
	w.WriteMsg(r2)
}
