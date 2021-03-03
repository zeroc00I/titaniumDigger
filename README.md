# titaniumDigger

Keep it secret ;)

# todo
Criar helper 
https://github.com/vavkamil/XFFenum/blob/master/xffenum.py

### Modos:
Discover
Xss

## 1 Parte: Payloads

### Argumentos:
Discover: dominios
Xss: wordlists payload, keys values

Cria diretorio com nome do dominio

Chamada hakrawler -url github.com -wayback -plain

Tratar urls com extensões que não queremos

echo site.com.br/teste.php | grep --color -E '\.[a-zA-Z0-9]{2,4}$'

site.com.br/teste.php

Tratar pra pegar paths keys

### Resultado:
Incrementa arquivo keys do diretorio do dominio
Incrementa path keys do diretorio do dominio

## 2 Xss
Checar argumento dominios e ver se url está viva

# Ideias
Pegar parametros e urls dentro dos JS
https://github.com/m4ll0k/Bug-Bounty-Toolz/blob/master/getjswords.py
https://github.com/KathanP19/JSFScan.sh/blob/master/install.sh
