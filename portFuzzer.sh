#!/bin/bash

GREEN=`tput setaf 2`

portByOpenFrequency=`
	sort -r -k3 /usr/share/nmap/nmap-services |
	tr '/' ' ' |
	awk '!/^#/{print $2}'`
allHosts="$1"

cat $allHosts | 
xargs -P400 -I@ sh -c '
	 timeout 1 bash -c "
	  2>/dev/null>/dev/udp/@/1 && echo @ | anew ipsUp
	 "
	 '

echo "$portByOpenFrequency" |
while read line;
	do echo "$line" |
		xargs -a "$allHosts" -P 500 -I@ sh -c "
		         timeout 1 bash -c '
		         >/dev/tcp/@/$line
		         ' 2>/dev/null && echo ${GREEN}[OPEN] @ $line";
	done |
    anew openPorts
