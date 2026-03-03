package email

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"sort"
)

const (
	AccountsName  = "accounts.json"
	DKIM_Selector = "gd-tools"

	SpamBarrier1 = "mx1.spambarrier.de"
	SpamBarrier2 = "mx2.spambarrier.de"
)

type DKIM struct {
	Selector string `json:"selector"`
	CNAME    string `json:"cname"`
	PrivKey  string `json:"priv_key"`
	PubValue string `json:"pub_value"`
}

type MX struct {
	FQDN string `json:"fqdn"`
	Prio int    `json:"prio"`
}

type Domain struct {
	Name    string   `json:"name"`              // The domain name (e.g. example.com)
	DKIMs   []DKIM   `json:"dkims,omitempty"`   // DKIM record value(s)
	DMARC   string   `json:"dmarc"`             // DMARC value (p=none, rua=mailto:<admin>)
	MXs     []MX     `json:"mxs,omitempty"`     // (external) MX records
	Aliases []string `json:"aliases,omitempty"` // alias name(s) - mainly for legacy
	SPFs    []string `json:"spfs,omitempty"`    // SPF additions (ip4:... or include:...)

	SpamBarrier string `json:"spam_barrier,omitempty"` // Verification for SpamBarrier (inbound)
	BrevoCode   string `json:"brevo_code,omitempty"`   // Verification for Brevo (outbound)
	BrevoValid  bool   `json:"brevo_valid,omitempty"`  // Brevo: Verified and Authenticated

	UserList []*User          `json:"users"` // List of all users within the domain
	UserMap  map[string]*User `json:"-"`
}

type DomainList struct {
	Domains []*Domain `json:"domains"`
}

func (dom *Domain) DotName() string {
	return "." + dom.Name
}

func (dom *Domain) NameDot() string {
	return dom.Name + "."
}

func (dom *Domain) DKIM_File() string {
	for _, dkim := range dom.DKIMs {
		if dkim.Selector == DKIM_Selector {
			return dom.Name + "." + dkim.Selector + ".key"
		}
	}
	return dom.Name + "." + DKIM_Selector + ".key"
}

func (dom *Domain) AddDKIM(dkim DKIM) {
	for index, check := range dom.DKIMs {
		if check.Selector == dkim.Selector {
			dom.DKIMs[index] = dkim
			return
		}
	}

	dom.DKIMs = append(dom.DKIMs, dkim)
}

func (dom *Domain) EnsureLocalDKIM() (*DKIM, error) {
	for _, dkim := range dom.DKIMs {
		if dkim.Selector == DKIM_Selector {
			return &dkim, nil
		}
	}

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key for %s: %w", dom.Name, err)
	}
	privData := x509.MarshalPKCS1PrivateKey(privKey)
	privBlk := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privData}
	privValue := string(pem.EncodeToMemory(privBlk))

	pubDER := x509.MarshalPKCS1PublicKey(&privKey.PublicKey)
	pubValue := base64.StdEncoding.EncodeToString(pubDER)

	dkim := DKIM{
		Selector: DKIM_Selector, // "gd-tools"
		CNAME:    "",
		PrivKey:  privValue,
		PubValue: pubValue,
	}
	dom.DKIMs = append(dom.DKIMs, dkim)

	return &dkim, nil
}

func GetDomains(sel map[string]bool) (*DomainList, map[string]*Domain, error) {
	var rawList, domainList DomainList
	domainMap := make(map[string]*Domain)

	content, err := os.ReadFile(AccountsName)
	if err != nil {
		if os.IsNotExist(err) {
			return &rawList, domainMap, nil
		}
		return nil, nil, err
	}

	if err := json.Unmarshal(content, &rawList); err != nil {
		return nil, nil, err
	}

	for index := range rawList.Domains {
		domain := rawList.Domains[index]
		if sel != nil {
			if _, ok := sel[domain.Name]; !ok {
				continue
			}
		}
		domainList.Domains = append(domainList.Domains, domain)
		domainMap[domain.Name] = domain

		domain.UserMap = make(map[string]*User)
		for usrIndex := range domain.UserList {
			user := domain.UserList[usrIndex]
			domain.UserMap[user.Email()] = domain.UserList[usrIndex]
		}
	}

	return &domainList, domainMap, nil
}

func GetDomainSANs() []string {
	sanList := []string{}

	domainList, _, err := GetDomains(nil)
	if err != nil {
		return sanList
	}

	for _, domain := range domainList.Domains {
		sanList = append(sanList,
			"imap."+domain.Name,
			"smtp."+domain.Name,
		)
		for _, alias := range domain.Aliases {
			sanList = append(sanList, alias)
		}
	}

	return sanList
}

func (list *DomainList) GetDKIMs(root string) []string {
	dkimList := []string{}

	for _, domain := range list.Domains {
		dkim, err := domain.EnsureLocalDKIM()
		if err != nil {
			continue
		}
		dkimList = append(dkimList,
			fmt.Sprintf("  %s {", domain.Name),
			fmt.Sprintf(`    selector = "%s";`, dkim.Selector),
			fmt.Sprintf(`    path = "%s/%s";`, root, domain.DKIM_File()),
			fmt.Sprintf(`    sign_authenticated = true;`),
			fmt.Sprintf("  }"),
		)
	}

	return dkimList
}

func (list *DomainList) Save() error {
	for _, domain := range list.Domains {
		sort.Slice(domain.UserList, func(i, j int) bool {
			return domain.UserList[i].Local < domain.UserList[j].Local
		})
	}

	sort.Slice(list.Domains, func(i, j int) bool {
		return list.Domains[i].Name < list.Domains[j].Name
	})

	content, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", AccountsName, err)
	}

	existing, err := os.ReadFile(AccountsName)
	if err == nil && bytes.Equal(existing, content) {
		return nil
	}

	if err := os.WriteFile(AccountsName, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", AccountsName, err)
	}

	return nil
}

func (dom *Domain) AddSpamBarrier() {
	dom.MXs = []MX{
		{FQDN: SpamBarrier1, Prio: 10},
		{FQDN: SpamBarrier2, Prio: 20},
	}
}

func (dom *Domain) AddAlias(alias string) {
	for _, current := range dom.Aliases {
		if current == alias {
			return
		}
	}

	dom.Aliases = append(dom.Aliases, alias)
}

func (dom *Domain) DeleteAlias(alias string) {
	for i, current := range dom.Aliases {
		if current == alias {
			dom.Aliases = append(dom.Aliases[:i], dom.Aliases[i+1:]...)
			break
		}
	}
}

func (dom *Domain) AddSPF(sender string) {
	for _, current := range dom.SPFs {
		if current == sender {
			return
		}
	}

	dom.SPFs = append(dom.SPFs, sender)
}

func (dom *Domain) DeleteSPF(sender string) {
	for i, current := range dom.SPFs {
		if current == sender {
			dom.SPFs = append(dom.SPFs[:i], dom.SPFs[i+1:]...)
			break
		}
	}
}

func (dom *Domain) GetSPF(args ...string) string {
	text := "v=spf1 mx"
	for _, arg := range args {
		text += " " + arg
	}
	for _, inc := range dom.SPFs {
		text += " " + inc
	}
	return text + " -all"
}
