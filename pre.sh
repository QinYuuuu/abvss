#!/bin/bash

if [ -d "abvss/" ];then
    rm -rf abvss;
fi
<<<<<<< HEAD
git clone https://github.com/QinYuuuu/abvss;
cd abvss;
/usr/local/go/bin/go env -w GO111MODULE=auto;
/usr/local/go/bin/go mod init github.com/QinYuuuu/abvss;
=======
git clone https://abvss;
cd abvss;
/usr/local/go/bin/go env -w GO111MODULE=auto;
/usr/local/go/bin/go mod init abvss;
>>>>>>> 19b0d27 (Initial commit)
/usr/local/go/bin/go mod tidy