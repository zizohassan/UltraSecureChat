export PATH=/Users/abdelazizhassan/gcc-linaro-7.5.0-2019.12-x86_64_arm-linux-gnueabihf/bin:$PATH
which arm-linux-gnueabi-gcc
export CC=arm-linux-gnueabi-gcc
export CXX=arm-linux-gnueabi-g++
export CGO_ENABLED=1
export GOOS=windows
export GOARCH=arm
go build -o chat_client_arm.exe gui.go
#

#export CC=x86_64-w64-mingw32-gcc
#export CXX=x86_64-w64-mingw32-g++
#export CGO_ENABLED=1
#export GOOS=windows
#export GOARCH=arm
#go build -o chat_client_arm.exe gui.go
