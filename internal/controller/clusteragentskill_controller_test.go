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

var _ = Describe("ClusterAgentSkill Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name: resourceName,
		}
		clusteragentskill := &creuzzv1alpha1.ClusterAgentSkill{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind ClusterAgentSkill")
			err := k8sClient.Get(ctx, typeNamespacedName, clusteragentskill)
			if err != nil && errors.IsNotFound(err) {
				resource := &creuzzv1alpha1.ClusterAgentSkill{
					ObjectMeta: metav1.ObjectMeta{
						Name: resourceName,
					},
					Spec: creuzzv1alpha1.ClusterAgentSkillSpec{
						FileName: "test-skill.md",
						Content:  "# Test Skill\nThis is a test skill.",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &creuzzv1alpha1.ClusterAgentSkill{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance ClusterAgentSkill")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &ClusterAgentSkillReconciler{
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
