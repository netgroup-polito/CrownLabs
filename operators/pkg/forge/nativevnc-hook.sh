#!/bin/sh
set -e
domain_xml="$4"
if ! echo "$domain_xml" | grep -q 'xmlns:qemu='; then
  domain_xml=$(echo "$domain_xml" | sed "s#<domain type='kvm'>#<domain type='kvm' xmlns:qemu='http://libvirt.org/schemas/domain/qemu/1.0'>#")
fi
qemu_cmdline='  <qemu:commandline>\n    <qemu:arg value="-vnc"/>\n    <qemu:arg value="0.0.0.0:0,websocket=$NATIVEVNCPORT"/>\n  </qemu:commandline>\n'
domain_xml=$(echo "$domain_xml" | sed "s#</domain>#${qemu_cmdline}</domain>#")
echo "$domain_xml"
