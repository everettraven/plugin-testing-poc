package e2e

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/everettraven/plugin-testing-poc/pkg/command"
	"github.com/everettraven/plugin-testing-poc/pkg/kubernetes"
	"github.com/everettraven/plugin-testing-poc/pkg/samples"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kbutil "sigs.k8s.io/kubebuilder/v3/pkg/plugin/util"
)

func LocalTest(sample samples.Sample) {
	BeforeEach(func() {
		By("Installing CRD's")
		cmd := exec.Command("make", "install")
		cmdCtx := sample.CommandContext()
		_, err := cmdCtx.Run(cmd, sample.Name())
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		By("Uninstalling CRD's")
		cmd := exec.Command("make", "uninstall")
		cmdCtx := sample.CommandContext()
		_, err := cmdCtx.Run(cmd, sample.Name())
		Expect(err).NotTo(HaveOccurred())
	})

	It("Should run correctly when run locally", func() {
		By("Running the project")
		cmd := exec.Command("make", "run")
		err := cmd.Start()
		Expect(err).NotTo(HaveOccurred())

		By("Killing the project")
		err = cmd.Process.Kill()
		Expect(err).NotTo(HaveOccurred())
	})
}

func BuildOperatorImage(sample samples.Sample, image string) error {
	cmd := exec.Command("make", "docker-build", "IMG="+image)
	_, err := sample.CommandContext().Run(cmd, sample.Name())
	if err != nil {
		fmt.Errorf("encountered an error when building the operator image: %w", err)
	}

	return nil
}

func DeployOperator(sample samples.Sample, image string) error {
	cmd := exec.Command("make", "deploy", "IMG="+image)
	_, err := sample.CommandContext().Run(cmd, sample.Name())
	if err != nil {
		fmt.Errorf("encountered an error when deploying the operator: %w", err)
	}

	return nil
}

func UndeployOperator(sample samples.Sample) error {
	cmd := exec.Command("make", "undeploy")
	_, err := sample.CommandContext().Run(cmd, sample.Name())
	if err != nil {
		fmt.Errorf("encountered an error when undeploying the operator: %w", err)
	}

	return nil
}

func InstallPrometheusOperator(kubectl kubernetes.Kubectl) error {
	url, err := getPrometheusOperatorUrl(kubectl)
	if err != nil {
		return fmt.Errorf("encountered an error when getting the bundle URL: %w", err)
	}

	_, err = kubectl.Apply(false, "-f", url)
	if err != nil {
		return fmt.Errorf("encountered an error when getting the bundle URL: %w", err)
	}

	return nil
}

func UninstallPrometheusOperator(kubectl kubernetes.Kubectl) error {
	url, err := getPrometheusOperatorUrl(kubectl)
	if err != nil {
		return fmt.Errorf("encountered an error when getting the bundle URL: %w", err)
	}
	_, err = kubectl.Apply(false, "-f", url)
	if err != nil {
		return fmt.Errorf("encountered an error when deleting the bundle: %w", err)
	}

	return nil
}

func getPrometheusOperatorUrl(kubectl kubernetes.Kubectl) (string, error) {
	prometheusOperatorLegacyVersion := "0.33"
	prometheusOperatorLegacyURL := "https://raw.githubusercontent.com/coreos/prometheus-operator/release-%s/bundle.yaml"
	prometheusOperatorVersion := "0.51"
	prometheusOperatorURL := "https://raw.githubusercontent.com/prometheus-operator/" +
		"prometheus-operator/release-%s/bundle.yaml"

	var url string

	kubeVersion, err := kubectl.Version()
	if err != nil {
		return "", fmt.Errorf("encountered an error trying to get Kubernetes Version: %w", err)
	}
	serverMajor, err := strconv.ParseUint(kubeVersion.ServerVersion().Major(), 10, 64)
	if err != nil {
		return "", fmt.Errorf("encountered an error trying to parse Kubernetes Major Version: %w", err)
	}

	serverMinor, err := strconv.ParseUint(kubeVersion.ServerVersion().Minor(), 10, 64)
	if err != nil {
		return "", fmt.Errorf("encountered an error trying to parse Kubernetes Minor Version: %w", err)
	}

	if serverMajor <= 1 && serverMinor < 16 {
		url = fmt.Sprintf(prometheusOperatorLegacyURL, prometheusOperatorLegacyVersion)
	} else {
		url = fmt.Sprintf(prometheusOperatorURL, prometheusOperatorVersion)
	}

	return url, nil
}

