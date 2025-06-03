#!/bin/bash

cd ~
if [ -d ./qtt]; then
  rm -rf ./qtt # remove existing installation
fi
mkdir qtt #create temp dir
cd qtt
wget github.com/ic1149/quicktraintimes/releases/latest/download/quicktraintimes.tar.xz
tar -xf quicktraintimes.tar.xz
sudo make install #install quicktraintimes
wget github.com/ic1149/quicktraintimes/releases/latest/download/qtt.desktop
mv qtt.desktop /usr/share/applications/qtt.desktop #desktop file
wget github.com/ic1149/quicktraintimes/releases/latest/download/qtt_icon.png
mv qtt_icon.png /usr/share/icons/qtt_icon.png #icon
cd ~
rm -rf qtt #remove installtion files