import requests, re, optparse
from requests.packages.urllib3.exceptions import InsecureRequestWarning
from bs4 import BeautifulSoup

## urllib3.exceptions.MaxRetryError: HTTPSConnectionPool(host='eplt1sourcing.att.com', port=443): Max retries exceeded with url: / (Caused by SSLError(SSLError(1, '[SSL: SSLV3_ALERT_HANDSHAKE_FAILURE] sslv3 alert handshake failure (_ssl.c:1056)')))
## urllib3.exceptions.MaxRetryError: HTTPConnectionPool(host='hvd-intl15.att.com', port=80): Max retries exceeded with url: / (Caused by ConnectTimeoutError(<urllib3.connection.HTTPConnection object at 0x7f545e803710>, 'Connection to hvd-intl15.att.com timed out. (connect timeout=20)'))


requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

def menu():
    parser = optparse.OptionParser()
    parser.add_option('-u', '--url', dest="url", help='Base target uri (ex. http://target-uri/)')
    parser.add_option('-f', '--file', dest="file", help='example.txt')
    parser.add_option('-c', '--concurrents', dest="threads", type=int,default=100)
    parser.add_option('-o', '--output', dest="output", type=str,default='resultsBlind.txt')
    parser.add_option('-v', '--verbose', action="store_true", dest="verbose_errors")
    
    options, args = parser.parse_args()
    if not options.url and not options.file:
        print('[+] Specify an url')
        print('[+] Example usage: exploit.py -u http://target-uri/ -t 10')
        exit()
    
    globals().update(locals())

def get_page():
    if not options.url.startswith('http'):
        options.url='http://'+options.url
    r = requests.get(options.url,headers='',allow_redirects=False, verify=False,timeout=20)
    options.url='http://'+options.url.split('/')[2]+'/'
    soup = BeautifulSoup(r.text, 'html.parser').findAll()
    return soup

def fix_local_file(value):
    if value.startswith('/'):
        value=value[1:]
    if not value.startswith('http'):
        value=options.url+value
    return value

def build(soup):
    queue = []
    valid_keys = ['action','name','src']

    for element in soup:
        for key,value in element.attrs.items():
            if key not in valid_keys:
                continue
            is_form = True if element.name == 'form' else False
            is_js = True if value.endswith('.js') else False
            
            if is_js or is_form:
                value = fix_local_file(value)
            if is_js:
                queue.append({"todo":"tospider","method":"GET","url":value})
            elif is_form:
                method = element.attrs.get("method", "post").lower()
                action = fix_local_file(element.attrs.get("action",options.url).lower())
                inputs = ""
                for input_tag in element.find_all("input"):
                    input_name = input_tag.attrs.get("name")
                    input_type = input_tag.attrs.get("type", "text")
                    input_value = input_tag.attrs.get("value", "zeroc00I")
                    if input_type == 'text':
                        if not inputs:
                            inputs+=input_name+"="+input_value
                        else:
                            inputs+="&"+input_name+"="+input_value
                if method == 'post':
                    queue.append({'todo':'torequest','action':action,'method':'POST','url':value,'data':inputs})
                else:
                    queue.append({'todo':'torequest','action':action,'method':'GET','url':value,'data':inputs})
    return queue

def main():
    menu()
    soup = get_page()
    queue = build(soup)
    if queue:
        print("[Result] {}".format(queue))

if __name__ == "__main__":
    main()
