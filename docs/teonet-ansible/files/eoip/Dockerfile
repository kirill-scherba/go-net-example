# EOIP tunnels
#
# EOIP tunnels are handy feature of RouterOS allowing easy setup of 
# high-performance ethernet VPNs. For documentation and setup guide, 
# see http://wiki.mikrotik.com/wiki/Manual:Interface/EoIP

# Docker builder
# 
FROM ubuntu AS builder

# Create eoip executable
WORKDIR /root
RUN apt update && apt install -y git build-essential
RUN git clone https://github.com/katlogic/eoip.git
WORKDIR /root/eoip
RUN make

# #############################################################
# Compose production image
#
FROM ubuntu:19.04 AS production
WORKDIR /root

# Install previously built application
COPY --from=builder /root/eoip/eoip /usr/local/bin/
CMD ["/usr/local/bin/eoip"]

# Set dev testing dependenses
RUN apt update && apt install -y net-tools traceroute iputils-ping


# Build docker image
#
# docker build --tag eoip .


# Run eoip container from this image and mount tap interface to host
#
# docker run --name eoip1 --cap-add=NET_ADMIN --device /dev/net/tun:/dev/net/tun --network host -d --rm eoip eoip tap0 172.17.0.1 172.17.0.3:10
# sudo ifconfig tap0 10.0.12.1/30


# Run eoip container from this image and mount tap interface inside container
#
# docker run --name eoip2 --cap-add=NET_ADMIN --device /dev/net/tun:/dev/net/tun -it --rm eoip bash 
# eoip tap0 172.17.0.3 172.17.0.1:10 &
# ifconfig tap0 10.0.12.2/30


# Run eoip (inside container or in host)
#
# sudo ./eoip tap0 172.17.0.1 172.17.0.3:10 &
# sudo ifconfig tap0 10.0.12.1/30


# Copy executable from container
#
# docker build --tag eoip .
# docker run --name eoip -it eoip ls -al
# docker cp eoip:/root/eoip/eoip ./
# docker rm eoip
# docker rmi eoip
#
# or use
#
# ./build_eoip.sh
