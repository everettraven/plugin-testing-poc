package e2e_go_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestE2EGo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite for Go")
}
