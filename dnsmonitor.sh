#!/bin/bash
tcpdump udp port 53 -i eth0 -U -A -w - >> /tmp/dnsMonitor.cap &!;disown
