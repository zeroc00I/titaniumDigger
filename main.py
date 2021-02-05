#!/bin/python3
from bs4 import BeautifulSoup
import requests
import sys

requests.packages.urllib3.disable_warnings() 
urlToFetch = str(sys.argv[1])
baseUrl = urlToFetch.split("/")
baseUrl.pop()
baseUrl = '/'.join(baseUrl)

def getPage(urlToFetch):
	html_content = requests.get(urlToFetch,verify=False).text
	extractInfoFromPage(html_content)

def extractInfoFromPage(page):
	pageHtml = BeautifulSoup(page, "lxml")
	allForms = pageHtml.findAll('form')
	for f in allForms:
		if f.get('action'):
			if f.get('action').startswith('/'):
				buildFormExploit(f,baseUrl+f['action']);
			else:
				buildFormExploit(f,baseUrl+"/"+f['action'])
		else:
			print('[DEBUG] Form without action / is this react?')
			if f.get('data-reactid'):
				print('[INFO] React Form')
			else:
				print('[INFO] What form is this') 

def buildFormExploit(form,targetAction):
	allInputs = form.findAll('input')
	request = {};
	request['action']=targetAction

	if form.get('method'):
		request['method'] = form.get('method')

	for input in allInputs:
		if input.get('name'):
			request['payload'] = request.get('payload', '')+"&"+input.get('name')
			if input.get('value'):
				request['payload'] = request.get('payload', '')+"="+input.get('value')
			else:
				request['payload'] = request.get('payload', '')+"=isThisWillBeReflectedOnpage?"
	print(request)

def fire(req):
	result = requests.get(req,verify=False)
	print(result)

if __name__ == "__main__":
	getPage(urlToFetch)