# 389ds / FreeIPA Exporter [![Build Status](https://travis-ci.org/terrycain/389ds_exporter.svg)][travis]

[![CircleCI](https://circleci.com/gh/terrycain/389ds_exporter/tree/master.svg?style=shield)][circleci]

Started out as just a replication status exporter, and evolved to export more FreeIPA related objects.

Is my first stab at go, it works but it could be better and I hope to improve it. The main thing is that it has a loop that
hits LDAP and performs the queries, it doesn't query LDAP when /metrics is queried. What I want is to query LDAP when /metrics
is hit and then cache for a period of time (incase multiple prometheus are running).

To run:
```bash
make
./389ds_exporter [flags]
```

## Exported Metrics

| Metric | Meaning | Labels |
| ------ | ------- | ------ |
| ldap_389ds_users | Number of FreeIPA users | active, staged, preserved |
| ldap_389ds_groups | Number of FreeIPA groups | |
| ldap_389ds_hosts | Number of FreeIPA hosts | |
| ldap_389ds_hostgroups | Number of FreeIPA hostgroups | |
| ldap_389ds_hbac_rules | Number of FreeIPA HBAC rules | |
| ldap_389ds_sudo_rules | Number of FreeIPA SUDO rules | |
| ldap_389ds_dns_zones | Number of FreeIPA DNS zones (including forward zones) | |
| ldap_389ds_replication_conflicts | Number of LDAP replication conflicts | |
| ldap_389ds_replication_status | Replication status of peered 389ds nodes (1 good, 0 bad) | server |
| ldap_389ds_scrape_count | Number of successful or unsuccessful scrapes | result |
| ldap_389ds_scrape_duration_seconds | How long the last scrape took |

### Flags

```bash
./389ds_exporter --help
```

* __`-debug`:__ Debug logging
* __`-interval`:__ Scrape interval (default 1m0s)
* __`-ipa-domain`:__ FreeIPA domain e.g. example.org
* __`-ldap.addr`:__ Address of 389ds server (default "localhost:389")
* __`-ldap.pass`:__ 389ds Directory Manager password
* __`-ldap.user`:__ 389ds Directory Manager user (default "cn=Directory Manager")
* __`-log-json`:__ JSON formatted log messages
* __`-web.listen-address`:__ Bind address for prometheus HTTP metrics server (default ":9496")
* __`-web.telemetry-path`:__ Path to expose metrics on (default "/metrics")

### Using docker

TBD


### Credits

This repo essentially started off as a clone of the openldap_exporter modified to query
some FreeIPA DNs. The openldap_exporter was a great help in getting this started, as was
the consul_exporter which served as a great reference on how to package a prometheus
exporter.

[circleci]: https://circleci.com/gh/terrycain/389ds_exporter
[travis]: https://travis-ci.org/terrycain/389ds_exporter
