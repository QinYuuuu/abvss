#!/bin/bash

if [ -d "abvss/" ];then
    rm -rf abvss;
fi
git clone https://github.com/QinYuuuu/abvss;
cd abvss;
/usr/local/go/bin/go env -w GO111MODULE=auto;
/usr/local/go/bin/go mod init github.com/QinYuuuu/abvss;
/usr/local/go/bin/go mod tidy