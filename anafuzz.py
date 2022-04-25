#!/usr/bin/python3.9
import requests, optparse, os
from requests.packages.urllib3.exceptions import InsecureRequestWarning
from urllib.parse import urlparse

requests.packages.urllib3.disable_warnings()

def menu():
    parser = optparse.OptionParser()
    parser.add_option('-u', '--url', dest="url_to_fuzz", help='https://example.zerocool.com/item?search=waf')
    parser.add_option('-w', '--wordlist', dest="raw_wordlist_without_extension", help='Wordlist file without extension')
    parser.add_option('-t', '--threads', dest="threads", default=20, type=int, help='options to set threads numbers to fuzz')
    parser.add_option('-v', '--verbose', dest="verbose_mode_status",default=False, help='should verbose mode be enabled?')

    options, args = parser.parse_args()

    if not options.url_to_fuzz:
        print("[+] Please enter some URL with -u parameter (Check -h for help)")
        exit()

    globals().update(locals())

class Requester:
    def __init__(self,url):
        self.url = url
        self.get()
    
    def get(self):
        try:

            s = requests.Session()
            req = requests.Request(
                method='GET',
                url=self.url,
                headers={
                "Accept": "text/html",
                "User-Agent":"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.60 Safari/537.36"
                }
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

    def get_final_url(self):
      return self.request.url

    def get_web_server(self):
      if 'server' in self.request.headers.keys():
        if '/' in self.request.headers['Server']:
          self.webserver = self.request.headers['Server'].split('/')[0]
          print('teste')
        else:
          self.webserver = self.request.headers['Server']

    def get_page_tech_detection(self):
      page_url_filename = os.path.basename(
        urlparse(
          self.get_final_url()
          ).path
        )
      
      page_extension = page_url_filename.split('.')[1] if '.' in page_url_filename else None
      self.page_technology = page_extension 

      return page_extension  

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

class fuzzer:
  def __init__(self,url):
    self.url = url
    self.baseline()

  def baseline(self):
    self.baseline = Requester(self.url)
    self.baseline_arquivo = Requester(self.url+"/xhopefullydoesntexistx.html")
    self.baseline_diretorio = Requester(self.url+"/xhopefullydoesntexistx/")

  def wordlist_based_on_webserver(self):
    print(type(self.webserver))

  def tech_detection(self):
    main_page_technology =  self.baseline.get_page_tech_detection()
    
    if main_page_technology is None:
      self.webserver = self.baseline.get_web_server() if self.baseline.get_web_server() else None
      print(self.webserver)
      self.wordlist_based_on_webserver()
    
    else:
      print("Technology used is {}".format(main_page_technology))

def main():
  menu()
  my_fuzzer = fuzzer(
    options.url_to_fuzz
    )

  my_fuzzer.tech_detection()


if __name__ == "__main__":
  main()
