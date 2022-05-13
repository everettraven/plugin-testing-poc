package e2e_go

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/everettraven/plugin-testing-poc/pkg/command"
	"github.com/everettraven/plugin-testing-poc/pkg/generator"
	"github.com/everettraven/plugin-testing-poc/pkg/samples"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kbutil "sigs.k8s.io/kubebuilder/v3/pkg/plugin/util"
)

func GenerateMemcachedOperator(dir string, image string) (samples.Sample, error) {
	sample := samples.NewGenericSample(
		samples.WithBinary("/usr/local/bin/operator-sdk"),
		samples.WithDomain("example.com"),
		samples.WithExtraApiOptions("--resource", "--controller"),
		samples.WithExtraWebhookOptions("--defaulting"),
		samples.WithGvk(schema.GroupVersionKind{
			Group:   "cache",
			Version: "v1alpha1",
			Kind:    "Memcached",
		}),
		samples.WithName("memcached-operator"),
		samples.WithCommandContext(
			command.NewGenericCommandContext(
				command.WithDir(dir),
			),
		),
	)

	// Generate the sample so it is populated for testing locally
	gen := generator.NewGenericGenerator(
		generator.WithNoWebhook(),
	)

	err := gen.GenerateSamples(sample)
	if err != nil {
		return nil, fmt.Errorf("encountered an error when scaffolding the sample: %w", err)
	}

	err = implementSampleLogic(sample, image)
	if err != nil {
		return nil, fmt.Errorf("encountered an error when implementing the sample: %w", err)
	}

	return sample, nil
}

func implementSampleLogic(sample samples.Sample, image string) error {
	dir := sample.CommandContext().Dir() + "/" + sample.Name()
	err := implementApi(dir, sample.GVK())
	if err != nil {
		fmt.Errorf("encountered an error implementing the api: %w", err)
	}

	err = implementController(dir, sample.GVK())
	if err != nil {
		fmt.Errorf("encountered an error implementing the controller: %w", err)
	}

	// err = implementWebhook(dir, sample.GVK())
	// if err != nil {
	// 	fmt.Errorf("encountered an error implementing the webhook: %w", err)
	// }

	// err = uncommentDefaultKustomization(dir)
	// if err != nil {
	// 	fmt.Errorf("encountered an error uncommenting default kustomization: %w", err)
	// }

	// err = uncommentManifestsKustomization(dir)
	// if err != nil {
	// 	fmt.Errorf("encountered an error uncommenting manifests kustomization: %w", err)
	// }

	cmd := exec.Command("go", "mod", "tidy")
	_, err = sample.CommandContext().Run(cmd, sample.Name())
	if err != nil {
		fmt.Errorf("encountered an error running go mod tidy: %w", err)
	}

	err = generateBundle(sample, image)
	if err != nil {
		fmt.Errorf("encountered an error creating the bundle: %w", err)
	}

	err = stripBundleAnnotations(sample)
	if err != nil {
		fmt.Errorf("encountered an error stripping bundle annotations: %w", err)
	}

	cmd = exec.Command("make", "fmt")
	_, err = sample.CommandContext().Run(cmd)
	if err != nil {
		fmt.Errorf("encountered an error formatting project: %w", err)
	}

	// Clean up built binaries, if any.
	err = os.RemoveAll(filepath.Join(sample.CommandContext().Dir(), sample.Name(), "bin"))
	if err != nil {
		fmt.Errorf("encountered an error cleaning up binaries: %w", err)
	}

	return nil
}

func implementApi(dir string, gvk schema.GroupVersionKind) error {
	err := kbutil.InsertCode(
		filepath.Join(dir, "api", gvk.Version, fmt.Sprintf("%s_types.go", strings.ToLower(gvk.Kind))),
		fmt.Sprintf("type %sSpec struct {\n\t// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster\n\t// Important: Run \"make\" to regenerate code after modifying this file", gvk.Kind),
		`
	// Size defines the number of Memcached instances
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Size int32 `+"`"+`json:"size,omitempty"`+"`"+`
`)

	if err != nil {
		return fmt.Errorf("encountered an error inserting spec Status: %w", err)
	}

	err = kbutil.InsertCode(
		filepath.Join(dir, "api", gvk.Version, fmt.Sprintf("%s_types.go", strings.ToLower(gvk.Kind))),
		fmt.Sprintf("type %sStatus struct {\n\t// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster\n\t// Important: Run \"make\" to regenerate code after modifying this file", gvk.Kind),
		`
	// Nodes store the name of the pods which are running Memcached instances
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Nodes []string `+"`"+`json:"nodes,omitempty"`+"`"+`
`)

	if err != nil {
		return fmt.Errorf("encountered an error implementing Status: %w", err)
	}

	// Add CSV marker that shows CRD owned resources
	err = kbutil.InsertCode(
		filepath.Join(dir, "api", gvk.Version, fmt.Sprintf("%s_types.go", strings.ToLower(gvk.Kind))),
		`//+kubebuilder:subresource:status`,
		`
		// +operator-sdk:csv:customresourcedefinitions:resources={{Deployment,v1,memcached-deployment}}
		`)

	if err != nil {
		return fmt.Errorf("encountered an error inserting Node Status: %w", err)
	}

	sampleFile := filepath.Join("config", "samples",
		fmt.Sprintf("%s_%s_%s.yaml", gvk.Group, gvk.Version, strings.ToLower(gvk.Kind)))

	err = kbutil.ReplaceInFile(filepath.Join(dir, sampleFile), "# TODO(user): Add fields here", "size: 1")
	if err != nil {
		return fmt.Errorf("encountered an error updating sample: %w", err)
	}

	return nil
}

