ARG UBUNTU_VERSION=20.04

FROM ubuntu:$UBUNTU_VERSION

ARG UBUNTU_VERSION  # FROM resets the vars so have to include this here
ARG ARKIME_VERSION=4.1.0
ARG ARKIME_PACKAGE=arkime_${ARKIME_VERSION}-1_amd64.deb
ARG ARKIME_URL=https://s3.amazonaws.com/files.molo.ch/builds/ubuntu-${UBUNTU_VERSION}/${ARKIME_PACKAGE}

# Required dependencies
RUN apt-get update && \
    apt-get install -yq curl libmagic-dev wget logrotate

RUN echo "curl -C - -O $ARKIME_URL"
RUN curl -C - -O "$ARKIME_URL"
RUN echo "apt install ./$ARKIME_PACKAGE"
RUN apt install ./$ARKIME_PACKAGE -y
RUN rm $ARKIME_PACKAGE

# Setup folders for logging and data output
RUN mkdir -p /opt/arkime/logs
RUN mkdir -p /opt/arkime/raw

# Add Arkime assets
COPY assets/oui.txt /opt/arkime/etc/oui.txt
COPY assets/ipv4-address-space.csv /opt/arkime/etc/ipv4-address-space.csv
COPY assets/GeoLite2-ASN.mmdb /opt/arkime/etc/GeoLite2-ASN.mmdb
COPY assets/GeoLite2-City.mmdb /opt/arkime/etc/GeoLite2-City.mmdb
COPY assets/GeoLite2-Country.mmdb /opt/arkime/etc/GeoLite2-Country.mmdb

# Add setup script
COPY setup.sh /opt/arkime/bin/setup.sh
RUN chmod 755 /opt/arkime/bin/setup.sh

EXPOSE 8005
