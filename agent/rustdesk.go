package agent

// RustDesk is the request/response payload section for managing hbbs/hbbr server keys.
// - If PrivateKeyB64 is present in the request, prod restores the keys.
// - If PrivateKeyB64 is empty and prod has keys, prod exports them in the response.
// - If keys do not exist on prod, prod boots hbbs once (or starts service) to generate them, then exports.
type RustDesk struct {
	// The host name when sent to the production server
	HostName   string `json:"host_name,omitempty"`
	DomainName string `json:"domain_name,omitempty"`

	// Keys are stored as raw file bytes encoded in base64.
	// Private key must never be logged.
	PrivateKeyB64 string `json:"private_key_b64,omitempty"`
	PublicKeyB64  string `json:"public_key_b64,omitempty"`

	// Optional metadata for safety/diagnostics.
	// Fingerprint should be derived from the public key (e.g. sha256) and is safe to log/display.
	Fingerprint string `json:"fingerprint,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"` // RFC3339, optional

	// Safety switch: allow replacing existing keys on prod (default false).
	ForceReplace bool `json:"force_replace,omitempty"`
}

func (rd *RustDesk) FQDN() string {
	return rd.HostName + "." + rd.DomainName
}

func RustDeskTest(req *Request) bool {
	return req != nil && req.RustDesk != nil
}