func implementController(dir string, gvk schema.GroupVersionKind) error {
	controllerPath := filepath.Join(dir, "controllers", fmt.Sprintf("%s_controller.go",
		strings.ToLower(gvk.Kind)))

	// Add imports
	err := kbutil.InsertCode(controllerPath,
		"import (",
		importsFragment)
	if err != nil {
		return fmt.Errorf("encountered an error adding imports: %w", err)
	}
	// Add RBAC permissions on top of reconcile
	err = kbutil.InsertCode(controllerPath,
		"/finalizers,verbs=update",
		rbacFragment)
	if err != nil {
		return fmt.Errorf("encountered an error adding rbac: %w", err)
	}
	// Replace reconcile content
	err = kbutil.ReplaceInFile(controllerPath,
		`"sigs.k8s.io/controller-runtime/pkg/log"`,
		`ctrllog "sigs.k8s.io/controller-runtime/pkg/log"`,
	)
	if err != nil {
		return fmt.Errorf("encountered an error replacing controller log import: %w", err)
	}

	err = kbutil.ReplaceInFile(controllerPath,
		"_ = log.FromContext(ctx)",
		"log := ctrllog.FromContext(ctx)",
	)
	if err != nil {
		return fmt.Errorf("encountered an error replacing logger construction %w", err)
	}

	// Add reconcile implementation
	err = kbutil.ReplaceInFile(controllerPath,
		"// TODO(user): your logic here", reconcileFragment)
	if err != nil {
		return fmt.Errorf("encountered an error replacing reconcile content: %w", err)
	}

	// Add helpers funcs to the controller
	err = kbutil.InsertCode(controllerPath,
		"return ctrl.Result{}, nil\n}", controllerFuncsFragment)
	if err != nil {
		return fmt.Errorf("encountered an error adding helper methods in the controller: %w", err)
	}
	// Add watch for the Kind
	err = kbutil.ReplaceInFile(controllerPath,
		fmt.Sprintf(watchOriginalFragment, gvk.Group, gvk.Version, gvk.Kind),
		fmt.Sprintf(watchCustomizedFragment, gvk.Group, gvk.Version, gvk.Kind))
	if err != nil {
		return fmt.Errorf("encountered an error replacing add controller to manager: %w", err)
	}

	return nil
}

func implementWebhook(dir string, gvk schema.GroupVersionKind) error {
	webhookPath := filepath.Join(dir, "api", gvk.Version, fmt.Sprintf("%s_webhook.go",
		strings.ToLower(gvk.Kind)))
	// Add webhook methods
	err := kbutil.InsertCode(webhookPath,
		"// TODO(user): fill in your defaulting logic.\n}",
		webhooksFragment)
	if err != nil {
		return fmt.Errorf("encountered an error replacing webhook validate implementation: %w", err)
	}
	err = kbutil.ReplaceInFile(webhookPath,
		"// TODO(user): fill in your defaulting logic.", "if r.Spec.Size == 0 {\n\t\tr.Spec.Size = 3\n\t}")
	if err != nil {
		return fmt.Errorf("encountered an error replacing webhook default implementation: %w", err)
	} // Add imports
	err = kbutil.InsertCode(webhookPath,
		"import (",
		// TODO(estroz): remove runtime dep when --programmatic-validation is added to `ccreate webhook` above.
		"\"errors\"\n\n\"k8s.io/apimachinery/pkg/runtime\"")
	if err != nil {
		return fmt.Errorf("encountered an error adding imports: %w", err)
	}
	return nil
}

