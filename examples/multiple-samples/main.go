package main

import (
	"fmt"
	"os"

	"github.com/everettraven/plugin-testing-poc/pkg/generator"
	"github.com/everettraven/plugin-testing-poc/pkg/samples"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func main() {
	simpleGoSample := samples.NewGenericSample(
		samples.WithBinary("/usr/local/bin/operator-sdk"),
		samples.WithDomain("simple.go.com"),
		samples.WithGvk(schema.GroupVersionKind{
			Group:   "simplego",
			Version: "v1alpha1",
			Kind:    "GoSample",
		}),
		samples.WithName("go-simple-sample"),
		samples.WithExtraApiOptions("--resource", "--controller"),
	)

	simpleHelmSample := samples.NewGenericSample(
		samples.WithBinary("/usr/local/bin/operator-sdk"),
		samples.WithDomain("simple.helm.com"),
		samples.WithGvk(schema.GroupVersionKind{
			Group:   "simplehelm",
			Version: "v1alpha1",
			Kind:    "HelmSample",
		}),
		samples.WithName("helm-simple-sample"),
		samples.WithPlugins("helm"),
	)

	simpleAnsibleSample := samples.NewGenericSample(
		samples.WithBinary("/usr/local/bin/operator-sdk"),
		samples.WithDomain("simple.ansible.com"),
		samples.WithGvk(schema.GroupVersionKind{
			Group:   "simpleansible",
			Version: "v1alpha1",
			Kind:    "AnsibleSample",
		}),
		samples.WithName("ansible-simple-sample"),
		samples.WithPlugins("ansible"),
	)

	generator := generator.NewGenericGenerator(
		generator.WithNoWebhook(),
	)

	err := generator.GenerateSamples(simpleGoSample, simpleHelmSample, simpleAnsibleSample)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Simple Samples Generated!")

}
