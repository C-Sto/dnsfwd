package main

import (
	"flag"
	"log"
	"strings"

	"github.com/miekg/dns"
)

var domain string
var upstream string

var domainsplits []string
var verbose bool

func main() {

	flag.StringVar(&domain, "d", "example.com,google.com", "highest level domain you'd like to filter on (can specify multiple, split on commas)")
	flag.StringVar(&upstream, "u", "127.0.0.1:5353", "Upstream server to send requests to. Requires port!!")
	flag.BoolVar(&verbose, "v", false, "enable verbose")
	flag.Parse()

	//split up the monitored domains if provided on the cli
	domainsplits = strings.Split(domain, ",")

	//listen via udp on localhost
	s := dns.Server{Addr: "0.0.0.0:53", Net: "udp"}

	dns.HandleFunc(domain+".", func(w dns.ResponseWriter, r *dns.Msg) { checkQuery(w, r) })
	for {
		if verbose {
			log.Printf("Listening for domains: %v\nSending to %s", domainsplits, upstream)
		}
		e := s.ListenAndServe()
		log.Println(e)
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
