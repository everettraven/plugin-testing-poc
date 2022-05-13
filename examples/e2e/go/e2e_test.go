package e2e_go_test

import (
	"fmt"
	"time"

	e2e_go "github.com/everettraven/plugin-testing-poc/examples/e2e/go"
	"github.com/everettraven/plugin-testing-poc/pkg/e2e"
	"github.com/everettraven/plugin-testing-poc/pkg/kubernetes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const test_dir = "e2e-test"
const image_name = "e2e-test-image:go"

var _ = Describe("e2e", Ordered, func() {

	sample, err := e2e_go.GenerateMemcachedOperator(test_dir, image_name)
	if err != nil {
		Fail("failed to generate sample")
	}

	kctl := kubernetes.NewKubectlUtil(
		kubernetes.WithNamespace(sample.Name()+"-system"),
		kubernetes.WithServiceAccount(sample.Name()+"-controller-manager"),
	)

	metricsClusterRoleBindingName := fmt.Sprintf("%s-metrics-reader", sample.Name())

	Context("Running locally", func() {
		e2e.LocalTest(sample)
	})

	Context("Running on cluster", Ordered, func() {
		BeforeAll(func() {
			By("Installing Prometheus Operator")
			Expect(e2e.InstallPrometheusOperator(kctl)).To(Succeed())

			By("Installing Cert Manager")
			Expect(e2e.InstallCertManagerBundle(false, kctl)).To(Succeed())

			By("Building Operator Image")
			Expect(e2e.BuildOperatorImage(sample, image_name)).To(Succeed())

			By("Checking if running on KinD")
			onKind, err := e2e.IsRunningOnKind(kctl)
			Expect(err).NotTo(HaveOccurred())

			if onKind {
				By("Loading image to KinD")
				Expect(e2e.LoadImageToKindCluster(sample.CommandContext(), image_name)).Should(Succeed())
			}

			By("Deploying Operator")
			Expect(e2e.DeployOperator(sample, image_name)).To(Succeed())
		})

		It("Should run correctly in the cluster", func() {
			By("Checking that the Operator Pod is running")
			controllerUp := func() error {
				return e2e.EnsureOperatorRunning(kctl, 1, "controller-manager", "controller-manager")
			}
			Eventually(controllerUp, 2*time.Minute, time.Second).Should(Succeed())

			// By("Ensuring ServiceMonitor is created for the manager")
			// out, err := kctl.Get(
			// 	true,
			// 	"ServiceMonitor")
			// fmt.Println("OUT --", out)
			// Expect(err).NotTo(HaveOccurred())

			By("Ensuring metrics Service is created for the manager")
			out, err := kctl.Get(
				true,
				"Service",
				fmt.Sprintf("%s-controller-manager-metrics-service", sample.Name()))
			fmt.Println("OUT --", out)
			Expect(err).NotTo(HaveOccurred())
			By("Createing an instance of the CustomResource")
			createResource := func() error {
				return e2e.CreateCustomResource(sample, kctl)
			}
			Eventually(createResource, time.Minute, time.Second).Should(Succeed())

			By("Getting the metrics")
			metrics := e2e.GetMetrics(sample, kctl, metricsClusterRoleBindingName)
			fmt.Println(metrics)
		})

		AfterAll(func() {
			By("Cleaning up metrics")
			Expect(e2e.CleanUpMetrics(kctl, metricsClusterRoleBindingName)).Should(Succeed())

			By("Undeploying the operator")
			Expect(e2e.UndeployOperator(sample)).Should(Succeed())

			By("Uninstalling Cert Manager")
			Expect(e2e.UninstallCertManagerBundle(false, kctl)).Should(Succeed())

			By("Uninstall Prometheus Operator")
			Expect(e2e.UninstallPrometheusOperator(kctl)).Should(Succeed())

			By("Ensuring namespace is deleted")
			_, err := kctl.Wait(false, "namespace", "foo", "--for", "delete", "--timeout", "2m")
			Expect(err).NotTo(HaveOccurred())
		})

	})

	AfterAll(func() {
		// Do cleanup logic
		e2e.CleanUpTestDir(test_dir)
	})
})
