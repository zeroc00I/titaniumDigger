#!/bin/bash
cat /tmp/dnsMonitor.cap | 
tr -cd '\n[a-zA-Z0-9_<]' | sed 's#v2_f#\n#g;s#zerocoolcf#zerocoolcf\n#g;s#<#\n#g;' | grep zerocool | sed 's#v2_ezerocoolcf##g'