func uncommentDefaultKustomization(dir string) error {
	kustomization := filepath.Join(dir, "config", "default", "kustomization.yaml")

	err := kbutil.UncommentCode(kustomization, "#- ../webhook", "#")
	if err != nil {
		return fmt.Errorf("encountered an error uncomment webhook: %w", err)
	}

	err = kbutil.UncommentCode(kustomization, "#- ../certmanager", "#")
	if err != nil {
		return fmt.Errorf("encountered an error uncomment certmanager: %w", err)
	}

	err = kbutil.UncommentCode(kustomization, "#- ../prometheus", "#")
	if err != nil {
		return fmt.Errorf("encountered an error uncomment prometheus: %w", err)
	}

	err = kbutil.UncommentCode(kustomization, "#- manager_webhook_patch.yaml", "#")
	if err != nil {
		return fmt.Errorf("encountered an error uncomment manager_webhook_patch.yaml: %w", err)
	}

	err = kbutil.UncommentCode(kustomization, "#- webhookcainjection_patch.yaml", "#")
	if err != nil {
		return fmt.Errorf("encountered an error uncomment webhookcainjection_patch.yaml: %w", err)
	}

	err = kbutil.UncommentCode(kustomization,
		`#- name: CERTIFICATE_NAMESPACE # namespace of the certificate CR
#  objref:
#    kind: Certificate
#    group: cert-manager.io
#    version: v1
#    name: serving-cert # this name should match the one in certificate.yaml
#  fieldref:
#    fieldpath: metadata.namespace
#- name: CERTIFICATE_NAME
#  objref:
#    kind: Certificate
#    group: cert-manager.io
#    version: v1
#    name: serving-cert # this name should match the one in certificate.yaml
#- name: SERVICE_NAMESPACE # namespace of the service
#  objref:
#    kind: Service
#    version: v1
#    name: webhook-service
#  fieldref:
#    fieldpath: metadata.namespace
#- name: SERVICE_NAME
#  objref:
#    kind: Service
#    version: v1
#    name: webhook-service`, "#")
	if err != nil {
		return fmt.Errorf("encountered an error uncommented certificate CR: %w", err)
	}

	return nil
}

func uncommentManifestsKustomization(dir string) error {
	kustomization := filepath.Join(dir, "config", "manifests", "kustomization.yaml")

	err := kbutil.UncommentCode(kustomization,
		`#patchesJson6902:
#- target:
#    group: apps
#    version: v1
#    kind: Deployment
#    name: controller-manager
#    namespace: system
#  patch: |-
#    # Remove the manager container's "cert" volumeMount, since OLM will create and mount a set of certs.
#    # Update the indices in this path if adding or removing containers/volumeMounts in the manager's Deployment.
#    - op: remove
#      path: /spec/template/spec/containers/1/volumeMounts/0
#    # Remove the "cert" volume, since OLM will create and mount a set of certs.
#    # Update the indices in this path if adding or removing volumes in the manager's Deployment.
#    - op: remove
#      path: /spec/template/spec/volumes/0`, "#")
	if err != nil {
		return fmt.Errorf("encountered an error uncommented webhook volume removal patch: %w", err)
	}

	return nil
}

func generateBundle(sample samples.Sample, image string) error {
	dir := sample.CommandContext().Dir() + "/" + sample.Name()
	// Todo: check if we cannot improve it since the replace/content will exists in the
	// pkgmanifest target if it be scaffolded before this call
	content := "operator-sdk generate kustomize manifests"
	replace := content + " --interactive=false"
	err := ReplaceInFile(filepath.Join(dir, "Makefile"), content, replace)
	if err != nil {
		return err
	}

	cmd := exec.Command("make", "bundle", "IMG="+image)
	if _, err := sample.CommandContext().Run(cmd, sample.Name()); err != nil {
		return err
	}

	return nil
}

func ReplaceInFile(path, old, new string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if !strings.Contains(string(b), old) {
		return errors.New("unable to find the content to be replaced")
	}
	s := strings.Replace(string(b), old, new, -1)
	err = ioutil.WriteFile(path, []byte(s), info.Mode())
	if err != nil {
		return err
	}
	return nil
}

func stripBundleAnnotations(sample samples.Sample) error {
	dir := sample.CommandContext().Dir() + "/" + sample.Name()

	// Remove metadata labels.
	metadataAnnotations := makeBundleMetadataLabels("")
	metadataFiles := []string{
		filepath.Join(dir, "bundle", "metadata", "annotations.yaml"),
		filepath.Join(dir, "bundle.Dockerfile"),
	}
	if err := removeAllAnnotationLines(metadataAnnotations, metadataFiles); err != nil {
		return err
	}

	// Remove manifests annotations.
	manifestsAnnotations := makeBundleObjectAnnotations("")
	manifestsFiles := []string{
		filepath.Join(dir, "bundle", "manifests", sample.Name()+".clusterserviceversion.yaml"),
		filepath.Join(dir, "config", "manifests", "bases", sample.Name()+".clusterserviceversion.yaml"),
	}
	if err := removeAllAnnotationLines(manifestsAnnotations, manifestsFiles); err != nil {
		return err
	}

	return nil
}

