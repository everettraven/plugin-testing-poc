module github.com/everettraven/plugin-testing-poc

go 1.17

require k8s.io/apimachinery v0.24.0

require (
	github.com/gobuffalo/flect v0.2.3 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sys v0.0.0-20220503163025-988cb79eb6c6 // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/onsi/ginkgo/v2 v2.1.4
	github.com/onsi/gomega v1.19.0
	sigs.k8s.io/kubebuilder/v3 v3.4.1
)
