#!/bin/bash

set -e 

GREEN=`tput setaf 2`
NONCOLOR=`tput sgr0`

formatHeaderListFuzzer(){
	headersWithReflectedValues=$(
	awk '{x=NR+1}(NR<=x){print $0": z3r0c00I"NR }' "$1" 
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
	valueReflected=$(echo -e "$1" | grep -Eo 'z3r0c00I[[:digit:]]{1,}')
	keyReflected=$(echo -e $allHeaders | grep -oE "[a-zA-Z0-9-]{1,}\:.$valueReflected\b")	
	echo -e "$GREEN[Reflected]$NONCOLOR $keyReflected"
}

main(){
	#should be recieved by arg parser
	wordList='/opt/SecLists/Discovery/Web-Content/BurpSuite-ParamMiner/lowercase-headers'
	export -f headerKeyPairToCurlFormat
	export -f scanReflectedHeaders

	allHeaders=$(formatHeaderListFuzzer "$wordList")
	headerKeyPair=$(buildRequestWithNHeaders "$allHeaders" "80" 'https://iam.zerocool.cf/headervul.php' "2")
	parallelFuzzerWithMaxHeader "$headerKeyPair" "2"
}

main