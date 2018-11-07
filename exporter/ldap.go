package exporter

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ldap.v2"
)

var (
	ldap_escaper = strings.NewReplacer("=", "\\=", ",", "\\,")
	usersGauge   = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "ldap_389ds",
			Name:      "users",
			Help:      "Number of user accounts",
		},
		[]string{"type"},
	)
	groupsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "ldap_389ds",
			Name:      "groups",
			Help:      "Number of groups",
		},
		[]string{},
	)
	hostsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "ldap_389ds",
			Name:      "hosts",
			Help:      "Number of hosts",
		},
		[]string{},
	)
	hostGroupsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "ldap_389ds",
			Name:      "hostgroups",
			Help:      "Number of hostgroups",
		},
		[]string{},
	)
	hbacRulesGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "ldap_389ds",
			Name:      "hbac_rules",
			Help:      "Number of hbac rules",
		},
		[]string{},
	)
	sudoRulesGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "ldap_389ds",
			Name:      "sudo_rules",
			Help:      "Number of sudo rules",
		},
		[]string{},
	)
	dnsZonesGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "ldap_389ds",
			Name:      "dns_zones",
			Help:      "Number of dns zones",
		},
		[]string{},
	)
	ldapConflictsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "ldap_389ds",
			Name:      "replication_conflicts",
			Help:      "Number of ldap conflicts",
		},
		[]string{},
	)
	replicationStatusGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "ldap_389ds",
			Name:      "replication_status",
			Help:      "Replication status by server",
		},
		[]string{"server"},
	)
	scrapeCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "ldap_389ds",
			Name:      "scrape_count",
			Help:      "successful vs unsuccessful ldap scrape attempts",
		},
		[]string{"result"},
	)
	scrapeDurationGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "ldap_389ds",
			Name:      "scrape_duration_seconds",
			Help:      "time taken per scrape",
		},
		[]string{},
	)
)

func init() {
	prometheus.MustRegister(
		usersGauge,
		groupsGauge,
		hostsGauge,
		hostGroupsGauge,
		hbacRulesGauge,
		sudoRulesGauge,
		dnsZonesGauge,
		ldapConflictsGauge,
		replicationStatusGauge,
		scrapeCounter,
		scrapeDurationGauge,
	)
}

func objectClass(name string) string {
	return fmt.Sprintf("(objectClass=%v)", name)
}

func ScrapeMetrics(ldapAddr, ldapUser, ldapPass, ipaDomain string) {
	start := time.Now()
	if err := scrapeAll(ldapAddr, ldapUser, ldapPass, ipaDomain); err != nil {
		scrapeCounter.WithLabelValues("fail").Inc()
		log.Error("Scrape failed, error is:", err)
	} else {
		scrapeCounter.WithLabelValues("ok").Inc()
	}
	elapsed := time.Since(start).Seconds()
	scrapeDurationGauge.WithLabelValues().Set(float64(elapsed))
	log.Infof("Scrape completed in %f seconds", elapsed)
}

