#!/bin/bash

GREEN=`tput setaf 2`

portByOpenFrequency=`sort -r -k3 /usr/share/nmap/nmap-services | tr '/' ' ' | awk '!/^#/{print $2}'`
allHosts="$1"

#cat $allHosts | xargs -P400 -I@ sh -c 'ip=$(dig +short @); [ ! -z "$ip" ]] && [ ! -z "${ip##*52.*}" ] && [ ! -z "${ip##*54.*}" ] && [ ! -z "${ip##*18.*}" ] && [ ! -z "${ip##10.*}" ] && echo "$ip" | grep -vE "[a-zA-Z]" | anew ipsUp'

echo "Comecando Scan nos IpsUp"

echo "$portByOpenFrequency" | tr ' ' '\n' |
while read line;do 
         xargs -a "$allHosts" -P 100 -I@ sh -c "timeout 1 bash -c '>/dev/tcp/@/$line' 2>/dev/null && echo ${GREEN}[OPEN] @ $line";done
