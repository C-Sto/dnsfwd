# DNSFWD

Redirect DNS traffic to an upstream.

Example:

This will forward all subdomains of example.com, and google.com to a host listening on 1053 at 192.168.0.53. It will not produce verbose output, and will not log to a file (see other options for that)

```
./dnsfwd -d example.com,google.com -u 192.168.0.53:1053
```

```
  -d string
        highest level domain you'd like to filter on (can specify multiple, split on commas) (default "example.com,google.com")
  -l string
        Local address to listen on. Defaults to all interfaces on 53. (default "0.0.0.0:53")
  -o    Log output to file (there will probably be a lot of junk here if verbose is turned on)
  -of string
        path of log file location (defaults to local dir) (default "dnsfwd.log")
  -t string
        Transport to use. Options are the Net value for a DNS Server (udp, udp4, udp6tcp, tcp4, tcp6, tcp-tls, tcp4-tls, tcp6-tls) (default "udp")
  -u string
        Upstream server to send requests to. Requires port!! (default "127.0.0.1:5353")
  -v    enable verbose
```
