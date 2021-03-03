import getopt, sys
   
wordList = []


def loadArgumentList(argumentListPath):
    with open(argumentListPath) as argumentList:
            for line in argumentList:
                wordList.append(line)

def main():
    isH = True
    full_cmd_arguments = sys.argv # Get full command-line arguments
    argument_list = full_cmd_arguments[1:] # Keep all but the first
    short_options = "hw:d"
    try:
        arguments, argv = getopt.getopt(argument_list, short_options, ["w="])
    except getopt.error as err: # Output error, and return with an error code
        print (str(err))
        sys.exit(2)
    for current_argument in arguments:
        if ("-h" in current_argument) or ("-help" in current_argument):
            print('headerFuzzer [--help] [-d|--domain] [-w|--wordlist]')
            print ('-m: Request Time Out limit')
            print ('-r: Max request header per request')
            isH = False
    if (isH == True):
        for current_argument, current_value in arguments:
            if ("-w" in current_argument) or ("-wordlist" in current_argument):
                loadPayloadList(current_value)

if __name__ == "__main__":
    main()
