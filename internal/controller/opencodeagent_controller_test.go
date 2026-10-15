package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	creuzzv1alpha1 "github.com/airklizz/creuzz/api/v1alpha1"
)

var _ = Describe("OpenCodeAgent Controller", func() {
	Context("When reconciling a resource", func() {
		const (
			resourceName      = "test-resource"
			resourceNamespace = "default"
		)

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: resourceNamespace,
		}
		opencodeagent := &creuzzv1alpha1.OpenCodeAgent{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind OpenCodeAgent")
			err := k8sClient.Get(ctx, typeNamespacedName, opencodeagent)
			if err != nil && errors.IsNotFound(err) {
				resource := &creuzzv1alpha1.OpenCodeAgent{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: resourceNamespace,
					},
					Spec: creuzzv1alpha1.OpenCodeAgentSpec{
						Project: creuzzv1alpha1.ProjectSpec{
							RepoURL:   "git@github.com:orga/test.git",
							Branch:    "main",
							SecretRef: "github-ssh-creds",
						},
						Runtime: creuzzv1alpha1.RuntimeSpec{
							Image:   "registry.interne/opencode-base:v2.0.0",
							CPU:     "1",
							Memory:  "2Gi",
							Storage: "5Gi",
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &creuzzv1alpha1.OpenCodeAgent{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				By("Cleanup the specific resource instance OpenCodeAgent")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &OpenCodeAgentReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
