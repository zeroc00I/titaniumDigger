#!/bin/bash

set -e 

RED=`tput setaf 1`
GREEN=`tput setaf 2`
YELLOW=`tput setaf 3`
NONCOLOR=`tput sgr0`

formatHeaderListFuzzer(){
	headersWithReflectedValues=$(
		awk '{x=NR+1}(NR<=x){print $0": '$prefixPayload'?"NR }' "$1" 
	)
	echo -e "$headersWithReflectedValues"
}

buildRequestWithNHeaders(){
	# $1 all keyPairs
	# $2 max n headers / req
	# $3 target domain
	# $4 req timeout

	echo -e "$1" |
	awk '{x=NR+1}(NR<=x){print $0}(NR%'$2'==0){print "#"}' |
	xargs -P100 -I@ -d '#' bash -c 'headerKeyPairToCurlFormat "@" "'$3'" '$4
}

headerKeyPairToCurlFormat(){
	# $1 all keyPairs
	# $2 target domain
	# $3 request timeout

	echo -e "$1" | 
	sed 's/^/-H "/g;s/$/"/g' | 
	tr '\n' ' ' | 
	sed 's#^#\ncurl "'$2'" -sf -m "'$3'" #g;s/-H ""//g'
}

parallelFuzzerWithMaxHeader(){
	fuzzResult=$(
		echo -e "$1" |
		xargs -0 -P "$2" -I@ sh -c '@'
	)
	scanReflectedHeaders "$fuzzResult"
}

scanReflectedHeaders(){
	valueReflected=$(echo -e "$1" | grep -Eo "$prefixPayload\?[[:digit:]]{1,}\b")
	keyReflected=$(echo -e "$allHeaders" | grep "$valueReflected\b")	
	echo -e "$GREEN[Reflected]$NONCOLOR $keyReflected"
}

checkEmptyArgs(){
	# default Values
	if [ -z "$prefixPayload" ]; then
		prefixPayload='z3r0c00I'
	fi

	if [ -z "$wordList" ]; then
		wordList='/opt/SecLists/Discovery/Web-Content/BurpSuite-ParamMiner/lowercase-headers'
		echo -e "$YELLOW[CONFIG]$NONCOLOR Tool will try to use default wordlist $wordList (check if it exists)"
	#else
		# will be enable next version with --verbose
		#echo -e "$GREEN[CONFIG]$NONCOLOR Using wordlist $YELLOW  $wordList $NONCOLOR"

	fi
	if [ -z "$domain" ]; then 
		echo -e "$RED[CONFIG]$NONCOLOR Please, set some target $YELLOW domain $NONCOLOR with -d(omain)"
		usage
		exit
	fi
	if [ -z "$maxWorkers" ]; then
		maxWorkers=100
	fi
	if [ -z "$maxTimeout" ]; then
		maxTimeout=10
	fi
	if [ -z "$maxRequestHeader" ]; then
		maxRequestHeader=90
	fi
}

usage(){
	echo -e "\n$YELLOW Usage:$NONCOLOR\n headerFuzzer [--help] [-d|--domain] [-w|--wordlist]\n"
	echo -e "$YELLOW Optional:$NONCOLOR\n -c: Max curl workers"
	echo -e " -m: Request Time Out limit"
	echo -e " -r: Max request header per request"
}

main(){
	checkEmptyArgs

	export -f headerKeyPairToCurlFormat
	export -f scanReflectedHeaders

	allHeaders=$(
		formatHeaderListFuzzer "$wordList"
	)

	headerKeyPair=$(
		buildRequestWithNHeaders "$allHeaders" "$maxRequestHeader" "$domain" "$maxTimeout"
	)

	parallelFuzzerWithMaxHeader "$headerKeyPair" "$maxWorkers"
}

while getopts ":r:c:w:d:m:p:" OPTION; do
	case $OPTION in
	    c) export maxWorkers=$OPTARG;;
		w) export wordList=$OPTARG;;
		d) export domain=$OPTARG;; 
		m) export maxTimeout=$OPTARG;; 
		r) export maxRequestHeader=$OPTARG;;
		p) export prefixPayload=$OPTARG;;
	    h | *) # Display help.
				usage
				exit
			;;
	esac
done

main