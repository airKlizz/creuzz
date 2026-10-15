package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	creuzzv1alpha1 "github.com/airklizz/creuzz/api/v1alpha1"
)

const (
	opencodePort       = 4096
	workspaceMountPath = "/workspace"
	configMountPath    = "/home/opencode/.config/opencode"
	gitSecretVolume    = "git-ssh"
	workspaceVolume    = "workspace"
	configVolume       = "opencode-config"

	labelAppName      = "app.kubernetes.io/name"
	labelAppInstance  = "app.kubernetes.io/instance"
	labelAppComponent = "app.kubernetes.io/component"
	appNameValue      = "opencode-agent"
)

// OpenCodeAgentReconciler reconciles a OpenCodeAgent object
type OpenCodeAgentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=creuzz.io,resources=opencodeagents,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=creuzz.io,resources=opencodeagents/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=creuzz.io,resources=opencodeagents/finalizers,verbs=update
// +kubebuilder:rbac:groups=creuzz.io,resources=clusteragenttools,verbs=get;list;watch
// +kubebuilder:rbac:groups=creuzz.io,resources=clusteragentskills,verbs=get;list;watch
// +kubebuilder:rbac:groups=creuzz.io,resources=clusteragentmcpservers,verbs=get;list;watch
// +kubebuilder:rbac:groups=creuzz.io,resources=clusteragentenvs,verbs=get;list;watch
// +kubebuilder:rbac:groups=creuzz.io,resources=clusteragentegressrules,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *OpenCodeAgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var agent creuzzv1alpha1.OpenCodeAgent
	if err := r.Get(ctx, req.NamespacedName, &agent); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get OpenCodeAgent")
		return ctrl.Result{}, err
	}

	if agent.Status.Phase == "" {
		if err := r.updateStatus(ctx, &agent, creuzzv1alpha1.AgentPhasePending, "Reconciliation started"); err != nil {
			return ctrl.Result{}, err
		}
	}

	resources, err := r.resolveClusterResources(ctx, &agent)
	if err != nil {
		logger.Error(err, "Failed to resolve cluster resources")
		if updateErr := r.updateStatus(ctx, &agent, creuzzv1alpha1.AgentPhaseFailed, err.Error()); updateErr != nil {
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{}, err
	}

	configMap := r.buildConfigMap(&agent, resources)
	if err := r.createOrUpdate(ctx, &agent, configMap); err != nil {
		logger.Error(err, "Failed to create/update ConfigMap")
		return ctrl.Result{}, err
	}

	netPol := r.buildNetworkPolicy(&agent, resources)
	if err := r.createOrUpdate(ctx, &agent, netPol); err != nil {
		logger.Error(err, "Failed to create/update NetworkPolicy")
		return ctrl.Result{}, err
	}

	pvc := r.buildPVC(&agent)
	if err := r.createOrUpdate(ctx, &agent, pvc); err != nil {
		logger.Error(err, "Failed to create/update PVC")
		return ctrl.Result{}, err
	}

	pod := r.buildPod(&agent, resources, configMap.Name, pvc.Name)
	if err := r.createOrUpdate(ctx, &agent, pod); err != nil {
		logger.Error(err, "Failed to create/update Pod")
		return ctrl.Result{}, err
	}

	svc := r.buildService(&agent)
	if err := r.createOrUpdate(ctx, &agent, svc); err != nil {
		logger.Error(err, "Failed to create/update Service")
		return ctrl.Result{}, err
	}

	agentURL := fmt.Sprintf("https://agent-%s-%s.internal.local", agent.Namespace, agent.Name)
	if err := r.updateStatusWithDetails(ctx, &agent, creuzzv1alpha1.AgentPhaseReady, "Agent is ready", agentURL, pod.Name, svc.Name); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Successfully reconciled OpenCodeAgent", "name", agent.Name, "namespace", agent.Namespace)
	return ctrl.Result{}, nil
}

type resolvedResources struct {
	tools       []creuzzv1alpha1.ClusterAgentTool
	skills      []creuzzv1alpha1.ClusterAgentSkill
	mcpServers  []creuzzv1alpha1.ClusterAgentMCPServer
	envs        []creuzzv1alpha1.ClusterAgentEnv
	egressRules []creuzzv1alpha1.ClusterAgentEgressRule
}

