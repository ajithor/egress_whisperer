# egress whisperer
> Find your way out before you commit to a C2 transport.

A lightweight egress diagnostic binary, written in golang. Drop the binary on a compromised host and point it at your listener. Produces a report on which ports and protocols survided the target's firewall (including Deep Packet Inspection). The report can then be used structure the fallback into your implant.
Tested against pfSense and Suricata [pfSense_suricata_rules.md]

# Problem Statement
You gain Initial access on a hardened internal network. Before you can structure your C2 fallback and establish a channel, you need to know what gets out from the firewall of the network (maybe ACL port blocks, or maybe DPI which kills anything that doesnt look like legitimate browzer traffic). 

Standard approach would be to mannually test with something like `netcat`, or find a way to enumerate the rules, or some logs. Alternatively, blindly structure the fallback channel would risk detection. 

egress_whisperer solves this, by crafting legitamete-looking requests, with optional features like jitter stealth, on common open ports (currently supports HTTPS/TLS over 443 and 8443, HTTP over 80 and 8080, DNS over 53 and NTP over 123). Then produces a report, that can be used to structure your fallback for the C2 implant.

---
# Features
- Concurrent probes -- all ports and protocols tested simultaneously, one go routine per probe. (Default, or -stealth 0)
- Jittering -- based on a stealth level, 0-4, 0 being loud and 4 being paranoid, probes are sent with varying levels of base delay and jitter variance.
- Protocol mimicry -- HTTP probes send hand-crafter request with real Chrome headers, over `net.DialTimeout`, which gives it full control. Same with DNS and NTP to send legitimate-looking headers. HTTPS/TLS mimicry is carried out using `utls`
- DPI-aware -- tested agains Suricata IPS rules inspecting payload content.
- Single Static binary -- no dependencies, no installation. Cross-compiled from attacker machine and dropped on target
- Report - output showing what ports/protocols went through, and what dint, with reasoning.

---
# Usage
If you just wanna use the tool --
Short version
```sh
GOOS=linux GOARCH=amd64 go build -o server server.go
GOOS=linux GOARCH=amd64 go build -o client client.go
./server [-debug on/off]
./client -target 10.0.0.1 [-stealth 2] [-timeout 3s] [-ports 80, 443, 53]
```
Longer
```sh
#on the attacker machine
#first, generate your 
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -sha256 -days 365 -nodes -subj "/CN=egress_whisperer.local"
#install the utls
go get github.com/refraction-networking/utls

#build the server
GOOS=linux GOARCH=amd64 go build -o server server.go
#cross-compile or build the client, 
#for liux
GOOS=linux GOARCH=amd64 go build -o client client.go
#for windows
GOOS=windows GOARCH=amd64 go build -o client.exe client.go
#or look up how to build your arch and os

#now, transfer the `client` to the victim machine
#start the server on attacker
./server [-debug on/off]
./client -target 10.0.0.1 [-stealth 2] [-timeout 3s] [-ports 80, 443, 53]
#OR
client.exe -target 10.0.0.1 [-stealth 2] [-timeout 3s] [-ports 80, 443, 53]
#this should give you a report
#Use the report to structure your C2 implant's fallback
```
If you're a someone who wants to work on this tool, or want to establish a baseline, and understand which part does what, check out the [test_reproduction.md]

---
Sample Report output
```txt
--- Rport ---
http   80    : SUCCESS
http   8080  : SUCCESS
tls    443   : SUCCESS
udp    53    : SUCCESS
udp    123   : SUCCESS
tcp    9001  : BLOCKED
```

---
References-
https://go.dev/tour/
https://www.youtube.com/watch?v=_FU6QC3q1mk&pp=ygUXZ28gbmV0IHBhY2thZ2UgdHV0b3JpYWw%3D
https://www.youtube.com/watch?v=uQBpJzAyn7Y
https://www.youtube.com/watch?v=y2jP45S9BHk&t=649s&pp=ygUQZ29sYW5nIHdhaXRncm91cNIHCQmjCwGHKiGM7w%3D%3D
https://www.youtube.com/watch?v=Y-Dj8lHmXy8
https://medium.com/learning-the-go-programming-language/lets-make-an-ntp-client-in-go-287c4b9a969f
https://medium.com/@openmohan/dns-basics-and-building-simple-dns-server-in-go-6cb8e1cfe461
https://github.com/refraction-networking/utls
https://medium.com/@ravikumarray92/concurrency-in-go-goroutines-channels-waitgroup-db6a50b87a04