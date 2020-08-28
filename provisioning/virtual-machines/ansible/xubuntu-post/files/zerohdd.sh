echo "u" > /proc/sysrq-trigger
mount /dev/mapper / -o remount,ro
zerofree -v /dev/sda1
shutdown -h now
