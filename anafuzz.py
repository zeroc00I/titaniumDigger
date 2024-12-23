import requests, optparse, os
from requests.packages.urllib3.exceptions import InsecureRequestWarning
from urllib.parse import urlparse
from lxml.html import soupparser

requests.packages.urllib3.disable_warnings()

def menu():
    parser = optparse.OptionParser()
    parser.add_option('-u', '--url', dest="url_to_fuzz", help='https://example.zerocool.com/item?search=waf')
    parser.add_option('-w', '--wordlist', dest="raw_wordlist_without_extension", help='Wordlist file without extension')
    parser.add_option('-t', '--threads', dest="threads", default=20, type=int, help='options to set threads numbers to fuzz')
    parser.add_option('-v', '--verbose', dest="verbose_mode_status",default=False, help='should verbose mode be enabled?')

    options, args = parser.parse_args()

    if not options.url_to_fuzz:
      print("[-] Please enter some URL with -u parameter (Check -h for help)")
      exit()
    if not options.raw_wordlist_without_extension:
      print("[-] Please enter some WORDLIST with -w parameter (Check -h for help)")
      exit()

    globals().update(locals())

class Requester:
    def __init__(self,url):
        self.url = url
        self.status_error = False
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

        except Exception:
          self.status_error = True
          return

        return self.request.text

    def get_response_word_number(self):
        array_response = self.request.text.split(' ')
        array_response_without_empty_values = [x for x in array_response if x]

        return len(array_response_without_empty_values)

    def get_final_url(self):
      return self.request.url

    def get_web_server(self):
      if 'Server' in self.request.headers.keys():
        if '/' in self.request.headers['Server']:
          self.webserver = self.request.headers['Server'].split('/')[0]
          return self.webserver
        else:
          self.webserver = self.request.headers['Server']
          return self.webserver

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

class setup:
  def __init__(self,url):
    self.url = url
    self.baseline()
    self.tech_detection()

  def baseline(self):
    self.baseline = Requester(self.url)
    self.baseline_arquivo = Requester(self.url+"/xhopefullydoesntexistx.html")
    self.baseline_diretorio = Requester(self.url+"/xhopefullydoesntexistx/")

  def wordlist_based_on_webserver(self):
    webserver = self.webserver.lower()
    wordlist_by_extension = {
    "apache":["php","html"],
    "tomcat":"jsp",
    "coldfusion":"cf",
    "iis":["asp","aspx"]
    } 
    if webserver in wordlist_by_extension.keys():
      self.wordlist_extension_to_be_used = wordlist_by_extension[webserver]
      print(self.wordlist_extension_to_be_used)

    else:
      print('Wordlist based on webserver {} not found'.format(webserver))
      exit()

  def tech_detection(self):
    main_page_technology =  self.baseline.get_page_tech_detection()
    
    if main_page_technology is None:
      self.webserver = self.baseline.get_web_server() if self.baseline.get_web_server() else None
      self.wordlist_based_on_webserver()
    
    else:
      self.wordlist_extension_to_be_used = main_page_technology

class fuzzer:
  def __init__(self,url,wordlist):
    self.url = url
    self.setup = setup(url)
    self.wordlist = wordlist
    self.common_words_length = []
    self.fire()

  def fire(self):
    words = open(self.wordlist).readlines()
    
    for fuzz_word in words:
      fuzz_word = fuzz_word.replace('\n','')
      url_to_fuzz = self.url+"/"+fuzz_word+"."+self.setup.wordlist_extension_to_be_used
      final_request = Requester(url_to_fuzz)

      if not final_request.status_error:
        word_number = final_request.get_response_word_number()

        if int(word_number) not in self.common_words_length:
          print('[+] {}'.format(url_to_fuzz))
          self.common_words_length.append(
            word_number
          )
        #else:
          #print('[-] {}'.format(self.common_words_length))

def main():
  menu()
  fuzzer(
    options.url_to_fuzz,
    options.raw_wordlist_without_extension
    )

if __name__ == "__main__":
  main()
