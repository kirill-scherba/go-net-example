################################################################################
# Teonun - teonet tunnel create TAP interface with selected IP and connect to 
# remote host by UDP 
#
#
FROM ubuntu

WORKDIR /root
COPY teotun .

# For development and debug
RUN apt update && apt install -y iproute2 net-tools iputils-ping

# Build it:
#
#   go get
#   go build
#   docker build --tag teotun .
#
# Run in host:
#
#   sudo ./teotun -l 10.0.0.1 -a 192.18.0.1
#
# Run in docker: 
#
#   docker run --name teotun --cap-add=NET_ADMIN --device /dev/net/tun:/dev/net/tun -it --rm teotun ./teotun -l 10.0.0.2 -a 172.17.0.1
#
# Run in another terminal:
#   
#   ping 10.0.0.2   
#   docker exec -it teotun bash
#   ping 10.0.0.1
#