#!/bin/bash

set -e 
formatHeaderListFuzzer(){
	headersWithReflectedValues=$(
	awk '{x=NR+1}(NR<=x){print $0": z3r0c00I"NR }' "$1" )
	echo -e "$headersWithReflectedValues" 
}

buildRequestWithNHeaders(){
	echo -e "$1" |
	awk '{x=NR+1}(NR<=x){print $0}(NR%'$2'==0){print "#"}' |
	xargs -P100 -I@ -d '#' sh -c 'echo "@"'
}

parallelFuzzerWithMaxHeader(){
	echo "$1" |
	xargs -P "$2" sh -c '' 
}

main(){
	#should be recieved by arg parser
	wordList='/opt/SecLists/Discovery/Web-Content/BurpSuite-ParamMiner/lowercase-headers'
	allHeaders=$(formatHeaderListFuzzer "$wordList")
	result=$(buildRequestWithNHeaders "$allHeaders" "10")
	echo -e "$result"
}

main