func (r *OpenCodeAgentReconciler) resolveClusterResources(ctx context.Context, agent *creuzzv1alpha1.OpenCodeAgent) (*resolvedResources, error) {
	res := &resolvedResources{}

	for _, ref := range agent.Spec.ClusterToolRefs {
		var tool creuzzv1alpha1.ClusterAgentTool
		if err := r.Get(ctx, types.NamespacedName{Name: ref.Name}, &tool); err != nil {
			return nil, fmt.Errorf("ClusterAgentTool %q not found: %w", ref.Name, err)
		}
		res.tools = append(res.tools, tool)
	}

	for _, ref := range agent.Spec.ClusterSkillRefs {
		var skill creuzzv1alpha1.ClusterAgentSkill
		if err := r.Get(ctx, types.NamespacedName{Name: ref.Name}, &skill); err != nil {
			return nil, fmt.Errorf("ClusterAgentSkill %q not found: %w", ref.Name, err)
		}
		res.skills = append(res.skills, skill)
	}

	for _, ref := range agent.Spec.ClusterMcpRefs {
		var mcp creuzzv1alpha1.ClusterAgentMCPServer
		if err := r.Get(ctx, types.NamespacedName{Name: ref.Name}, &mcp); err != nil {
			return nil, fmt.Errorf("ClusterAgentMCPServer %q not found: %w", ref.Name, err)
		}
		res.mcpServers = append(res.mcpServers, mcp)
	}

	for _, ref := range agent.Spec.ClusterEnvRefs {
		var env creuzzv1alpha1.ClusterAgentEnv
		if err := r.Get(ctx, types.NamespacedName{Name: ref.Name}, &env); err != nil {
			return nil, fmt.Errorf("ClusterAgentEnv %q not found: %w", ref.Name, err)
		}
		res.envs = append(res.envs, env)
	}

	for _, ref := range agent.Spec.ClusterEgressRefs {
		var rule creuzzv1alpha1.ClusterAgentEgressRule
		if err := r.Get(ctx, types.NamespacedName{Name: ref.Name}, &rule); err != nil {
			return nil, fmt.Errorf("ClusterAgentEgressRule %q not found: %w", ref.Name, err)
		}
		res.egressRules = append(res.egressRules, rule)
	}

	return res, nil
}

func (r *OpenCodeAgentReconciler) buildConfigMap(agent *creuzzv1alpha1.OpenCodeAgent, res *resolvedResources) *corev1.ConfigMap {
	data := make(map[string]string)

	for _, skill := range res.skills {
		data["skill-"+skill.Spec.FileName] = skill.Spec.Content
	}

	mcpConfig := map[string]any{
		"mcp": map[string]any{
			"servers": map[string]any{},
		},
	}
	servers := mcpConfig["mcp"].(map[string]any)["servers"].(map[string]any)
	for _, mcp := range res.mcpServers {
		servers[mcp.Name] = map[string]any{
			"type": "sse",
			"url":  fmt.Sprintf("http://localhost:%d", mcp.Spec.Port),
		}
	}
	opencodeJSON, _ := json.MarshalIndent(mcpConfig, "", "  ")
	data["opencode.json"] = string(opencodeJSON)

	return &corev1.ConfigMap{
		Name:      agent.Name + "-config",
		Namespace: agent.Namespace,
		Data:      data,
	}
}

func buildConfigMapItems(res *resolvedResources) []corev1.KeyToPath {
	items := make([]corev1.KeyToPath, 1, 1+len(res.skills))
	items[0] = corev1.KeyToPath{Key: "opencode.json", Path: "opencode.json"}
	for _, skill := range res.skills {
		items = append(items, corev1.KeyToPath{
			Key:  "skill-" + skill.Spec.FileName,
			Path: "skills/" + skill.Spec.FileName,
		})
	}
	return items
}

