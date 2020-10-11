# Configor

Golang Configuration module that support YAML, JSON, Shell Environment

This is based on [jinzhu/configor's](https://github.com/jinzhu/configor) and [sherifabdlnaby/configuro's](https://github.com/sherifabdlnaby/configuro) work, with some bug fixes and enhancements. 

## Features

- Strongly typed config with tags
- Reflection based config validation, for syntax amd examples refer [govalidator](https://github.com/asaskevich/govalidator)
    - Required fields
    - Optional fields
    - Enum fields
    - Min, Max, email, phone etc
- Setting defaults for fields not in the config files. for syntax amd examples refer [creasty's defaults](https://github.com/creasty/defaults)
- Config Sources
    - YAML files
    - Environment Variables
    - [ ] Environment Variables Expanding
    - Command line flags
    - [ ] Kubernetes ConfigMaps
    - Merge multiple config sources (Overlays)
    - Detect Runtime Environment (test, development, production), auto merge overlay files
- Dynamic Configuration Management (Hot Reconfiguration)
    - Remote config push
    - Externalized configuration
    - Live component reloading / zero-downtime    
    - [ ] Observe Config [Changes](https://play.golang.org/p/41ygGZ-QaB https://gist.github.com/patrickmn/1549985)
- Support Embed config files in Go binaries via [pkger](https://github.com/markbates/pkger)

```golang
type Item struct {
    Name int `yaml:"full_name,omitempty"`
    Age int  `yaml:",omitempty"` //  Removing Empty JSON Fields
    City string `yaml:",omitempty"`
    TLS      bool   `default:"true" yaml:",omitempty"` // Use default when Empty
    Password string `yaml:"-"` // Ignoring Private Fields
    Name     string `valid:"-"`
    Title    string `valid:"alphanum,required"`
    AuthorIP string `valid:"ipv4"`
    Email    string `valid:"email"`
}
```

## Usage

```go
package main

import (
	"fmt"
	"github.com/xmlking/configor"
)

var Config = struct {
	APPName string `default:"app name" yaml:",omitempty"`

	DB struct {
		Name     string
		User     string `default:"root" yaml:",omitempty"`
		Password string `required:"true" env:"DBPassword"`
		Port     uint   `default:"3306" yaml:",omitempty"`
	}

	Contacts []struct {
		Name  string
		Email string `required:"true"`
	}
}{}

func main() {
	configor.Load(&Config, "config.yml")
	fmt.Printf("config: %#v", Config)
}
```

With configuration file *config.yml*:

```yaml
appname: test

db:
    name:     test
    user:     test
    password: test
    port:     1234

contacts:
- name: i test
  email: test@test.com
```

## Debug Mode & Verbose Mode

Debug/Verbose mode is helpful when debuging your application, `debug mode` will let you know how `configor` loaded your configurations, like from which file, shell env, `verbose mode` will tell you even more, like those shell environments `configor` tried to load.

```go
// Enable debug mode or set env `CONFIGOR_DEBUG_MODE` to true when running your application
configor.New(&configor.Config{Debug: true}).Load(&Config, "config.yaml")

// Enable verbose mode or set env `CONFIGOR_VERBOSE_MODE` to true when running your application
configor.New(&configor.Config{Verbose: true}).Load(&Config, "config.yaml")

// You can create custom Configor once and reuse to load multiple different configs  
Configor := configor.New(&configor.Config{Debug: true})
Configor.Load(&Config2, "config2.yaml")
Configor.Load(&Config3, "config3.yaml")
```

## Load

# Advanced Usage

* Load mutiple configurations

```go
// Earlier configurations have higher priority
configor.Load(&Config, "application.yml", "database.json")
```

* Return error on unmatched keys

Return an error on finding keys in the config file that do not match any fields in the config struct.
In the example below, an error will be returned if config.toml contains keys that do not match any fields in the ConfigStruct struct.
If ErrorOnUnmatchedKeys is not set, it defaults to false.

Note that for json files, setting ErrorOnUnmatchedKeys to true will have an effect only if using go 1.10 or later.

```go
err := configor.New(&configor.Config{ErrorOnUnmatchedKeys: true}).Load(&ConfigStruct, "config.toml")
```

* Load configuration by environment

Use `CONFIGOR_ENV` to set environment, if `CONFIGOR_ENV` not set, environment will be `development` by default, and it will be `test` when running tests with `go test`

```go
// config.go
configor.Load(&Config, "config.json")

$ go run config.go
// Will load `config.json`, `config.development.json` if it exists
// `config.development.json` will overwrite `config.json`'s configuration
// You could use this to share same configuration across different environments

$ CONFIGOR_ENV=production go run config.go
// Will load `config.json`, `config.production.json` if it exists
// `config.production.json` will overwrite `config.json`'s configuration

$ go test
// Will load `config.json`, `config.test.json` if it exists
// `config.test.json` will overwrite `config.json`'s configuration

$ CONFIGOR_ENV=production go test
// Will load `config.json`, `config.production.json` if it exists
// `config.production.json` will overwrite `config.json`'s configuration
```

```go
// Set environment by config
configor.New(&configor.Config{Environment: "production"}).Load(&Config, "config.json")
```

* Example Configuration

```go
// config.go
configor.Load(&Config, "config.yml")

$ go run config.go
// Will load `config.example.yml` automatically if `config.yml` not found and print warning message
```

* Load files Via [Pkger](https://github.com/markbates/pkger)

> Enable Pkger or set via env `CONFIGOR_VERBOSE_MODE` to true to use Pkger for loading files

```go
// config.go
configor.New(&configor.Config{UsePkger: true}).Load(&Config, "/config/config.json")
# or set via Environment Variable 
$ CONFIGOR_USE_PKGER=true  go run config.go
```

* Load From Shell Environment

Struct field names will be converted to **UpperSnakeCase**
```go
$ CONFIGOR_APP_NAME="hello world" CONFIGOR_DB_NAME="hello world" go run config.go
// Load configuration from shell environment, it's name is {{prefix}}_FieldName
```

```go
// You could overwrite the prefix with environment CONFIGOR_ENV_PREFIX, for example:
$ CONFIGOR_ENV_PREFIX="WEB" WEB_APP_NAME="hello world" WEB_DB_NAME="hello world" go run config.go

// Set prefix by config
configor.New(&configor.Config{ENVPrefix: "WEB"}).Load(&Config, "config.json")
```

* Anonymous Struct

Add the `anonymous:"true"` tag to an anonymous, embedded struct to NOT include the struct name in the environment
variable of any contained fields.  For example:

```go
type Details struct {
	Description string
}

type Config struct {
	Details `anonymous:"true"`
}
```

With the `anonymous:"true"` tag specified, the environment variable for the `Description` field is `CONFIGOR_DESCRIPTION`.
Without the `anonymous:"true"`tag specified, then environment variable would include the embedded struct name and be `CONFIGOR_DETAILS_DESCRIPTION`.

* With flags

```go
func main() {
	config := flag.String("file", "config.yml", "configuration file")
	flag.StringVar(&Config.APPName, "name", "", "app name")
	flag.StringVar(&Config.DB.Name, "db-name", "", "database name")
	flag.StringVar(&Config.DB.User, "db-user", "root", "database user")
	flag.Parse()

	os.Setenv("CONFIGOR_ENV_PREFIX", "-")
	configor.Load(&Config, *config)
	// configor.Load(&Config) // only load configurations from shell env & flag
}
```

## Gotchas
- Defaults not initialized for `Map` type fields
- Overlaying (merging) not working for `Map` type fields

## TODO
- use [mergo](https://github.com/imdario/mergo) to merge to fix Overlaying `Map` type fields
- Adopt Environment Variables Expanding from [sherifabdlnaby/configuro](https://github.com/sherifabdlnaby/configuro)