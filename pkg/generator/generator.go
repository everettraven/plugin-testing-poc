package generator

import (
	"fmt"

	"github.com/everettraven/plugin-testing-poc/pkg/samples"
)

type GenericGenerator struct {
	init    bool
	api     bool
	webhook bool
}

type GenericGeneratorOptions func(gg *GenericGenerator)

func WithNoInit() GenericGeneratorOptions {
	return func(gg *GenericGenerator) {
		gg.init = false
	}
}

func WithNoApi() GenericGeneratorOptions {
	return func(gg *GenericGenerator) {
		gg.api = false
	}
}

func WithNoWebhook() GenericGeneratorOptions {
	return func(gg *GenericGenerator) {
		gg.webhook = false
	}
}

func NewGenericGenerator(opts ...GenericGeneratorOptions) *GenericGenerator {
	gg := &GenericGenerator{
		init:    true,
		api:     true,
		webhook: true,
	}

	for _, opt := range opts {
		opt(gg)
	}

	return gg
}

func (gg *GenericGenerator) GenerateSamples(samples ...samples.Sample) error {
	for _, sample := range samples {
		fmt.Println("scaffolding sample: ", sample.Name())
		if gg.init {
			err := sample.GenerateInit()
			if err != nil {
				return fmt.Errorf("error in init generation for sample %s: %w", sample.Name(), err)
			}
		}

		if gg.api {
			err := sample.GenerateApi()
			if err != nil {
				return fmt.Errorf("error in api generation for sample %s: %w", sample.Name(), err)
			}
		}

		if gg.webhook {
			err := sample.GenerateWebhook()
			if err != nil {
				return fmt.Errorf("error in webhook generation for sample %s: %w", sample.Name(), err)
			}
		}
	}

	return nil
}
