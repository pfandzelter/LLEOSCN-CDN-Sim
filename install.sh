#!/bin/sh

# update yum repos
sudo yum update -y

# install pip3 (if needed)
sudo yum install python3-pip -y
pip3 install pip --upgrade --user

# install dependencies
sudo yum install libtool -y
sudo yum install python-devel -y
sudo yum install python3-devel -y
sudo yum install gcc gcc-c++ byacc -y
sudo yum install llvm -y

pip3 install --user -r ./requirements.txt

# install go
sudo yum install git -y
sudo wget https://golang.org/dl/go1.15.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.15.5.linux-amd64.tar.gz
yes | rm go1.15.5.linux-amd64.tar.gz
echo "export PATH=/usr/local/go/bin/:\$PATH" >> .bash_profile
export PATH=/usr/local/go/bin/:\$PATH

# build binaries
(
  cd ./analysis || exit
  go get ./...
  go build .
)

(
  cd ./caching || exit
  go get ./...
  go build .
)