func makeBundleMetadataLabels(layout string) map[string]string {
	mediaTypeBundleAnnotation := "operators.operatorframework.io.metrics.mediatype.v1"
	builderBundleAnnotation := "operators.operatorframework.io.metrics.builder"
	layoutBundleAnnotation := "operators.operatorframework.io.metrics.project_layout"

	return map[string]string{
		mediaTypeBundleAnnotation: "metrics+v1",
		builderBundleAnnotation:   "operator-sdk-1.20.0",
		layoutBundleAnnotation:    layout,
	}
}

func makeBundleObjectAnnotations(layout string) map[string]string {
	BuilderObjectAnnotation := "operators.operatorframework.io/builder"
	LayoutObjectAnnotation := "operators.operatorframework.io/project_layout"
	return map[string]string{
		BuilderObjectAnnotation: "operator-sdk-1.20.0",
		LayoutObjectAnnotation:  layout,
	}
}

func removeAllAnnotationLines(annotations map[string]string, filePaths []string) error {
	var annotationREs []*regexp.Regexp
	for annotation := range annotations {
		re, err := regexp.Compile(".+" + regexp.QuoteMeta(annotation) + ".+\n")
		if err != nil {
			return fmt.Errorf("compiling annotation regexp: %v", err)
		}
		annotationREs = append(annotationREs, re)
	}

	for _, file := range filePaths {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		for _, re := range annotationREs {
			b = re.ReplaceAll(b, []byte{})
		}
		err = ioutil.WriteFile(file, b, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

const rbacFragment = `
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch`

const reconcileFragment = `// Fetch the Memcached instance
	memcached := &cachev1alpha1.Memcached{}
	err := r.Get(ctx, req.NamespacedName, memcached)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Memcached resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Memcached")
		return ctrl.Result{}, err
	}
	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: memcached.Name, Namespace: memcached.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForMemcached(memcached)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}
	// Ensure the deployment size is the same as the spec
	size := memcached.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Ask to requeue after 1 minute in order to give enough time for the
		// pods be created on the cluster side and the operand be able
		// to do the next update step accurately.
		return ctrl.Result{RequeueAfter: time.Minute }, nil
	}
	// Update the Memcached status with the pod names
	// List the pods for this memcached's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(memcached.Namespace),
		client.MatchingLabels(labelsForMemcached(memcached.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "Memcached.Namespace", memcached.Namespace, "Memcached.Name", memcached.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)
	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, memcached.Status.Nodes) {
		memcached.Status.Nodes = podNames
		err := r.Status().Update(ctx, memcached)
		if err != nil {
			log.Error(err, "Failed to update Memcached status")
			return ctrl.Result{}, err
		}
	}
`

const controllerFuncsFragment = `
// deploymentForMemcached returns a memcached Deployment object
func (r *MemcachedReconciler) deploymentForMemcached(m *cachev1alpha1.Memcached) *appsv1.Deployment {
	ls := labelsForMemcached(m.Name)
	replicas := m.Spec.Size
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:   "memcached:1.4.36-alpine",
						Name:    "memcached",
						Command: []string{"memcached", "-m=64", "-o", "modern", "-v"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 11211,
							Name:          "memcached",
						}},
					}},
				},
			},
		},
	}
	// Set Memcached instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}
// labelsForMemcached returns the labels for selecting the resources
// belonging to the given memcached CR name.
func labelsForMemcached(name string) map[string]string {
	return map[string]string{"app": "memcached", "memcached_cr": name}
}
// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
`

const importsFragment = `
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"time"
`

const watchOriginalFragment = `return ctrl.NewControllerManagedBy(mgr).
		For(&%s%s.%s{}).
		Complete(r)
`

const watchCustomizedFragment = `return ctrl.NewControllerManagedBy(mgr).
		For(&%s%s.%s{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
`

const webhooksFragment = `
// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-cache-example-com-v1alpha1-memcached,mutating=false,failurePolicy=fail,sideEffects=None,groups=cache.example.com,resources=memcacheds,verbs=create;update,versions=v1alpha1,name=vmemcached.kb.io,admissionReviewVersions=v1
var _ webhook.Validator = &Memcached{}
// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Memcached) ValidateCreate() error {
	memcachedlog.Info("validate create", "name", r.Name)
	return validateOdd(r.Spec.Size)
}
// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Memcached) ValidateUpdate(old runtime.Object) error {
	memcachedlog.Info("validate update", "name", r.Name)
	return validateOdd(r.Spec.Size)
}
// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Memcached) ValidateDelete() error {
	memcachedlog.Info("validate delete", "name", r.Name)
	return nil
}
func validateOdd(n int32) error {
	if n%2 == 0 {
		return errors.New("Cluster size must be an odd number")
	}
	return nil
}
`
