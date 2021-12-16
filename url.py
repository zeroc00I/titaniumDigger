#!/usr/bin/python3

import urllib.parse, requests, optparse, json
from urllib3.exceptions import NewConnectionError
from requests.packages.urllib3.exceptions import InsecureRequestWarning

requests.packages.urllib3.disable_warnings()

def menu():
    parser = optparse.OptionParser()
    parser.add_option('-u', '--url', dest="url", help='Base target uri (ex. http://target-uri/)')
    parser.add_option('-c', '--concurrents', dest="threads", type=int,default=100)
    parser.add_option('-r', '--replace', dest="string_to_replace")
    parser.add_option('-a', '--add', action="store_true", dest="add_string")
    parser.add_option('-f', '--filter', dest="filter_status_code",type=int,default=200)
    
    options, args = parser.parse_args()

    if not options.url:
        print('[+] Specify an url')
        print('[+] Example usage: exploit.py -u "http://target-uri/?query=value" --replace "value_to_replace"')
        exit()
    
    globals().update(locals())

class Requester:
    def __init__(self):
        pass
    
    def get(self,url):
        try:
            self.request = requests.get(url,allow_redirects=True, verify=False)

        except NewConnectionError:
            exit
        except Exception:
            exit
    
    def get_response_line_number(self):
        array_response = self.request.text.split('\n')
        array_response_without_empty_values = [x for x in array_response if x]

        return len(array_response_without_empty_values)

    def get_status_code(self):

        return self.request.status_code

class URL_Builder:
    def __init__(self,url,replace_to):
        if "?" not in url or "=" not in url:
            exit()

        self.url = url
        self.replace_to = urllib.parse.quote_plus(replace_to)
        self.query_string = url.split('?')[1].split('&')
        self.replaces_result = [url]
        self.result = []
    
    def get_querystring_from_url(self):
        for x in self.query_string:
            key = x.split('=')[0]
            value = x.split('=')[1]

            if options.add_string:
                replaced_results = self.url.replace(x,key+'='+value+self.replace_to)
            else:
                replaced_results = self.url.replace(x,key+'='+self.replace_to)

            self.replaces_result.append(replaced_results)
    
    def iterate_all_urls(self):
        for x in self.replaces_result:
            request = Requester()
            request.get(x)

            self.result.append({
                "url":x,
                "status_code":request.get_status_code(),
                "line_numbers":request.get_response_line_number()
            })
                            


def main():
    menu()

    builder = URL_Builder(options.url,options.string_to_replace)
    builder.get_querystring_from_url()
    builder.iterate_all_urls()
    print(json.dumps(builder.result,indent=2))

if __name__ == "__main__":
    main()
