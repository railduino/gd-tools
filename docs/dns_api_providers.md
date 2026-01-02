# DNS-Provider mit API (Alternative zu Hetzner)

Es gibt mehrere **Nameserver-/DNS-Provider**, die eine **REST-API** anbieten, mit der du DNS-Zonen und Records ähnlich wie bei Hetzner automatisiert verwalten kannst.

---

## ⚙️ Alternativen mit DNS‑API

### 1. Cloudflare DNS
- Voll ausgestattete REST‑API für Zonen und DNS‑Records (inkl. TXT, A, CNAME…)
- Häufig genutzt mit lego, acme.sh etc.

### 2. DigitalOcean DNS
- Bietet ebenfalls eine REST‑API für alle Standard‑DNS‑Operationen  
- Unterstützt Automatisierung und Integration in ACME Tools

### 3. DNSimple
- DNS mit klarer REST‑API und Fokus auf Entwicklerfreundlichkeit  
- Lego & acme.sh Integrationen verfügbar

### 4. Google Cloud DNS / AWS Route 53 / Azure DNS
- Große Anbieter, REST‑API für DNS-Management, skalierbar und redundant  
- Klar dokumentiert & weit verbreitet

### 5. OVHcloud / easyDNS / Dyn / NS1 / Vultr / Linode / DNS Made Easy …
- Diese bieten APIs an und werden von ACME-Tools wie lego unterstützt  
- Teilweise kostenlos oder günstige Tarife

### 6. deSEC
- Deutscher, gemeinnütziger DNS‑Provider mit offener API  
- Besonders für DNS-01 Validierung geeignet und kostenlos nutzbar

### 7. ACME‑DNS (Self-hosted)
- Eigenständig gehosteter DNS‑Server für ausschließlich TXT‑Records  
- Ideal für DNS-01 Challenge via CNAME-Delegation  
- API‑basiert und lightweight

---

## 🧠 Stimmen aus der Community

> „I found desec.io. They run an open source stack and offer free DNS hosting up to 15 zones.”  
→ **desec.io** ist eine solide, freie Option mit API-Unterstützung.  

---

## 🔍 Vergleichstabelle

| Anbieter        | API-basiert | LEGO-Support | Kostenlos?     | Lego-Lib kompatibel |
|-----------------|-------------|--------------|----------------|---------------------|
| Cloudflare      | ✔️          | ✔️           | Ja, mit Free‑Plan | ✔️ (`dns/cloudflare`) |
| DigitalOcean    | ✔️          | ✔️           | Ja, eingeschränkt | ✔️ (`dns/digitalocean`) |
| DNSimple        | ✔️          | ✔️           | Kostenpflichtig (ab ca. 5 $/Monat) | ✔️ (`dns/dnsimple`) |
| deSEC           | ✔️          | ✔️           | Ja (non-profit)   | ✔️ (`dns/desec`) |
| Route 53 (AWS)  | ✔️          | ✔️           | Nutzungsabhängig | ✔️ (`dns/route53`) |
| OVH             | ✔️          | ✔️           | Ja (abhängig vom Tarif) | ✔️ (`dns/ovh`) |
| Linode          | ✔️          | ✔️           | Ja, teilweise    | ✔️ (`dns/linode`) |
| NS1             | ✔️          | ✔️           | Kostenpflichtig | ✔️ (`dns/ns1`) |
| ACME-DNS        | ✔️          | ✔️ (über CNAME) | Self-hosted     | ✔️ (`dns/acme-dns`) |

---

## 🧭 Empfehlung

Wenn du eine **leichtgewichtige, kostenfreie und ACME-kompatible Lösung** suchst, empfiehlt sich **deSEC** oder **desec.io**:

- **deSEC**: API-freundlich, speziell für ACME‑DNS-Validierung optimiert.
- **desec.io**: Open‑Source-Projekt, einfach einzurichten, frei bis zu 15 Zonen.

Für mehr Features, Skalierung und höhere Verfügbarkeit sind **Cloudflare**, **DigitalOcean** oder **DNSimple** hervorragende Alternativen mit stabiler API‑Integration.

