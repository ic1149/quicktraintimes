#!/bin/bash

cd ~
if [ -d ./qtt]; then
  rm -rf ./qtt # remove existing installer
fi
mkdir qtt #create temp dir
cd qtt

wget github.com/ic1149/quicktraintimes/releases/latest/download/quicktraintimes.tar.xz
tar -xf quicktraintimes.tar.xz

if which quicktraintimes; then
  sudo rm $(which quicktraintimes) # remove existing installation
fi

sudo make install #install quicktraintimes


wget github.com/ic1149/quicktraintimes/blob/main/qtt.desktop
if [ -f /usr/share/applications/qtt.desktop ]; then
  sudo rm /usr/share/applications/qtt.desktop
fi
sudo mv qtt.desktop /usr/share/applications/qtt.desktop #desktop file

wget github.com/ic1149/quicktraintimes/blob/main/qtt_icon.png
if [ -f /usr/share/icons/qtt_icon.png ]; then
  sudo rm /usr/share/icons/qtt_icon.png
fi
sudo mv qtt_icon.png /usr/share/icons/qtt_icon.png #icon
cd ~
rm -rf qtt #remove installtion files