func InstallCertManagerBundle(hasv1beta1CRs bool, kubectl kubernetes.Kubectl) error {
	url, err := getCertManagerURL(hasv1beta1CRs, kubectl)
	if err != nil {
		return fmt.Errorf("encountered an error when getting the bundle URL: %w", err)
	}

	_, err = kubectl.Apply(false, "-f", url, "--validate=false")
	if err != nil {
		return fmt.Errorf("encountered an error when applying the bundle: %w", err)
	}
	// Wait for cert-manager-webhook to be ready, which can take time if cert-manager
	// was re-installed after uninstalling on a cluster.
	_, err = kubectl.Wait(false, "deployment.apps/cert-manager-webhook",
		"--for", "condition=Available",
		"--namespace", "cert-manager",
		"--timeout", "5m",
	)
	if err != nil {
		return fmt.Errorf("encountered an error when waiting for the webhook to be ready: %w", err)
	}

	return nil
}

func UninstallCertManagerBundle(hasv1beta1CRs bool, kubectl kubernetes.Kubectl) error {
	url, err := getCertManagerURL(hasv1beta1CRs, kubectl)
	if err != nil {
		return fmt.Errorf("encountered an error when getting the bundle URL: %w", err)
	}

	_, err = kubectl.Delete(false, "-f", url)
	if err != nil {
		return fmt.Errorf("encountered an error when deleting the bundle: %w", err)
	}

	return nil
}

func getCertManagerURL(hasv1beta1CRs bool, kubectl kubernetes.Kubectl) (string, error) {
	certmanagerVersionWithv1beta2CRs := "v0.11.0"
	certmanagerLegacyVersion := "v1.0.4"
	certmanagerVersion := "v1.5.3"

	certmanagerURLTmplLegacy := "https://github.com/jetstack/cert-manager/releases/download/%s/cert-manager-legacy.yaml"
	certmanagerURLTmpl := "https://github.com/jetstack/cert-manager/releases/download/%s/cert-manager.yaml"
	// Return a URL for the manifest bundle with v1beta1 CRs.

	kubeVersion, err := kubectl.Version()
	if err != nil {
		return "", fmt.Errorf("encountered an error trying to get Kubernetes Version: %w", err)
	}
	serverMajor, err := strconv.ParseUint(kubeVersion.ServerVersion().Major(), 10, 64)
	if err != nil {
		return "", fmt.Errorf("encountered an error trying to parse Kubernetes Major Version: %w", err)
	}

	serverMinor, err := strconv.ParseUint(kubeVersion.ServerVersion().Minor(), 10, 64)
	if err != nil {
		return "", fmt.Errorf("encountered an error trying to parse Kubernetes Minor Version: %w", err)
	}

	if hasv1beta1CRs {
		return fmt.Sprintf(certmanagerURLTmpl, certmanagerVersionWithv1beta2CRs), nil
	}

	// Determine which URL to use for a manifest bundle with v1 CRs.
	// The most up-to-date bundle uses v1 CRDs, which were introduced in k8s v1.16.
	if serverMajor <= 1 && serverMinor < 16 {
		return fmt.Sprintf(certmanagerURLTmplLegacy, certmanagerLegacyVersion), nil
	}
	return fmt.Sprintf(certmanagerURLTmpl, certmanagerVersion), nil
}

func CleanUpTestDir(path string) {
	err := os.RemoveAll(path)
	Expect(err).NotTo(HaveOccurred())
}

