#!/bin/bash

cd ~
if [ -d ./qtt]; then
  rm -rf ./qtt # remove existing installer
fi
mkdir qtt #create temp dir
cd qtt

wget github.com/ic1149/quicktraintimes/releases/latest/download/quicktraintimes_1.0.1.tar.xz
tar -xf quicktraintimes_1.0.1.tar.xz

if which quicktraintimes; then
  sudo rm $(which quicktraintimes) # remove existing installation
fi

sudo make install #install quicktraintimes

cd ~
rm -rf qtt #remove installtion files

if ! grep -Fxq "alias qtt=" .bash_aliases; then
  echo 'alias qtt="quicktraintimes"' >> .bash_aliases
fi