func (r *OpenCodeAgentReconciler) buildNetworkPolicy(agent *creuzzv1alpha1.OpenCodeAgent, res *resolvedResources) *networkingv1.NetworkPolicy {
	egressRules := make([]networkingv1.NetworkPolicyEgressRule, 1, 1+len(res.egressRules))
	egressRules[0] = networkingv1.NetworkPolicyEgressRule{
		Ports: []networkingv1.NetworkPolicyPort{
			{Protocol: protocolPtr(corev1.ProtocolUDP), Port: intstrPtr(53)},
			{Protocol: protocolPtr(corev1.ProtocolTCP), Port: intstrPtr(53)},
		},
		To: []networkingv1.NetworkPolicyPeer{
			{
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"kubernetes.io/metadata.name": "kube-system"},
				},
				PodSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"k8s-app": "kube-dns"},
				},
			},
		},
	}

	for _, rule := range res.egressRules {
		peers := []networkingv1.NetworkPolicyPeer{}
		if len(rule.Spec.Domains) > 0 {
			dnsName := rule.Spec.Domains[0]
			for _, domain := range rule.Spec.Domains[1:] {
				_ = domain
			}
			peers = append(peers, networkingv1.NetworkPolicyPeer{
				PodSelector:       &metav1.LabelSelector{},
				NamespaceSelector: &metav1.LabelSelector{},
			})
			_ = dnsName
		}

		ports := []networkingv1.NetworkPolicyPort{}
		for _, p := range rule.Spec.Ports {
			protocol := corev1.Protocol(p.Protocol)
			ports = append(ports, networkingv1.NetworkPolicyPort{
				Protocol: &protocol,
				Port:     intstrPtr(int(p.Port)),
			})
		}
		if len(ports) == 0 {
			ports = append(ports, networkingv1.NetworkPolicyPort{
				Protocol: protocolPtr(corev1.ProtocolTCP),
				Port:     intstrPtr(443),
			})
		}

		egressRules = append(egressRules, networkingv1.NetworkPolicyEgressRule{
			Ports: ports,
			To:    peers,
		})
	}

	return &networkingv1.NetworkPolicy{
		Name:      agent.Name + "-netpol",
		Namespace: agent.Namespace,
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{labelAppName: appNameValue, labelAppInstance: agent.Name},
			},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					Ports: []networkingv1.NetworkPolicyPort{
						{Protocol: protocolPtr(corev1.ProtocolTCP), Port: intstrPtr(opencodePort)},
					},
				},
			},
			Egress: egressRules,
		},
	}
}

func (r *OpenCodeAgentReconciler) buildPVC(agent *creuzzv1alpha1.OpenCodeAgent) *corev1.PersistentVolumeClaim {
	storageQty := resource.MustParse(agent.Spec.Runtime.Storage)
	return &corev1.PersistentVolumeClaim{
		Name:      agent.Name + "-workspace",
		Namespace: agent.Namespace,
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: storageQty,
				},
			},
		},
	}
}

