package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	creuzzv1alpha1 "github.com/airklizz/creuzz/api/v1alpha1"
)

var _ = Describe("ClusterAgentTool Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name: resourceName,
		}
		clusteragenttool := &creuzzv1alpha1.ClusterAgentTool{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind ClusterAgentTool")
			err := k8sClient.Get(ctx, typeNamespacedName, clusteragenttool)
			if err != nil && errors.IsNotFound(err) {
				resource := &creuzzv1alpha1.ClusterAgentTool{
					Name: resourceName,
					Spec: creuzzv1alpha1.ClusterAgentToolSpec{
						Image:     "registry.interne/tools/test:1.0.0",
						MountPath: "/opt/tools/test",
						BinPath:   "/opt/tools/test/bin",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &creuzzv1alpha1.ClusterAgentTool{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance ClusterAgentTool")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &ClusterAgentToolReconciler{
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
