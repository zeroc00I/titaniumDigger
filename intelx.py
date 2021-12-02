#
# by @zeroc00I - 2021
# api-key configured bellow is public and related to unauth users according intelx.io
#

import requests, optparse, json
from requests.packages.urllib3.exceptions import InsecureRequestWarning
requests.packages.urllib3.disable_warnings() 

def menu():
    parser = optparse.OptionParser()
    parser.add_option('-e', '--email', dest="email", help='admin@example.com')
    parser.add_option('-k','--key', dest="key", default="e65df1a7-689a-4738-a078-598860554bf8")

    options, args = parser.parse_args()

    if not options.email:
        print("[INFO] Please, enter some email e.g: -e admin@example.com")
        exit()

    header = {
        "x-key": options.key,
        "Content-Type": "application/x-www-form-urlencoded", 
        "charset":"UTF-8"
    }
    globals().update(locals())

def search():
    url = 'https://api.intelx.io/intelligent/search'

    payload = '{{"term":"{}","lookuplevel":0,"maxresults":100,"timeout":null,"datefrom":"","dateto":"","sort":2,"media":0,"terminate":[]}}'.format(options.email)

    r = requests.post(url,verify=False,data=payload,headers=header)
    
    if r.status_code == 401:
        print("[Error] Oh no! \n:: Intelx.io said that this key got banned ¯\_(ツ)_/¯\n:: Exiting ...")
        exit()

    if r.status_code == 402:
        print("[Error] Oh no! \n:: Intelx.io have just blocked your IP\n:: Wait some minutes to get some more free queries\n:: Or re-run with -k entering with your apikey already registered")
        exit()

    return json.loads(r.content)['id']

def find_documents(uid):
    url = 'https://api.intelx.io/intelligent/search/result?id={}&limit=10&statistics=1&previewlines=8'.format(uid)

    r = requests.get(url,verify=False,headers=header)

    return json.loads(r.content)

def get_documents_preview(documents):
    
    if len(documents['records']) == 0:
        print(":: Ops! No results found with {}".format(options.email))
    
    for x in range(0,len(documents['records'])):
        storageid = documents['records'][x]["storageid"]
        url = 'https://api.intelx.io/file/preview?sid={}&f=0&l=22&c=1&m=24&b=leaks.private.general&k={}'.format(storageid,options.key)

        r = requests.get(url,verify=False,headers=header)
        print(r.content.decode())

def main():
    menu()
    search_result_uid = search()
    documents = find_documents(search_result_uid)
    get_documents_preview(documents)

if __name__== "__main__":
    main()
