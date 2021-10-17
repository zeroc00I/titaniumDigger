import requests, sys, optparse, multiprocessing, json
from requests.exceptions import MissingSchema, InvalidURL

manager = multiprocessing.Manager() # https://blog.ruanbekker.com/blog/2019/02/19/sharing-global-variables-in-python-using-multiprocessing/
requests.packages.urllib3.disable_warnings()

global all_results
all_results = manager.list()

def menu():
    parser = optparse.OptionParser()
    parser.add_option('-u', '--url', action="store", dest="url", help='Base target uri (ex. http://target-uri/)')
    parser.add_option('-f', '--file', dest="file", help='example.txt')
    parser.add_option('-t', '--threads', dest="threads", type=int,default=10)
    parser.add_option('-m', '--maxtimeout', dest="timeout", type=int,default=8)
    parser.add_option('-o', '--output', dest="output", type=str,default='cookies.txt')

    options, args = parser.parse_args()
    
    if not options.threads:
        print('[+] Specify a thread number')
        print('[+] Example usage: exploit.py -u http://target-uri/ -t 100')
        exit()
    
    globals().update(locals())

def getPage(url):
    globals().update(locals())
    if not url.startswith('http'):
        url='http://'+url

    urlParts = url.split('/')
    protocol = urlParts[0]
    host = urlParts[2]
    headers = {"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36"}

    try:
        r = requests.get(url,headers=headers,allow_redirects=False, verify=False,timeout=5)
    except requests.exceptions.ConnectionError:
        pass
    except KeyboardInterrupt:
        sys.exit(0)
    except (MissingSchema, InvalidURL):
        pass
    build_wordlist(r,protocol,host)

def build_wordlist(r,protocol,host):
    globals().update(locals())
    if "location" in r.headers:
        cookies = dict(r.cookies)
        url_redirect=protocol+"//"+host+r.headers["location"]

        for key,value in cookies.items():
            output_json = {'url':url,'key':key,'value':value}
            all_results.append(json.dumps(output_json))
            print(json.dumps(output_json))
        getPage(url_redirect)
    else:
        cookies = dict(r.cookies)
        for key,value in cookies.items():
            output_json = {'url':url,'key':key,'value':value}
            all_results.append(json.dumps(output_json))
            print(json.dumps(output_json))


def main():
    menu()
    
    if not sys.stdin.isatty():
        urls = sys.stdin.read()
    else:
        f = open(options.file)
        urls = map(str.strip, f.readlines())

    fire = multiprocessing.Pool(options.threads)
    try:
        fire.map(getPage, urls)
        fire.close()
        fire.join()
    except UnboundLocalError:
        pass
    except KeyboardInterrupt:
        sys.exit(0)

    if options.output:
        print("Escrevendo no arquivo ... %s" % options.output)

        with open(options.output, "w") as f:
            for result in all_results:
                f.write("%s\n" % result)
        f.flush()
if __name__ == "__main__":
        main()