func (r *OpenCodeAgentReconciler) buildPod(agent *creuzzv1alpha1.OpenCodeAgent, res *resolvedResources, configMapName, pvcName string) *corev1.Pod {
	labels := map[string]string{
		labelAppName:      appNameValue,
		labelAppInstance:  agent.Name,
		labelAppComponent: "agent",
	}

	volumes := make([]corev1.Volume, 0, 3+len(res.tools))
	volumes = append(volumes,
		corev1.Volume{
			Name: workspaceVolume,
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvcName,
			},
		},
		corev1.Volume{
			Name: configVolume,
			ConfigMap: &corev1.ConfigMapVolumeSource{
				Name:  configMapName,
				Items: buildConfigMapItems(res),
			},
		},
		corev1.Volume{
			Name: gitSecretVolume,
			Secret: &corev1.SecretVolumeSource{
				SecretName: agent.Spec.Project.SecretRef,
			},
		},
	)

	volumeMounts := make([]corev1.VolumeMount, 0, 2+len(res.tools))
	volumeMounts = append(volumeMounts,
		corev1.VolumeMount{Name: workspaceVolume, MountPath: workspaceMountPath},
		corev1.VolumeMount{Name: configVolume, MountPath: configMountPath},
	)

	for _, tool := range res.tools {
		volName := "tool-" + tool.Name
		volumes = append(volumes, corev1.Volume{
			Name: volName,
			Image: &corev1.ImageVolumeSource{
				Reference:  tool.Spec.Image,
				PullPolicy: corev1.PullIfNotPresent,
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      volName,
			MountPath: tool.Spec.MountPath,
			ReadOnly:  true,
		})
	}

	pathParts := []string{}
	for _, tool := range res.tools {
		if tool.Spec.BinPath != "" {
			pathParts = append(pathParts, tool.Spec.BinPath)
		}
	}
	pathEnv := strings.Join(pathParts, ":") + ":/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

	envVars := []corev1.EnvVar{
		{Name: "PATH", Value: pathEnv},
		{Name: "OPENCODE_WORKSPACE", Value: workspaceMountPath},
	}

	for _, env := range res.envs {
		for _, v := range env.Spec.Variables {
			envVars = append(envVars, corev1.EnvVar{Name: v.Name, Value: v.Value})
		}
		for _, s := range env.Spec.Secrets {
			envVars = append(envVars, corev1.EnvVar{
				Name: s.Name,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						Name: s.SecretRef,
						Key:  s.Key,
					},
				},
			})
		}
	}

	containers := make([]corev1.Container, 0, len(res.mcpServers)+1)

	for _, mcp := range res.mcpServers {
		mcpContainer := corev1.Container{
			Name:  "mcp-" + mcp.Name,
			Image: mcp.Spec.Image,
			Ports: []corev1.ContainerPort{
				{ContainerPort: mcp.Spec.Port, Protocol: corev1.ProtocolTCP},
			},
		}
		if len(mcp.Spec.Args) > 0 {
			mcpContainer.Args = mcp.Spec.Args
		}
		if mcp.Spec.EnvFromSecret != "" {
			mcpContainer.EnvFrom = []corev1.EnvFromSource{
				{
					SecretRef: &corev1.SecretEnvSource{
						Name: mcp.Spec.EnvFromSecret,
					},
				},
			}
		}
		containers = append(containers, mcpContainer)
	}

	mainContainer := corev1.Container{
		Name:       "opencode",
		Image:      agent.Spec.Runtime.Image,
		WorkingDir: workspaceMountPath,
		Command:    []string{"opencode", "serve", "--port", fmt.Sprintf("%d", opencodePort), "--host", "0.0.0.0"},
		Ports: []corev1.ContainerPort{
			{ContainerPort: opencodePort, Protocol: corev1.ProtocolTCP},
		},
		Env:          envVars,
		VolumeMounts: volumeMounts,
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(agent.Spec.Runtime.CPU),
				corev1.ResourceMemory: resource.MustParse(agent.Spec.Runtime.Memory),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(agent.Spec.Runtime.CPU),
				corev1.ResourceMemory: resource.MustParse(agent.Spec.Runtime.Memory),
			},
		},
	}
	containers = append(containers, mainContainer)

	initContainers := []corev1.Container{
		{
			Name:  "git-clone",
			Image: "alpine/git:latest",
			Command: []string{"sh", "-c", fmt.Sprintf(`
				set -e
				mkdir -p /root/.ssh
				cp /git-ssh/id_ed25519 /root/.ssh/id_ed25519 2>/dev/null || true
				cp /git-ssh/known_hosts /root/.ssh/known_hosts 2>/dev/null || true
				chmod 600 /root/.ssh/id_ed25519 2>/dev/null || true
				git config --global user.email "agent@opencode.local"
				git config --global user.name "OpenCode Agent"
				git clone --branch %s --single-branch %s %s
			`, agent.Spec.Project.Branch, agent.Spec.Project.RepoURL, workspaceMountPath)},
			VolumeMounts: []corev1.VolumeMount{
				{Name: workspaceVolume, MountPath: workspaceMountPath},
				{Name: gitSecretVolume, MountPath: "/git-ssh", ReadOnly: true},
			},
			Env: []corev1.EnvVar{
				{Name: "GIT_SSH_COMMAND", Value: "ssh -o StrictHostKeyChecking=accept-new -i /root/.ssh/id_ed25519"},
			},
		},
	}

	return &corev1.Pod{
		Name:      agent.Name + "-agent",
		Namespace: agent.Namespace,
		Labels:    labels,
		Spec: corev1.PodSpec{
			InitContainers: initContainers,
			Containers:     containers,
			Volumes:        volumes,
			RestartPolicy:  corev1.RestartPolicyAlways,
		},
	}
}

