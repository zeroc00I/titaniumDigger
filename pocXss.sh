function analizingCorrectPathWalk(){

	fixingFromWayBack=$(echo $1 | tr " []'@" "\n" | grep "$rawUrl")
	
	url=$(echo "$fixingFromWayBack" | httpx -follow-redirects -silent -timeout 30)
	[[ -z $url ]] && 
		printf "\033[0;33m[Avoiding timeout host]\e[m" &&
		echo -e "$url\n" &&
		exit
	
	urlCount=$(echo "$url" | awk -F/ '{print NF-2}');
    
    urlHaveQueryString=$(echo $url | grep -c "?.*=")

	if [[ "$urlCount" -gt "1" ]];then
		walkThroughPath "$1"

		if [[ "$urlHaveQueryString" -gt 0 ]]; then

			howMuchQueryString=$(echo "$url" | grep -o '[?|&][a-zA-Z0-9_]\{1,\}=' | wc -l)

			if [[ $howMuchQueryString -eq 1 ]]; then

	    		pocXss "$pass" "1" # SingleQS

	    	else

	    		pocXss "$pass" "$howMuchQueryString" #MultipleQS
	    	fi
		fi
	else

		pocXss "$url"

	fi
}

function walkThroughPath(){

	for i in $(eval echo {1..$urlCount});do 
		
			pass=$(echo "$1" | anew collectedPaths)
			[[ -z "$pass" ]] && 
			printf "\033[0;33m[Avoiding dup scan]\e[m" &&
			echo -e "$url\n" || 
			pocXss "$pass"

	done
}


function pocXss(){

[[ ! -f "payloads.txt" ]] && "payloads.txt file not found! Exiting ..." && exit

payload=$(cat payloads.txt | tr -d "\n")

if [[ "$2" ]]; then

# There is QueryString

	if [[ "$2" -eq "1" ]];then #just one parameter found
		urlQueryString=$(echo "$url" | sed 's#?&00KNX=##g' | qsreplace 'KNX>')

		curl -k -sL "$urlQueryString" | 
		grep -qi 'KNX>' &&
		printf "\033[0;32m[Reflected QS]\e[m" &&
		echo -e "$urlQueryString\n" ||
		printf "\033[0;31m[MISS QS]\e[m" &&
		echo -e "$url\n"

	else
		#more than one
		for i in $(eval echo {1..$2});do

			domain=$(echo $url | unfurl domains)
			payload='KNXQS1>'
			queryStringIndex=$(echo $url | awk -F'&' '{for(i=1;i<=NF;i++) print $i}'| sed -n $i"p" -)

			[[ -z $queryStringIndex ]] && continue

			queryStringIndexReplaced=$(echo "$queryStringIndex" | sed 's#^#?#g' | qsreplace "$payload" | sed 's#^?##g')
			replacedQueryString=$(echo "$url" | sed "s|$queryStringIndex|$queryStringIndexReplaced|g")

			curl -k -sL "$replacedQueryString" | 
			grep -qi 'KNXQS1>' &&
			printf "\033[0;32m[Reflected MQS]\e[m" &&
			echo -e "$replacedQueryString\n" ||
			printf "\033[0;31m[MISS MQS]\e[m" &&
			echo -e "$domain\n"
			
		done
	fi
else

#Normal workflow - without queryString
	req=$(curl -k -sL "$1$payload")
	payloadReflected=$(echo "$req" | grep -oE '[[:digit:]]{1,}KNX>')

	echo "$req" | 
	grep -qi 'KNX>' &&
	printf "\033[0;32m[Reflected NW]\e[m" &&
	echo -e "$payloadReflected\n" ||
	printf "\033[0;31m[MISS]\e[m" &&
	echo -e "$1$payload\n"

fi
}

function main(){
	# Check execution from pipeline
	if [ -p /dev/stdin ]; then
		checkThisUrl=$(</dev/stdin)
	else
		checkThisUrl="$1"
	fi

	rawUrl=$(echo "$checkThisUrl" | unfurl domains)

	analizingCorrectPathWalk "$checkThisUrl"

}

main "$1"

