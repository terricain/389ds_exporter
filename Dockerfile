FROM        quay.io/prometheus/busybox:latest
LABEL maintainer="terrycain"

COPY 389ds_exporter /bin/389ds_exporter

EXPOSE     9496
ENTRYPOINT [ "/bin/389ds_exporter" ]
