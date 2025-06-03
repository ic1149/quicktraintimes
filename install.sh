cd ~
if [ -d ./qtt]; then
  rm -rf ./qtt
fi
mkdir qtt
cd qtt
wget github.com/ic1149/quicktraintimes/releases/latest/download/quicktraintimes.tar.xz
tar -xf quicktraintimes.tar.xz
sudo make install
wget github.com/ic1149/quicktraintimes/releases/latest/download/qtt.desktop
mv qtt.desktop /usr/share/applications/qtt.desktop
wget github.com/ic1149/quicktraintimes/releases/latest/download/qtt_icon.png
mv qtt_icon.png /usr/share/icons/qtt_icon.png
cd ~
rm -rf qtt