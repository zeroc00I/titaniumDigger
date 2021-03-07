import requests
import argparse
import os

argumentList = []
urlList = []

def loadArgumentList(argumentListPath):
    for line in argumentListPath.readlines():
        lineArgumentWithoutN = line.replace('\n','')
        argumentList.append(lineArgumentWithoutN)

def loadURLList(urlListPath):
    for line in urlListPath.readlines():
        lineFileWithoutN = line.replace('\n','')
        urlList.append(lineFileWithoutN)

def makeDefinitiveList(urlList, argument_list):
    definitiveList = []

    for urlEntry in urlList:
        for argumentEntry in argumentList:
            definitiveList.append(urlEntry+"?"+argumentEntry+"=198522355kkll")
    return definitiveList

def makeRequest(definitiveList):
    payload={}
    for url in definitiveList:
        try:
            response = requests.request("GET", url, data=payload)
        except:
            print('Erro no request em: '+url)
            continue
        if ("198522355kkll" in response.text):
            print("Refletido em:"+url)
        else:
            print("Não achou para %s" % url)     

def menu():

    parser = argparse.ArgumentParser() 
    group = parser.add_mutually_exclusive_group(required=True)
    parser.add_argument('-m', help='Max Seconds Request timeout', dest='maxReqTimeout', type=int)
    parser.add_argument('-r', help='Max request header per request', dest='maxHeaderReq', type=int)
    parser.add_argument('-w', help='Payloads from wordlist file', required='True', dest='payloadWordlist', type=argparse.FileType('r'))
    group.add_argument('-u', help='Target single url', dest='url')
    group.add_argument('-U', help='Target multiple url from file', dest='urlsFromFile', type=argparse.FileType('r'))
    args = parser.parse_args()
    
    return args

def main():

    args = menu()
    loadArgumentList(args.payloadWordlist)
    loadURLList(args.urlsFromFile) 
    
    definitiveList = makeDefinitiveList(urlList, argumentList) 
    makeRequest(definitiveList)

if __name__ == "__main__":
    main()




