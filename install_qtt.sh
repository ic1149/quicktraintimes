#!/bin/bash

cd ~
if [ -d ./quicktraintimes_installer]; then
  rm -rf ./quicktraintimes_installer # remove existing installer
fi
mkdir quicktraintimes_installer #create temp dir
cd quicktraintimes_installer

wget https://github.com/ic1149/quicktraintimes/releases/latest/download/quicktraintimes.tar.xz
tar -xf quicktraintimes.tar.xz

if which quicktraintimes; then
  sudo rm $(which quicktraintimes) # remove existing installation
fi

sudo make install #install quicktraintimes

cd ~
rm -rf quicktraintimes_installer #remove installtion files

if ! grep -Fq "alias qtt=" .bash_aliases; then
  echo 'alias qtt="quicktraintimes"' >> .bash_aliases
fi
