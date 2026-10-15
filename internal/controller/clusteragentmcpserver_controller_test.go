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

var _ = Describe("ClusterAgentMCPServer Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name: resourceName,
		}
		clusteragentmcpserver := &creuzzv1alpha1.ClusterAgentMCPServer{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind ClusterAgentMCPServer")
			err := k8sClient.Get(ctx, typeNamespacedName, clusteragentmcpserver)
			if err != nil && errors.IsNotFound(err) {
				resource := &creuzzv1alpha1.ClusterAgentMCPServer{
					Name: resourceName,
					Spec: creuzzv1alpha1.ClusterAgentMCPServerSpec{
						Image: "registry.interne/mcp/test:1.0.0",
						Port:  8001,
						Args:  []string{"--port", "8001"},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &creuzzv1alpha1.ClusterAgentMCPServer{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance ClusterAgentMCPServer")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &ClusterAgentMCPServerReconciler{
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
