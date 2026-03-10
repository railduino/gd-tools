package utils

import (
	"fmt"
	"regexp"
	"strings"
)

/*
Naming scheme for gd-tools:

	host-<fqdn>.conf
	site-<prefix>-<fqdn>.conf

Examples:

	host-host00.example.com.conf
	site-wp-blog.example.com.conf
	site-nc-cloud.example.com.conf
*/

const (
	PrefixAdmidio     = "ad"
	PrefixBookStack   = "bk"
	PrefixFirefly     = "fi"
	PrefixImmich      = "im"
	PrefixMediaWiki   = "mw"
	PrefixNextcloud   = "nc"
	PrefixOCIS        = "oc"
	PrefixPaperless   = "pl"
	PrefixRoundcube   = "rc"
	PrefixRustDesk    = "rd"
	PrefixUptimeKuma  = "uk"
	PrefixVaultwarden = "vw"
	PrefixWordPress   = "wp"
)

/*
SiteType describes static metadata for a site class.
*/
type SiteType struct {
	Prefix string
	Name   string

	NeedsPHP     bool
	NeedsMariaDB bool
	NeedsTLS     bool
}

/*
Registry of known site types.
*/
var SiteTypes = map[string]SiteType{
	PrefixWordPress: {
		Prefix:       PrefixWordPress,
		Name:         "wordpress",
		NeedsPHP:     true,
		NeedsMariaDB: true,
		NeedsTLS:     true,
	},
	PrefixNextcloud: {
		Prefix:       PrefixNextcloud,
		Name:         "nextcloud",
		NeedsPHP:     true,
		NeedsMariaDB: true,
		NeedsTLS:     true,
	},
	PrefixRustDesk: {
		Prefix:       PrefixRustDesk,
		Name:         "rustdesk",
		NeedsPHP:     false,
		NeedsMariaDB: false,
		NeedsTLS:     true,
	},
	PrefixRoundcube: {
		Prefix:       PrefixRoundcube,
		Name:         "roundcube",
		NeedsPHP:     true,
		NeedsMariaDB: true,
		NeedsTLS:     true,
	},
}

var dbNameSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

/*
LookupSiteType returns the SiteType for a given prefix.
*/
func LookupSiteType(prefix string) (SiteType, error) {
	st, ok := SiteTypes[prefix]
	if !ok {
		return SiteType{}, fmt.Errorf("unknown site type: %s", prefix)
	}
	return st, nil
}

/*
IsKnownPrefix checks if a prefix is supported.
*/
func IsKnownPrefix(prefix string) bool {
	_, ok := SiteTypes[prefix]
	return ok
}

/*
HostConfName returns the apache config filename for the canonical host.

Example:

	host-host00.example.com.conf
*/
func HostConfName(fqdn string) string {
	return fmt.Sprintf("host-%s.conf", fqdn)
}

/*
SiteConfName returns the apache config filename for a site.

Example:

	site-wp-blog.example.com.conf
*/
func SiteConfName(prefix, fqdn string) string {
	return fmt.Sprintf("site-%s-%s.conf", prefix, fqdn)
}

/*
DBName converts an arbitrary name into a valid database name.

Rules:
- lowercase
- replace non [a-z0-9] with "_"
- must not start with a digit
- maximum length 64 characters
*/
func DBName(name string) string {
	name = strings.ToLower(name)
	name = dbNameSanitizer.ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")

	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = "db_" + name
	}

	if len(name) > 64 {
		name = name[:64]
	}

	return name
}

/*
SiteKey returns a stable internal key for a site.

Example:

	SiteKey("wp", "blog.example.com") -> "wp_blog_example_com"
	SiteKey("nc", "cloud.example.com") -> "nc_cloud_example_com"
*/
func SiteKey(prefix, fqdn string) string {
	return DBName(prefix + "_" + fqdn)
}

/*
DBUserName returns a stable database user name for a site.
*/
func DBUserName(prefix, fqdn string) string {
	return SiteKey(prefix, fqdn)
}

/*
DBPrefix returns a stable table prefix for a site.

Example:

	wp_blog_example_com_
*/
func DBPrefix(prefix, fqdn string) string {
	return SiteKey(prefix, fqdn) + "_"
}
