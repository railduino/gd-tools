package acme

import (
	"crypto"
	"fmt"
	"os"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/ionos"
	"github.com/go-acme/lego/v4/registration"
)

type IonosUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *IonosUser) GetEmail() string {
	return u.Email
}

func (u *IonosUser) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *IonosUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

// GetIonosCertificate uses lego and the IONOS DNS-01 challenge to create an ACME certificate.
// ACME is short for "Automatic Certificate Management Environment"
func GetIonosCertificate(domains []string, email string, key crypto.PrivateKey) (*certificate.Resource, error) {
	if len(domains) == 0 {
		return nil, fmt.Errorf("missing domains in certificate request")
	}
	if os.Getenv("IONOS_API_KEY") == "" {
		return nil, fmt.Errorf("missing IONOS_API_KEY environment variable")
	}

	user := &IonosUser{
		Email: email,
		key:   key,
	}

	config := lego.NewConfig(user)
	config.CADirURL = lego.LEDirectoryProduction
	config.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to create lego.NewClient: %w", err)
	}

	nameservers := []string{
		"8.8.8.8:53", // Google
		"1.1.1.1:53", // Cloudflare
	}

	dnsProvider, err := ionos.NewDNSProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create ionos.NewDNSProvider: %w", err)
	}

	challengeOpts := []dns01.ChallengeOption{
		dns01.CondOption(true, dns01.AddRecursiveNameservers(nameservers)),
	}

	if err := client.Challenge.SetDNS01Provider(dnsProvider, challengeOpts...); err != nil {
		return nil, fmt.Errorf("failed to set DNS provider: %w", err)
	}

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, fmt.Errorf("failed to register lego client: %w", err)
	}
	user.Registration = reg

	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}

	resource, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain certificate: %w", err)
	}

	return resource, nil
}