func GetMetrics(sample samples.Sample, kubectl kubernetes.Kubectl, metricsClusterRoleBindingName string) string {
	By("granting permissions to access the metrics and read the token")
	out, err := kubectl.Command("create", "clusterrolebinding", metricsClusterRoleBindingName,
		fmt.Sprintf("--clusterrole=%s-metrics-reader", sample.Name()),
		fmt.Sprintf("--serviceaccount=%s:%s", kubectl.Namespace(), kubectl.ServiceAccount()))
	fmt.Println("OUT --", out)
	Expect(err).NotTo(HaveOccurred())

	By("reading the metrics token")
	// Filter token query by service account in case more than one exists in a namespace.
	query := fmt.Sprintf(`{.items[?(@.metadata.annotations.kubernetes\.io/service-account\.name=="%s")].data.token}`,
		kubectl.ServiceAccount(),
	)
	out, err = kubectl.Get(true, "secrets")
	fmt.Println("OUT --", out)
	b64Token, err := kubectl.Get(true, "secrets", "-o=jsonpath="+query)
	fmt.Println("OUT--", b64Token)
	Expect(err).NotTo(HaveOccurred())
	token, err := base64.StdEncoding.DecodeString(strings.TrimSpace(b64Token))
	Expect(err).NotTo(HaveOccurred())
	Expect(len(token)).To(BeNumerically(">", 0))

	By("creating a curl pod")
	cmdOpts := []string{
		"run", "curl", "--image=curlimages/curl:7.68.0", "--restart=OnFailure", "--",
		"curl", "-v", "-k", "-H", fmt.Sprintf(`Authorization: Bearer %s`, token),
		fmt.Sprintf("https://%s-controller-manager-metrics-service.%s.svc:8443/metrics", sample.Name(), kubectl.Namespace()),
	}
	out, err = kubectl.CommandInNamespace(cmdOpts...)
	fmt.Println("OUT --", out)
	Expect(err).NotTo(HaveOccurred())

	By("validating that the curl pod is running as expected")
	verifyCurlUp := func() error {
		// Validate pod status
		status, err := kubectl.Get(
			true,
			"pods", "curl", "-o", "jsonpath={.status.phase}")
		if err != nil {
			return err
		}
		if status != "Completed" && status != "Succeeded" {
			return fmt.Errorf("curl pod in %s status", status)
		}
		return nil
	}
	Eventually(verifyCurlUp, 2*time.Minute, time.Second).Should(Succeed())

	By("validating that the metrics endpoint is serving as expected")
	var metricsOutput string
	getCurlLogs := func() string {
		metricsOutput, err = kubectl.Logs(true, "curl")
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		return metricsOutput
	}
	Eventually(getCurlLogs, time.Minute, time.Second).Should(ContainSubstring("< HTTP/2 200"))

	return metricsOutput
}

func CleanUpMetrics(kubectl kubernetes.Kubectl, metricsClusterRoleBindingName string) error {
	_, err := kubectl.Delete(true, "pod", "curl")
	if err != nil {
		return fmt.Errorf("encountered an error when deleting the metrics pod: %w", err)
	}

	_, err = kubectl.Delete(false, "clusterrolebinding", metricsClusterRoleBindingName)
	if err != nil {
		return fmt.Errorf("encountered an error when deleting the metrics clusterrolebinding: %w", err)
	}

	return nil
}

func CreateCustomResource(sample samples.Sample, kubectl kubernetes.Kubectl) error {
	sampleFile := filepath.Join(sample.CommandContext().Dir(),
		sample.Name(),
		"config",
		"samples",
		fmt.Sprintf("%s_%s_%s.yaml", sample.GVK().Group, sample.GVK().Version, strings.ToLower(sample.GVK().Kind)))

	_, err := kubectl.Apply(true, "-f", sampleFile)
	return err
}

func EnsureOperatorRunning(kubectl kubernetes.Kubectl, expectedNumPods int, podNameShouldContain string, controlPlane string) error {
	// Get the controller-manager pod name
	podOutput, err := kubectl.Get(
		true,
		"pods", "-l", "control-plane="+controlPlane,
		"-o", "go-template={{ range .items }}{{ if not .metadata.deletionTimestamp }}{{ .metadata.name }}"+
			"{{ \"\\n\" }}{{ end }}{{ end }}")
	if err != nil {
		return fmt.Errorf("could not get pods: %v", err)
	}
	podNames := kbutil.GetNonEmptyLines(podOutput)
	if len(podNames) != expectedNumPods {
		return fmt.Errorf("expecting %d pod(s), have %d", expectedNumPods, len(podNames))
	}
	controllerPodName := podNames[0]
	if !strings.Contains(controllerPodName, podNameShouldContain) {
		return fmt.Errorf("expecting pod name %q to contain %q", controllerPodName, podNameShouldContain)
	}

	// Ensure the controller-manager Pod is running.
	status, err := kubectl.Get(
		true,
		"pods", controllerPodName, "-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to get pod status for %q: %v", controllerPodName, err)
	}
	if status != "Running" {
		return fmt.Errorf("controller pod in %s status", status)
	}
	return nil
}

func IsRunningOnKind(kubectl kubernetes.Kubectl) (bool, error) {
	kubectx, err := kubectl.Command("config", "current-context")
	if err != nil {
		return false, err
	}
	return strings.Contains(kubectx, "kind"), nil
}

func LoadImageToKindCluster(cc command.CommandContext, image string) error {
	cluster := "kind"
	if v, ok := os.LookupEnv("KIND_CLUSTER"); ok {
		cluster = v
	}
	kindOptions := []string{"load", "docker-image", image, "--name", cluster}
	cmd := exec.Command("kind", kindOptions...)
	_, err := cc.Run(cmd)
	return err
}
