module github.com/gnemade360/logicoapp

go 1.24.2

require (
	github.com/gnemade360/go-cli v0.1.0
	github.com/gnemade360/go-config v0.1.3
	github.com/gnemade360/go-logger v0.0.0
)

require gopkg.in/yaml.v3 v3.0.1 // indirect

replace github.com/gnemade360/go-logger => ../go-logger
