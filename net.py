from bs4 import BeautifulSoup
import requests, optparse, json
from requests.packages.urllib3.exceptions import InsecureRequestWarning

def menu():
    parser = optparse.OptionParser()
    parser.add_option('-t', '--threads', dest="threads", help='100')
    parser.add_option('-f', '--file', dest="file", help='example.txt')
    parser.add_option('-a', '--asn', dest="asn", help='AS9090')
    parser.add_option('-o', '--org', dest="org", help='AT&T')
    parser.add_option('--ec', '--extract_cidr',action="store_true", dest="extract_cidr", help='Extract all cidrs from ASNs')
    parser.add_option('--ed', '--extract_domains',action="store_true", dest="extract_domains", help='Extract all domains from orgs ASNs')


    options, args = parser.parse_args()
    globals().update(locals())

def get_asn_from_org():

    endpoint = 'https://api.asrank.caida.org/v2/graphql/'
    query = """query{
                        asns(name:\""""+options.org+"""\",first:40,offset:0,sort:"rank",seen:true) {
                        totalCount
                            edges {
                                node {
                                    rank
                                    asn
                                    asnName
                                    organization {
                                        orgId
                                        orgName
                                    }
                                }
                            }
                        }
                    }
                """
    r = requests.post(endpoint, json={"query": query})
    if r.status_code == 200:
        return json.loads(r.content)
    else:
        raise Exception(f"Query failed to run with a {r.status_code}.")

def format_html_with_asns(prefixes):
    soup = BeautifulSoup(prefixes, 'html.parser')
    soup2 = soup.find_all('td')
    for x in soup2:
        if '/' in x.get_text():
            print(x.get_text())

def get_asn_prefixes(asn):
    r = requests.get('https://www.bigdatacloud.com/asn-lookup/{}/prefixes'.format(asn))
    return r.content

def get_domains_from_org(org):
    r = requests.get('https://crt.sh/?q={}&output=json'.format(org))
    domains = json.loads(r.content)
    for domain in domains:
        print(domain['common_name'])
    #return r.content

def main():
    menu()
    if options.asn:
        print('[ASN] Querying ...')
        if "," in options.asn:
            array_asn = options.asn.split(',')
            print('[Multiple ASN] {}'.format(options.asn))
            for asn in array_asn:
                prefixes = get_asn_prefixes(asn)
                format_html_with_asns(prefixes)
        else:
            prefixes = get_asn_prefixes(options.asn)
            format_html_with_asns(prefixes)

    if options.org:
        print('[Graphql Org] Querying ...')
        all_asn_from_org = get_asn_from_org()
        print("[Total] {} ".format(all_asn_from_org['data']['asns']['totalCount']))

        if options.extract_cidr:
            for count in range(0,len(all_asn_from_org['data']['asns']['edges'])):
                main_node = all_asn_from_org['data']['asns']['edges'][count]['node']['asn']
                prefixes = get_asn_prefixes(main_node)
                format_html_with_asns(prefixes)

        if options.extract_domains:
            for count in range(0,len(all_asn_from_org['data']['asns']['edges'])):
                main_node = all_asn_from_org['data']['asns']['edges'][count]['node']
                org = main_node['organization']['orgName']
                get_domains_from_org(org)

        else:
            for count in range(0,len(all_asn_from_org['data']['asns']['edges'])):
                main_node = all_asn_from_org['data']['asns']['edges'][count]['node']
                print("[ASN] {} [Org] {}".format(main_node['asn'],main_node['organization']['orgName']))
            #print("[All] {} ".format(json.dumps(all_asn_from_org['data'],indent=2)))

if __name__ == "__main__":
    main()
