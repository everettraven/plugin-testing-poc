package main

import (
	"fmt"
	"os"

	"github.com/everettraven/plugin-testing-poc/pkg/generator"
	"github.com/everettraven/plugin-testing-poc/pkg/samples"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func main() {
	simpleSample := samples.NewGenericSample(
		samples.WithBinary("/usr/local/bin/operator-sdk"),
		samples.WithDomain("sample.com"),
		samples.WithGvk(schema.GroupVersionKind{
			Group:   "simple",
			Version: "v1alpha1",
			Kind:    "Sample",
		}),
		samples.WithName("simple-sample"),
		samples.WithExtraApiOptions("--resource", "--controller"),
	)

	generator := generator.NewGenericGenerator(
		generator.WithNoWebhook(),
	)

	err := generator.GenerateSamples(simpleSample)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Simple Sample Generated!")

}
