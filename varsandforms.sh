#!/bin/bash
## Usage
## xargs -a firsthttpxhosts -I@ -P20 sh -c 'hakrawler -nocolor -url "@" -plain | ./varsandforms.sh' | anew customScriptResult
##

RED=`tput setaf 1`
GREEN=`tput setaf 2`
YELLOW=`tput setaf 3`
NON=`tput sgr0`
#set -x

function fire(){
payload=$(curl -sfkL "$line"| tr ';' '\n' | grep -Eo 'var.*[^=]*=|name=".*[^"]* ' | tr '="' ' ' | awk '{print $2}' | sort -u| sed -z 's#\n#\=zeroc00I"\&#g' | sed 's#^#/?#g' | sed "s#^#$line#g")
if [ ! -z "$payload" ]; then
 echo "$YELLOW[Requesting]$NON $payload"	
 curl -sfLk "$payload" | grep -q 'zeroc00I"' && echo "$GREEN[Reflected]$NON $payload"
else
 echo "$RED[Failed]$NON $line"
fi
}


if [ -p /dev/stdin ]; then
 checkThisUrl=$(</dev/stdin)
 echo "$checkThisUrl" | tr ' ' '\n' | while read line;do fire "$line";done
else
 echo "Single Thread"
 checkThisUrl="$1"
 fire "$checkThisUrl"
fi
