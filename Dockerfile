FROM alpine

LABEL org.opencontainers.image.source https://github.com/carldanley/ha-fpp-mqtt

RUN apk upgrade --no-cache \
  && apk --no-cache add \
    tzdata zip ca-certificates

WORKDIR /usr/share/zoneinfo
RUN zip -r -0 /zoneinfo.zip .
ENV ZONEINFO /zoneinfo.zip

WORKDIR /
ADD dist/ha-fpp-mqtt_linux_arm64/ha-fpp-mqtt /bin/

ENTRYPOINT [ "/bin/ha-fpp-mqtt" ]