func scrapeAll(ldapAddr, ldapUser, ldapPass, ipaDomain string) error {
	suffix := "dc=" + strings.Replace(ipaDomain, ".", ",dc=", -1)

	l, err := ldap.Dial("tcp", ldapAddr)
	if err != nil {
		return err
	}
	defer l.Close()

	err = l.Bind(ldapUser, ldapPass)
	if err != nil {
		return err
	}

	var errs error
	// Search for standard accounts
	log.Debug("getting active accounts")
	num, err := ldapSubordinatesQuery(l, fmt.Sprintf("cn=users,cn=accounts,%s", suffix), "(objectClass=*)")
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	usersGauge.WithLabelValues("active").Set(num)

	// Search for staged accounts
	log.Debug("getting staged accounts")
	num, err = ldapSubordinatesQuery(l, fmt.Sprintf("cn=staged users,cn=accounts,cn=provisioning,%s", suffix), "(objectClass=*)")
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	usersGauge.WithLabelValues("staged").Set(num)

	// Search for deleted accounts
	log.Debug("getting preserved accounts")
	num, err = ldapSubordinatesQuery(l, fmt.Sprintf("cn=deleted users,cn=accounts,cn=provisioning,%s", suffix), "(objectClass=*)")
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	usersGauge.WithLabelValues("preserved").Set(num)

	// Search for groups
	log.Debug("getting groups")
	num, err = ldapSubordinatesQuery(l, fmt.Sprintf("cn=groups,cn=accounts,%s", suffix), "(objectClass=*)")
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	groupsGauge.WithLabelValues().Set(num)

	// Search for hosts
	log.Debug("getting hosts")
	num, err = ldapSubordinatesQuery(l, fmt.Sprintf("cn=computers,cn=accounts,%s", suffix), "(objectClass=*)")
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	hostsGauge.WithLabelValues().Set(num)

	// Search for hostgroups
	log.Debug("getting hostgroups")
	num, err = ldapSubordinatesQuery(l, fmt.Sprintf("cn=hostgroups,cn=accounts,%s", suffix), "(objectClass=*)")
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	hostGroupsGauge.WithLabelValues().Set(num)

	// Search for sudo rules
	log.Debug("getting sudo rules")
	num, err = ldapSubordinatesQuery(l, fmt.Sprintf("cn=sudorules,cn=sudo,%s", suffix), "(objectClass=*)")
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	sudoRulesGauge.WithLabelValues().Set(num)

	// Search for hbac rules
	log.Debug("getting hbac rules")
	num, err = ldapCountQuery(l, fmt.Sprintf("cn=hbac,%s", suffix), "(objectClass=ipahbacrule)", "ipaUniqueID", ldap.ScopeSingleLevel)
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	hbacRulesGauge.WithLabelValues().Set(num)

	// Search for dns zones
	log.Debug("getting dns zones")
	num, err = ldapCountQuery(l, fmt.Sprintf("cn=dns,%s", suffix), "(|(objectClass=idnszone)(objectClass=idnsforwardzone))", "idnsName", ldap.ScopeSingleLevel)
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	dnsZonesGauge.WithLabelValues().Set(num)

	// Search for ldap conflicts
	log.Debug("getting ldap conflicts")
	num, err = ldapCountQuery(l, suffix, "(nsds5ReplConflict=*)", "nsds5ReplConflict", ldap.ScopeWholeSubtree)
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	ldapConflictsGauge.WithLabelValues().Set(num)

	// Process ldap replication agreements
	log.Debug("getting replication agreements")
	err = ldapReplicationQuery(l, suffix)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs
}

func ldapSubordinatesQuery(l *ldap.Conn, baseDN, searchFilter string) (float64, error) {
	req := ldap.NewSearchRequest(
		baseDN, ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		searchFilter, []string{"numSubordinates"}, nil,
	)
	sr, err := l.Search(req)
	if err != nil {
		return -1, err
	}

	if len(sr.Entries) == 0 {
		return -1, errors.New(fmt.Sprintf("No entries contain numSubordinates for %s (%s)", baseDN, searchFilter))
	}

	val := sr.Entries[0].GetAttributeValue("numSubordinates")
	num, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return -1, err
	}

	return num, nil
}

func ldapCountQuery(l *ldap.Conn, baseDN, searchFilter, attr string, scope int) (float64, error) {
	req := ldap.NewSearchRequest(
		baseDN, scope, ldap.NeverDerefAliases, 0, 0, false,
		searchFilter, []string{attr}, nil,
	)
	sr, err := l.Search(req)
	if err != nil {
		return -1, err
	}

	num := float64(len(sr.Entries))

	return num, nil
}

func ldapReplicationQuery(l *ldap.Conn, suffix string) error {
	escaped_suffix := ldap_escaper.Replace(suffix)
	base_dn := fmt.Sprintf("cn=replica,cn=%s,cn=mapping tree,cn=config", escaped_suffix)

	req := ldap.NewSearchRequest(
		base_dn, ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=nsds5replicationagreement)", []string{"nsDS5ReplicaHost", "nsds5replicaLastUpdateStatus"}, nil,
	)
	sr, err := l.Search(req)
	if err != nil {
		return err
	}

	for _, entry := range sr.Entries {
		host := entry.GetAttributeValue("nsDS5ReplicaHost")
		status := entry.GetAttributeValue("nsds5replicaLastUpdateStatus")
		if strings.Contains(status, "Incremental update succeeded") { // Error (0) Replica acquired successfully: Incremental update succeeded
			replicationStatusGauge.WithLabelValues(host).Set(1)
		} else if strings.Contains(status, "Problem connecting to replica") { // Error (-1) Problem connecting to replica - LDAP error: Can't contact LDAP server (connection error)
			replicationStatusGauge.WithLabelValues(host).Set(0)
		} else if strings.Contains(status, "Can't acquire busy replica") { // Error (1) Can't acquire busy replica
			// We assume all is ok, so use 1
			replicationStatusGauge.WithLabelValues(host).Set(1)
		} else {
			log.Warnf("Unknown replication status host: %s, status: %s", host, status)
			replicationStatusGauge.WithLabelValues(host).Set(0)
		}
	}

	return nil
}
