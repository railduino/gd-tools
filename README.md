# gd-tools
Toolset zum Verwalten von Docker-Projekten unter Ubuntu Linux

## Systemstruktur und Rechte für `gd-tools`

| Pfad                        | Eigentümer         | Rechte | Beschreibung                                                  |
|-----------------------------|--------------------|--------|---------------------------------------------------------------|
| `/etc/gd-tools-env`         | `root:root`        | `0444` | Betriebsmodus (`prod`/`dev`) – nur lesbar für root            |
| `/etc/gd-tools-config.json` | `root:root`        | `0400` | Systemkonfiguration – streng geschützt                        |
| `/etc/gd-tools/`            | `gd-tools:gd-tools`| `0755` | Projektkonfigurationen (`compose.yaml` etc.) – nur lesbar     |
| `/var/gd-tools/`            | `root:root`        | `0755` | Wurzelverzeichnis für Daten – `gd-tools` hat keinen Schreibzugriff |
| `/var/gd-tools/logs/`       | `gd-tools:gd-tools`| `0755` | Protokolle der Dienste – `gd-tools` kann schreiben            |
| `/var/gd-tools/volumes/`    | `gd-tools:gd-tools`| `0755` | Informationen zu Volumes – `gd-tools` kann schreiben          |
| `/var/gd-tools/lost+found/` | `root:root`        | `0700` | Vom Dateisystem erzeugt – unzugänglich                        |

### Sicherheitsprinzipien

- `gd-tools` kann **keine Systemdateien oder Konfigurationen verändern**.
- Laufzeitzugriff ist beschränkt auf `/etc/gd-tools/` (lesen) und `/var/gd-tools/` (protokollieren und verwalten).
- Die Initialkonfiguration erfolgt ausschließlich durch **root** mittels `gd-tools system`.

