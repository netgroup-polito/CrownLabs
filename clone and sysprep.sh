#!/bin/bash
#This script clone, sparsify and sysprep linux guest
original=$1
destination=$2
name=$original

check_original () {
if [ -z "${original}" ]; then
echo "This script clones, sparsifies and syspreps linux guest"
echo "Usage : '$0 <original guest> <destination guest>'"
echo "Please provide the guest name of a destroyed guest: exit"
exit
fi
guests_defined="$(virsh list --all --name)"
if grep -qvw "$original" <<< ${guests_defined}  ; then
echo "Please provide a defined guest name : exit"
echo "Guests avaible :"
echo "$(virsh list --all --name)"
exit
fi
}

check_destination () {
if [ -z $destination ] ; then
echo "Please provide the name of the destination guest: exit"
echo "For example : '$0 $name <destination guest>'"
exit
fi
if grep -qw "$destination" <<< $(virsh list --all --name)  ; then
echo "Please provide a non defined guest name : exit"
echo "Different than :"
echo "$(virsh list --all --name)"
exit
fi
}


clone () {
virt-clone -o $original -n $destination -f /var/lib/libvirt/images/$destination.qcow2
}

prepare () {
virt-sysprep -d $destination --operations customize --firstboot-command " sudo dbus-uuidgen > /etc/machine-id ; sudo hostnamectl set-hostname $destination ; touch /.autorelabel ; sudo reboot"
}

sparsify () {
echo "Sparse disk optimization"
virt-sparsify --check-tmpdir ignore --compress --convert qcow2 --format qcow2 /var/lib/libvirt/images/$destination.qcow2 /var/lib/libvirt/images/$destination.sparse
rm -rf /var/lib/libvirt/images/$destination.qcow2
mv /var/lib/libvirt/images/$destination.sparse /var/lib/libvirt/images/$destination.qcow2
chown qemu:qemu /var/lib/libvirt/images/$destination.qcow2 || chown libvirt-qemu:libvirt-qemu /var/lib/libvirt/images/$destination.qcow2
}

check_original
check_destination
clone
sparsify
prepare