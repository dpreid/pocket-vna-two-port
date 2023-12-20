#!/bin/bash

# we assume jump is already installed
# we assume that relay access and token files are already present in /etc/practable
# assume that /etc/practable/id has the correct id for the experiment, e.g. pvna05

curl -sSL https://get.docker.com | sh
sudo usermod -aG docker $USER 
sudo systemctl enable docker

cp ./files/vna-data /usr/local/bin/
cp ./files/relay-rules /usr/local/bin/
cp ./files/odroid/vna /usr/local/bin
cp ./files/odroid/relay /usr/local/bin 
cp ./services/* /etc/systemd/system
cp ../lib/arm64/libPocketVnaApi.so /usr/lib/libPocketVnaApi.so.0
cp ../lib/arm64/libPocketVnaApi.so /usr/lib/libPocketVnaApi.so.1

#programme the arduino
curl -fsSL https://raw.githubusercontent.com/arduino/arduino-cli/master/install.sh | sh
./bin/arduino-cli core update-index
./bin/arduino-cli core install arduino:avr
./bin/arduino-cli lib install timerinterrupt
./bin/arduino-cli compile --fqbn arduino:avr:nano ../fw/RFSwitch/ 
./bin/arduino-cli upload --port /dev/ttyUSB0 --fqbn arduino:avr:nano ../fw/RFSwitch/

#get relay tokens
export FILES=$(cat /home/odroid/files.link)
export PRACTABLE_ID=$(cat /etc/practable/id)
cd /etc/practable
wget $FILES/st-ed0-data.access.$PRACTABLE_ID -O  st-ed0-data.access
wget $FILES/st-ed0-data.token.$PRACTABLE_ID -O   st-ed0-data.token

# start services
systemctl enable calibration.service
systemctl enable relay.service
systemctl enable relay-rules.service
systemctl enable vna-data.service
systemctl start calibration.service
systemctl start relay.service
systemctl start relay-rules.service
systemctl start vna-data.service







