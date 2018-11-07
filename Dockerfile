FROM        quay.io/prometheus/busybox:latest
LABEL maintainer="The Prometheus Authors <prometheus-developers@googlegroups.com>"

COPY 389ds_exporter /bin/389ds_exporter

EXPOSE     9496
ENTRYPOINT [ "/bin/389ds_exporter" ]
