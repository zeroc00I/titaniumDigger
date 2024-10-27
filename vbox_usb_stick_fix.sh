#!/bin/bash

#
# Heavily inspired by https://github.com/dnschneid/crouton/wiki/VirtualBox-udev-integration
#

vbox_usbnode_path=$(find / -name VBoxCreateUSBNode.sh 2> /dev/null | head -n 1)

if [[ -z $vbox_usbnode_path ]]; then
    echo "Error: VBoxCreateUSBNode.sh file has not been found."
    exit 1
fi

login_name=${1:-"$(logname)"}

if [[ -z $login_name ]]; then
    echo "Error: Login name has not been found; run script again with user name as the first argument."
    exit 1
fi

vboxusers_gid=$(getent group vboxusers | awk -F: '{printf "%d\n", $3}')

if [[ -z $vboxusers_gid ]]; then
    echo "Error: Group 'vboxusers' has not been found; you may need to reinstall VirtualBox."
    exit 1
fi

vbox_rules="SUBSYSTEM==\"usb_device\", ACTION==\"add\", RUN+=\"$vbox_usbnode_path \$major \$minor \$attr{bDeviceClass} $vboxusers_gid\"
SUBSYSTEM==\"usb\", ACTION==\"add\", ENV{DEVTYPE}==\"usb_device\", RUN+=\"$vbox_usbnode_path \$major \$minor \$attr{bDeviceClass} $vboxusers_gid\"
SUBSYSTEM==\"usb_device\", ACTION==\"remove\", RUN+=\"$vbox_usbnode_path --remove \$major \$minor\"
SUBSYSTEM==\"usb\", ACTION==\"remove\", ENV{DEVTYPE}==\"usb_device\", RUN+=\"$vbox_usbnode_path --remove \$major \$minor\""

chmod 755 $vbox_usbnode_path
chown root:root $vbox_usbnode_path
rm -f /etc/udev/rules.d/*-virtualbox.rules
echo "$vbox_rules" > /etc/udev/rules.d/99-virtualbox.rules
udevadm control --reload

if command -v usermod > /dev/null 2>&1; then
	usermod -aG vboxusers "$login_name"
else
	adduser "$login_name" vboxusers
fi

echo "All actions succeeded."
echo "Log out and log in to see if the issue got fixed."
