# DNSFWD

Redirect DNS traffic to an upstream.

Get Latest:

- `wget https://github.com/C-Sto/dnsfwd/releases/latest/download/dnsfwd_linux` (replace linux with darwin or windows.exe for other OS versions)

Example Terraform compatible provisioner section (why is resolved so painful, pls give me a better solution):

```
  provisioner "remote-exec" {
    inline = [
      "sudo systemctl disable systemd-resolved",
      "sudo systemctl stop systemd-resolved",
      "sed -i 's/127.0.0.53/1.1.1.1/g' /etc/resolv.conf",
      "wget https://github.com/C-Sto/dnsfwd/releases/latest/download/dnsfwd_linux",
      "chmod +x dnsfwd_linux",
      "tmux new -d './dnsfwd_linux -v -o -u ${var.upstream} -d ${var.zone}'"
    ]
  }
```

Example:

This will forward all subdomains of example.com, and google.com to a host listening on 1053 at 192.168.0.53. It will not produce verbose output, and will not log to a file (see other options for that)

```
./dnsfwd -d example.com,google.com -u 192.168.0.53:1053
```

```
  -d string
        highest level domain you'd like to filter on (can specify multiple, split on commas)
  -full
        log full dns queries and responses
  -l string
        Local address to listen on. Defaults to all interfaces on 53. (default "0.0.0.0:53")
  -o    Log output to file (there will probably be a lot of junk here if verbose, and full queries are turned on)
  -of string
        Path of log file location (defaults to local dir) (default "dnsfwd.log")
  -t string
        Transport to use. Options are the Net value for a DNS Server (udp, udp4, udp6tcp, tcp4, tcp6, tcp-tls, tcp4-tls, tcp6-tls). Multiple can be supplied - comma separate (default "tcp,udp")
  -timeout int
        default timeout value for read/write/dial (default 2)
  -u string
        Upstream server to send requests to. Requires port!! (default "127.0.0.1:5353")
  -ut string
        Transport to use for upstream. Defaults to UDP. (default "udp")
  -v    enable verbose
  -version
        show version and exit
```
