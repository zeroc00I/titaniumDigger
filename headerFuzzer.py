import getopt, sys



# Get full command-line arguments
full_cmd_arguments = sys.argv

# Keep all but the first
argument_list = full_cmd_arguments[1:]

short_options = "hv"
long_options = ["help", "verbose"]

try:
    arguments, values = getopt.getopt(argument_list, short_options, long_options)
except getopt.error as err:
    # Output error, and return with an error code
    print (str(err))
    sys.exit(2)


for current_argument, current_value in arguments:
    if current_argument in ("-v", "--verbose"):
        print ("Enabling verbose mode")
    elif current_argument in ("-h", "--help"):
        print('headerFuzzer [--help] [-d|--domain] [-w|--wordlist]')
        print ('-c: Max curl workers')
        print ('-m: Request Time Out limit')
        print ('-r: Max request header per request')

