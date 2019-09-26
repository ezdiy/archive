#!/bin/bash
for n in *.a; do ld -r -o ../${n%.*}_windows_amd64.syso --whole-archive $n; done
