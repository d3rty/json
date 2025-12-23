module github.com/d3rty/json

go 1.25

require (
	github.com/BurntSushi/toml v1.6.0
	github.com/amberpixels/k1 v0.1.4 // latest
	github.com/amberpixels/years v0.0.7 // latest
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/djherbis/times v1.6.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

//replace github.com/amberpixels/years => ../years
//replace github.com/amberpixels/k1  => ../k1