func (r *OpenCodeAgentReconciler) buildService(agent *creuzzv1alpha1.OpenCodeAgent) *corev1.Service {
	return &corev1.Service{
		Name:      agent.Name + "-svc",
		Namespace: agent.Namespace,
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				labelAppName:     appNameValue,
				labelAppInstance: agent.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       opencodePort,
					TargetPort: intstr.FromInt(opencodePort),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

func (r *OpenCodeAgentReconciler) createOrUpdate(ctx context.Context, owner *creuzzv1alpha1.OpenCodeAgent, obj client.Object) error {
	if err := ctrl.SetControllerReference(owner, obj, r.Scheme); err != nil {
		return err
	}

	existing := obj.DeepCopyObject().(client.Object)
	key := types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}

	err := r.Get(ctx, key, existing)
	if errors.IsNotFound(err) {
		return r.Create(ctx, obj)
	}
	if err != nil {
		return err
	}

	switch o := obj.(type) {
	case *corev1.ConfigMap:
		ex := existing.(*corev1.ConfigMap)
		ex.Data = o.Data
		ex.BinaryData = o.BinaryData
		return r.Update(ctx, ex)
	case *corev1.Pod:
		ex := existing.(*corev1.Pod)
		if err := r.Delete(ctx, ex); err != nil {
			return fmt.Errorf("failed to delete existing pod for recreation: %w", err)
		}
		return fmt.Errorf("pod deleted for recreation, will be recreated on next reconciliation")
	case *corev1.Service:
		ex := existing.(*corev1.Service)
		ex.Spec.Ports = o.Spec.Ports
		ex.Spec.Selector = o.Spec.Selector
		return r.Update(ctx, ex)
	case *corev1.PersistentVolumeClaim:
		return nil
	case *networkingv1.NetworkPolicy:
		ex := existing.(*networkingv1.NetworkPolicy)
		ex.Spec = o.Spec
		return r.Update(ctx, ex)
	}
	return nil
}

func (r *OpenCodeAgentReconciler) updateStatus(ctx context.Context, agent *creuzzv1alpha1.OpenCodeAgent, phase creuzzv1alpha1.AgentPhase, message string) error {
	return r.updateStatusWithDetails(ctx, agent, phase, message, "", "", "")
}

func (r *OpenCodeAgentReconciler) updateStatusWithDetails(ctx context.Context, agent *creuzzv1alpha1.OpenCodeAgent, phase creuzzv1alpha1.AgentPhase, message, url, podName, svcName string) error {
	agent.Status.Phase = phase
	if url != "" {
		agent.Status.URL = url
	}
	if podName != "" {
		agent.Status.PodName = podName
	}
	if svcName != "" {
		agent.Status.ServiceName = svcName
	}

	condition := metav1.Condition{
		Type:               string(phase),
		Status:             metav1.ConditionTrue,
		Reason:             "Reconciliation",
		Message:            message,
		LastTransitionTime: metav1.Now(),
	}
	if phase == creuzzv1alpha1.AgentPhaseFailed {
		condition.Status = metav1.ConditionFalse
	}

	agent.Status.Conditions = []metav1.Condition{condition}
	return r.Status().Update(ctx, agent)
}

func (r *OpenCodeAgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&creuzzv1alpha1.OpenCodeAgent{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Watches(&creuzzv1alpha1.ClusterAgentTool{}, handler.EnqueueRequestsFromMapFunc(r.makeFindAgentsFunc(func(a *creuzzv1alpha1.OpenCodeAgent) []creuzzv1alpha1.ClusterResourceRef {
			return a.Spec.ClusterToolRefs
		}))).
		Watches(&creuzzv1alpha1.ClusterAgentSkill{}, handler.EnqueueRequestsFromMapFunc(r.makeFindAgentsFunc(func(a *creuzzv1alpha1.OpenCodeAgent) []creuzzv1alpha1.ClusterResourceRef {
			return a.Spec.ClusterSkillRefs
		}))).
		Watches(&creuzzv1alpha1.ClusterAgentMCPServer{}, handler.EnqueueRequestsFromMapFunc(r.makeFindAgentsFunc(func(a *creuzzv1alpha1.OpenCodeAgent) []creuzzv1alpha1.ClusterResourceRef {
			return a.Spec.ClusterMcpRefs
		}))).
		Watches(&creuzzv1alpha1.ClusterAgentEnv{}, handler.EnqueueRequestsFromMapFunc(r.makeFindAgentsFunc(func(a *creuzzv1alpha1.OpenCodeAgent) []creuzzv1alpha1.ClusterResourceRef {
			return a.Spec.ClusterEnvRefs
		}))).
		Watches(&creuzzv1alpha1.ClusterAgentEgressRule{}, handler.EnqueueRequestsFromMapFunc(r.makeFindAgentsFunc(func(a *creuzzv1alpha1.OpenCodeAgent) []creuzzv1alpha1.ClusterResourceRef {
			return a.Spec.ClusterEgressRefs
		}))).
		Named("opencodeagent").
		Complete(r)
}

func (r *OpenCodeAgentReconciler) makeFindAgentsFunc(getRefs func(*creuzzv1alpha1.OpenCodeAgent) []creuzzv1alpha1.ClusterResourceRef) handler.MapFunc {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		logger := log.FromContext(ctx)
		var agentList creuzzv1alpha1.OpenCodeAgentList
		if err := r.List(ctx, &agentList); err != nil {
			logger.Error(err, "Failed to list OpenCodeAgents")
			return nil
		}

		objName := obj.GetName()
		var requests []reconcile.Request

		for i := range agentList.Items {
			agent := &agentList.Items[i]
			refs := getRefs(agent)
			for _, ref := range refs {
				if ref.Name == objName {
					requests = append(requests, reconcile.Request{
						Name:      agent.Name,
						Namespace: agent.Namespace,
					})
					break
				}
			}
		}

		return requests
	}
}

func protocolPtr(p corev1.Protocol) *corev1.Protocol {
	q := new(corev1.Protocol)
	*q = p
	return q
}
func intstrPtr(i int) *intstr.IntOrString {
	v := intstr.FromInt(i)
	return &v
}
