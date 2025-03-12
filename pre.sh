#!/bin/bash

if [ -d "abvss/" ];then
    rm -rf abvss;
fi
<<<<<<< HEAD
<<<<<<< HEAD
git clone https://github.com/QinYuuuu/abvss;
cd abvss;
/usr/local/go/bin/go env -w GO111MODULE=auto;
/usr/local/go/bin/go mod init github.com/QinYuuuu/abvss;
=======
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
git clone https://abvss;
cd abvss;
/usr/local/go/bin/go env -w GO111MODULE=auto;
/usr/local/go/bin/go mod init abvss;
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
/usr/local/go/bin/go mod tidy