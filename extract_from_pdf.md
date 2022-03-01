## Dependencies

```
apt-get install build-essential libpoppler-cpp-dev pkg-config python-dev
```

## extract.sh

```
#!/bin/bash
tmp="$(mktemp)"
curl -L --insecure "$1" 2>/dev/null > $tmp
pdftotext -enc "UTF-8" $tmp - |
tr ' ' '\n' |
grep "\.mil" |
grep -v "@" |
tr -d "()"
rm -f "$tmp"
```

## Fire!

```
grep "\.pdf$" hakrawlerResultsFromManyHosts| xargs -P300 -I@ bash -c '/opt/pdftotext/extract.sh @ 2>/dev/null' 2>/dev/null  | anew linksfrompdf
```
