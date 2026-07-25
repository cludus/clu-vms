ip a add 10.146.101.43/24 dev eth0
ip link set eth0 up
ip route add default via 10.146.101.1 dev eth0
echo "nameserver 1.1.1.1" > /etc/resolv.conf