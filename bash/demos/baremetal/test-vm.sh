wget -O alpine.iso https://dl-cdn.alpinelinux.org/alpine/v3.24/releases/x86_64/alpine-standard-3.24.1-x86_64.iso
rm localhost.apkovl.tar.gz
rm my-alpine.iso
tar --owner=0 --group=0 -czf localhost.apkovl.tar.gz -C ovl .
xorriso -indev alpine.iso -outdev my-alpine.iso -map localhost.apkovl.tar.gz /localhost.apkovl.tar.gz -boot_image any replay

qemu-img create -f qcow2 data/system.img 100G

qemu-system-x86_64 -enable-kvm -m "4G" -cpu host \
  -drive file="data/system.img",if=virtio,format=qcow2 \
  -cdrom my-alpine.iso -boot d \
  -netdev bridge,id=net0,br="vmbr0" \
  -device virtio-net-pci,netdev=net0

ip a add 10.146.101.43/24 dev eth0
ip link set eth0 up
ip route add default via 10.146.101.1 dev eth0
echo "nameserver 1.1.1.1" > /etc/resolv.conf