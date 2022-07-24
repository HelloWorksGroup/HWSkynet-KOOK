module github.com/Nigh/YUI-KHL

go 1.18

require (
	github.com/adrg/strutil v0.3.0
	github.com/jpillora/overseer v1.1.6
	github.com/lithammer/fuzzysearch v1.1.5
	github.com/lonelyevil/khl v0.0.27-0.20220429093902-17fb9d330fc2
	github.com/lonelyevil/khl/log_adapter/plog v0.0.27-0.20220429093902-17fb9d330fc2
	github.com/nanobox-io/golang-scribble v0.0.0-20190309225732-aa3e7c118975
	github.com/phuslu/log v1.0.77
	github.com/spf13/viper v1.12.0
	local/khlcard v0.0.0-00010101000000-000000000000
)

require (
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/bits-and-blooms/bitset v1.2.2 // indirect
	github.com/bits-and-blooms/bloom/v3 v3.2.0 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jcelliott/lumber v0.0.0-20160324203708-dd349441af25 // indirect
	github.com/jpillora/s3 v1.1.4 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.1 // indirect
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.0 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/ini.v1 v1.66.6 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace local/khlcard => ./kcard
