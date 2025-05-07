module github.com/railduino/gd-tools

go 1.23.5

toolchain go1.23.8

replace github.com/docker/docker => github.com/docker/docker v24.0.7+incompatible

replace github.com/docker/distribution => github.com/docker/distribution v2.7.1+incompatible

require (
	github.com/docker/docker v0.0.0-00010101000000-000000000000
	github.com/golang-jwt/jwt/v4 v4.5.2
	github.com/joho/godotenv v1.5.1
	github.com/leonelquinteros/gotext v1.7.1
	github.com/urfave/cli/v2 v2.27.6
	golang.org/x/crypto v0.37.0
	golang.org/x/exp v0.0.0-20250408133849-7e4ce0ab07d0
	golang.org/x/net v0.21.0
	golang.org/x/text v0.24.0
	gopkg.in/ini.v1 v1.67.0
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/driver/sqlite v1.5.7
	gorm.io/gorm v1.26.0
)

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.5 // indirect
	github.com/docker/distribution v0.0.0-00010101000000-000000000000 // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	gotest.tools/v3 v3.5.2 // indirect
)
