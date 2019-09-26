#!/bin/bash
ld -r -o ../libarchive_windows_amd64.syso --whole-archive libarchive.a --whole-archive libbz2.a --whole-archive liblzma.a --whole-archive libz.a
