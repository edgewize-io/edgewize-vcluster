package pods

import (
	"fmt"
	"testing"

	podtranslate "github.com/loft-sh/vcluster/pkg/controllers/resources/pods/translate"
	synccontext "github.com/loft-sh/vcluster/pkg/controllers/syncer/context"
	generictesting "github.com/loft-sh/vcluster/pkg/controllers/syncer/testing"
	"github.com/loft-sh/vcluster/pkg/controllers/syncer/translator"
	"github.com/loft-sh/vcluster/pkg/util/translate"
	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/pod-security-admission/api"
	"k8s.io/utils/pointer"
)

func TestSync(t *testing.T) {
	PodLogsVolumeName := "pod-logs"
	LogsVolumeName := "logs"
	KubeletPodVolumeName := "kubelet-pods"
	HostpathPodName := "test-hostpaths"

	pPodContainerEnv := []corev1.EnvVar{
		{
			Name:  "KUBERNETES_PORT",
			Value: "tcp://1.2.3.4:443",
		},
		{
			Name:  "KUBERNETES_PORT_443_TCP",
			Value: "tcp://1.2.3.4:443",
		},
		{
			Name:  "KUBERNETES_PORT_443_TCP_ADDR",
			Value: "1.2.3.4",
		},
		{
			Name:  "KUBERNETES_PORT_443_TCP_PORT",
			Value: "443",
		}, {
			Name:  "KUBERNETES_PORT_443_TCP_PROTO",
			Value: "tcp",
		}, {
			Name:  "KUBERNETES_SERVICE_HOST",
			Value: "1.2.3.4",
		}, {
			Name:  "KUBERNETES_SERVICE_PORT",
			Value: "443",
		}, {
			Name:  "KUBERNETES_SERVICE_PORT_HTTPS",
			Value: "443",
		},
	}

	pVclusterService := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      generictesting.DefaultTestVclusterServiceName,
			Namespace: generictesting.DefaultTestCurrentNamespace,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "1.2.3.4",
		},
	}
	translate.Suffix = generictesting.DefaultTestVclusterName
	pDNSService := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      translate.PhysicalName("kube-dns", "kube-system"),
			Namespace: generictesting.DefaultTestTargetNamespace,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "2.2.2.2",
		},
	}
	vNamespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testns",
		},
	}
	vObjectMeta := metav1.ObjectMeta{
		Name:      "testpod",
		Namespace: vNamespace.Name,
	}
	pObjectMeta := metav1.ObjectMeta{
		Name:      translate.PhysicalName("testpod", "testns"),
		Namespace: "test",
		Annotations: map[string]string{
			podtranslate.ClusterAutoScalerAnnotation:  "false",
			podtranslate.LabelsAnnotation:             "",
			podtranslate.NameAnnotation:               vObjectMeta.Name,
			podtranslate.NamespaceAnnotation:          vObjectMeta.Namespace,
			translator.NameAnnotation:                 vObjectMeta.Name,
			translator.NamespaceAnnotation:            vObjectMeta.Namespace,
			podtranslate.ServiceAccountNameAnnotation: "",
			podtranslate.UIDAnnotation:                string(vObjectMeta.UID),
		},
		Labels: map[string]string{
			translate.NamespaceLabel: vObjectMeta.Namespace,
			translate.MarkerLabel:    translate.Suffix,
		},
	}
	pPodBase := &corev1.Pod{
		ObjectMeta: pObjectMeta,
		Spec: corev1.PodSpec{
			AutomountServiceAccountToken: pointer.Bool(false),
			EnableServiceLinks:           pointer.Bool(false),
			HostAliases: []corev1.HostAlias{{
				IP:        pVclusterService.Spec.ClusterIP,
				Hostnames: []string{"kubernetes", "kubernetes.default", "kubernetes.default.svc"},
			}},
			Hostname: vObjectMeta.Name,
		},
	}
	vPodWithNodeName := &corev1.Pod{
		ObjectMeta: vObjectMeta,
		Spec: corev1.PodSpec{
			NodeName: "test123",
		},
	}
	pPodWithNodeName := pPodBase.DeepCopy()
	pPodWithNodeName.Spec.NodeName = "test456"

	vPodWithNodeSelector := &corev1.Pod{
		ObjectMeta: vObjectMeta,
		Spec: corev1.PodSpec{
			NodeSelector: map[string]string{
				"labelA": "valueA",
				"labelB": "valueB",
			},
		},
	}
	nodeSelectorOption := "labelB=enforcedB,otherLabel=abc"
	pPodWithNodeSelector := pPodBase.DeepCopy()
	pPodWithNodeSelector.Spec.NodeSelector = map[string]string{
		"labelA":     "valueA",
		"labelB":     "enforcedB",
		"otherLabel": "abc",
	}

	// pod security standards test objects
	vPodPSS := &corev1.Pod{
		ObjectMeta: vObjectMeta,
	}

	pPodPss := pPodBase.DeepCopy()

	vPodPSSR := &corev1.Pod{
		ObjectMeta: vObjectMeta,
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "test-container",
					Ports: []corev1.ContainerPort{
						{HostPort: 80},
					},
				},
			},
		},
	}

	vHostpathNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: generictesting.DefaultTestCurrentNamespace,
		},
	}

	vHostPathPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      HostpathPodName,
			Namespace: generictesting.DefaultTestCurrentNamespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nginx-placeholder",
					Image: "nginx",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      PodLogsVolumeName,
							MountPath: PodLoggingHostpathPath,
						},
						{
							Name:      LogsVolumeName,
							MountPath: LogHostpathPath,
						},
						{
							Name:      KubeletPodVolumeName,
							MountPath: KubeletPodPath,
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: PodLogsVolumeName,
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: PodLoggingHostpathPath,
						},
					},
				},
				{
					Name: LogsVolumeName,
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: LogHostpathPath,
						},
					},
				},
				{
					Name: KubeletPodVolumeName,
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: KubeletPodPath,
						},
					},
				},
			},
		},
	}

	vHostPath := fmt.Sprintf(VirtualPathTemplate, generictesting.DefaultTestTargetNamespace, generictesting.DefaultTestVclusterName)

	pHostPathPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      translate.PhysicalName(vHostPathPod.Name, generictesting.DefaultTestCurrentNamespace),
			Namespace: generictesting.DefaultTestTargetNamespace,

			Annotations: map[string]string{
				podtranslate.ClusterAutoScalerAnnotation:  "false",
				podtranslate.LabelsAnnotation:             "",
				podtranslate.NameAnnotation:               vHostPathPod.Name,
				podtranslate.NamespaceAnnotation:          vHostPathPod.Namespace,
				translator.NameAnnotation:                 vHostPathPod.Name,
				translator.NamespaceAnnotation:            vHostPathPod.Namespace,
				podtranslate.ServiceAccountNameAnnotation: "",
				podtranslate.UIDAnnotation:                string(vHostPathPod.UID),
			},
			Labels: map[string]string{
				translate.NamespaceLabel: vHostPathPod.Namespace,
				translate.MarkerLabel:    translate.Suffix,
			},
			// CreationTimestamp: metav1.Time{},
			// ResourceVersion:   "999",
		},
		Spec: corev1.PodSpec{
			AutomountServiceAccountToken: pointer.Bool(false),
			EnableServiceLinks:           pointer.Bool(false),
			HostAliases: []corev1.HostAlias{{
				IP:        pVclusterService.Spec.ClusterIP,
				Hostnames: []string{"kubernetes", "kubernetes.default", "kubernetes.default.svc"},
			}},
			Hostname: vHostPathPod.Name,
			Containers: []corev1.Container{
				{
					Name:  "nginx-placeholder",
					Image: "nginx",
					Env:   pPodContainerEnv,
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      PodLogsVolumeName,
							MountPath: PodLoggingHostpathPath,
						},
						{
							Name:      LogsVolumeName,
							MountPath: LogHostpathPath,
						},
						{
							Name:      KubeletPodVolumeName,
							MountPath: KubeletPodPath,
						},
						{
							Name:      fmt.Sprintf("%s-%s", PodLogsVolumeName, PhysicalVolumeNameSuffix),
							MountPath: PhysicalPodLogVolumeMountPath,
						},
						{
							Name:      fmt.Sprintf("%s-%s", LogsVolumeName, PhysicalVolumeNameSuffix),
							MountPath: PhysicalLogVolumeMountPath,
						},
						{
							Name:      fmt.Sprintf("%s-%s", KubeletPodVolumeName, PhysicalVolumeNameSuffix),
							MountPath: PhysicalKubeletVolumeMountPath,
						},
					},
				},
			},

			Volumes: []corev1.Volume{
				{
					Name: PodLogsVolumeName,
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: vHostPath + "/log/pods",
						},
					},
				},
				{
					Name: LogsVolumeName,
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: vHostPath + "/log",
						},
					},
				},
				{
					Name: KubeletPodVolumeName,
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: vHostPath + "/kubelet/pods",
						},
					},
				},
				{
					Name: fmt.Sprintf("%s-%s", PodLogsVolumeName, PhysicalVolumeNameSuffix),
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: PodLoggingHostpathPath,
						},
					},
				},
				{
					Name: fmt.Sprintf("%s-%s", LogsVolumeName, PhysicalVolumeNameSuffix),
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: LogHostpathPath,
						},
					},
				},
				{
					Name: fmt.Sprintf("%s-%s", KubeletPodVolumeName, PhysicalVolumeNameSuffix),
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: KubeletPodPath,
						},
					},
				},
			},
		},
	}

	generictesting.RunTests(t, []*generictesting.SyncTest{
		{
			Name:                 "Delete virtual pod",
			InitialVirtualState:  []runtime.Object{vPodWithNodeName.DeepCopy()},
			InitialPhysicalState: []runtime.Object{pPodWithNodeName},
			ExpectedVirtualState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("Pod"): {},
			},
			ExpectedPhysicalState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("Pod"): {
					pPodWithNodeName,
				},
			},
			Sync: func(ctx *synccontext.RegisterContext) {
				syncCtx, syncer := generictesting.FakeStartSyncer(t, ctx, New)
				_, err := syncer.(*podSyncer).Sync(syncCtx, pPodWithNodeName.DeepCopy(), vPodWithNodeName)
				assert.NilError(t, err)
			},
		},
		{
			Name:                 "Sync and enforce NodeSelector",
			InitialVirtualState:  []runtime.Object{vPodWithNodeSelector.DeepCopy(), vNamespace.DeepCopy()},
			InitialPhysicalState: []runtime.Object{pVclusterService.DeepCopy(), pDNSService.DeepCopy()},
			ExpectedVirtualState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("Pod"): {vPodWithNodeSelector.DeepCopy()},
			},
			ExpectedPhysicalState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("Pod"): {
					pPodWithNodeSelector,
				},
			},
			Sync: func(ctx *synccontext.RegisterContext) {
				ctx.Options.EnforceNodeSelector = true
				ctx.Options.NodeSelector = nodeSelectorOption
				syncCtx, syncer := generictesting.FakeStartSyncer(t, ctx, New)
				_, err := syncer.(*podSyncer).SyncDown(syncCtx, vPodWithNodeSelector.DeepCopy())
				assert.NilError(t, err)
			},
		},
		{
			Name:                 "SyncDown pods without any pod security standards",
			InitialVirtualState:  []runtime.Object{vPodPSS.DeepCopy(), vNamespace.DeepCopy()},
			InitialPhysicalState: []runtime.Object{pVclusterService.DeepCopy(), pDNSService.DeepCopy()},
			ExpectedVirtualState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("Pod"): {vPodPSS.DeepCopy()},
			},
			ExpectedPhysicalState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("Pod"): {pPodPss.DeepCopy()},
			},
			Sync: func(ctx *synccontext.RegisterContext) {
				syncCtx, syncer := generictesting.FakeStartSyncer(t, ctx, New)
				_, err := syncer.(*podSyncer).SyncDown(syncCtx, vPodPSS.DeepCopy())
				assert.NilError(t, err)
			},
		},
		{
			Name:                 "Enforce privileged pod security standard",
			InitialVirtualState:  []runtime.Object{vPodPSS.DeepCopy(), vNamespace.DeepCopy()},
			InitialPhysicalState: []runtime.Object{pVclusterService.DeepCopy(), pDNSService.DeepCopy()},
			ExpectedVirtualState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("Pod"): {vPodPSS.DeepCopy()},
			},
			ExpectedPhysicalState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("Pod"): {pPodPss.DeepCopy()},
			},
			Sync: func(ctx *synccontext.RegisterContext) {
				ctx.Options.EnforcePodSecurityStandard = string(api.LevelPrivileged)
				syncCtx, syncer := generictesting.FakeStartSyncer(t, ctx, New)
				_, err := syncer.(*podSyncer).SyncDown(syncCtx, vPodPSS.DeepCopy())
				assert.NilError(t, err)
			},
		},
		{
			Name:                 "Enforce restricted pod security standard",
			InitialVirtualState:  []runtime.Object{vPodPSSR.DeepCopy(), vNamespace.DeepCopy()},
			InitialPhysicalState: []runtime.Object{pVclusterService.DeepCopy(), pDNSService.DeepCopy()},
			ExpectedVirtualState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("Pod"): {vPodPSSR.DeepCopy()},
			},
			ExpectedPhysicalState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("Pod"): {},
			},
			Sync: func(ctx *synccontext.RegisterContext) {
				ctx.Options.EnforcePodSecurityStandard = string(api.LevelRestricted)
				syncCtx, syncer := generictesting.FakeStartSyncer(t, ctx, New)
				_, err := syncer.(*podSyncer).SyncDown(syncCtx, vPodPSSR.DeepCopy())
				assert.NilError(t, err)
			},
		},
		{
			Name:                 "Map hostpaths",
			InitialVirtualState:  []runtime.Object{vHostPathPod, vHostpathNamespace},
			InitialPhysicalState: []runtime.Object{pVclusterService.DeepCopy(), pDNSService.DeepCopy()},
			ExpectedVirtualState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("Pod"): {vHostPathPod.DeepCopy()},
			},
			ExpectedPhysicalState: map[schema.GroupVersionKind][]runtime.Object{
				corev1.SchemeGroupVersion.WithKind("Pod"): {pHostPathPod.DeepCopy()},
			},
			Sync: func(ctx *synccontext.RegisterContext) {
				ctx.TargetNamespace = generictesting.DefaultTestTargetNamespace
				synccontext, syncer := generictesting.FakeStartSyncer(t, ctx, New)
				_, err := syncer.(*podSyncer).SyncDown(synccontext, vHostPathPod.DeepCopy())
				assert.NilError(t, err)
			},
		},
	})
}

func TestRewritePVCToNFS(t *testing.T) {
	s := &podSyncer{
		pvcReplaceType: "nfs",
		nfsServer:      "10.10.10.10",
		nfsPath:        "/nfsdata",
	}
	vPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mypod",
			Namespace: "testns",
		},
	}
	pPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mypod",
			Namespace: "vcluster1",
		},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{
				{
					Name: "data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "myclaim",
						},
					},
				},
			},
		},
	}

	// 正常情况
	pPod1 := pPod.DeepCopy()
	pPod1 = s.rewritePVC(vPod, pPod1)
	assert.Equal(t, len(pPod1.Spec.Volumes), 1)
	vol := pPod1.Spec.Volumes[0]
	assert.Equal(t, vol.Name, "data")
	assert.Assert(t, vol.VolumeSource.NFS != nil)
	assert.Equal(t, vol.VolumeSource.NFS.Server, "10.10.10.10")
	assert.Equal(t, vol.VolumeSource.NFS.Path, "/nfsdata/vcluster1/testns/myclaim")

	// pvcReplaceType 为空时，不应替换
	s2 := &podSyncer{
		pvcReplaceType: "",
		nfsServer:      "10.10.10.10",
		nfsPath:        "/nfsdata",
	}
	pPod2 := pPod.DeepCopy()
	pPod2 = s2.rewritePVC(vPod, pPod2)
	vol2 := pPod2.Spec.Volumes[0]
	assert.Assert(t, vol2.VolumeSource.NFS == nil)
	assert.Assert(t, vol2.VolumeSource.PersistentVolumeClaim != nil)

	// pvcReplaceType 为 none 时，不应替换
	s3 := &podSyncer{
		pvcReplaceType: "none",
		nfsServer:      "10.10.10.10",
		nfsPath:        "/nfsdata",
	}
	pPod3 := pPod.DeepCopy()
	pPod3 = s3.rewritePVC(vPod, pPod3)
	vol3 := pPod3.Spec.Volumes[0]
	assert.Assert(t, vol3.VolumeSource.NFS == nil)
	assert.Assert(t, vol3.VolumeSource.PersistentVolumeClaim != nil)

	// nfsPath 前缀不带斜杠（已由初始化逻辑和 TestPodSyncerNfsPathNormalization 覆盖，无需在此重复测试）
	// Pod 没有 PVC
	pPod5 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mypod",
			Namespace: "vcluster1",
		},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{
				{
					Name: "config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: "cm"}},
					},
				},
			},
		},
	}
	pPod5 = s.rewritePVC(vPod, pPod5)
	assert.Assert(t, pPod5.Spec.Volumes[0].VolumeSource.NFS == nil)
	assert.Assert(t, pPod5.Spec.Volumes[0].VolumeSource.ConfigMap != nil)

	// Pod 有多个 PVC 和其他类型 Volume
	pPod6 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mypod",
			Namespace: "vcluster1",
		},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{
				{
					Name: "data1",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "claim1"},
					},
				},
				{
					Name: "data2",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "claim2"},
					},
				},
				{
					Name: "cm",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: "cm"}},
					},
				},
			},
		},
	}
	pPod6 = s.rewritePVC(vPod, pPod6)
	assert.Assert(t, pPod6.Spec.Volumes[0].VolumeSource.NFS != nil)
	assert.Equal(t, pPod6.Spec.Volumes[0].VolumeSource.NFS.Path, "/nfsdata/vcluster1/testns/claim1")
	assert.Assert(t, pPod6.Spec.Volumes[1].VolumeSource.NFS != nil)
	assert.Equal(t, pPod6.Spec.Volumes[1].VolumeSource.NFS.Path, "/nfsdata/vcluster1/testns/claim2")
	assert.Assert(t, pPod6.Spec.Volumes[2].VolumeSource.NFS == nil)
	assert.Assert(t, pPod6.Spec.Volumes[2].VolumeSource.ConfigMap != nil)

	// PVC 名称、namespace、cluster 含特殊字符（改为合法 DNS-1123 label）
	vPod7 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mypod",
			Namespace: "ns-1-test",
		},
	}
	pPod7 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mypod",
			Namespace: "clu-2-test",
		},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{
				{
					Name: "data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "claim-special"},
					},
				},
			},
		},
	}
	pPod7 = s.rewritePVC(vPod7, pPod7)
	assert.Equal(t, pPod7.Spec.Volumes[0].VolumeSource.NFS.Path, "/nfsdata/clu-2-test/ns-1-test/claim-special")
}
func TestNormalizeNfsPath(t *testing.T) {
	assert.Equal(t, normalizeNfsPath("nfsdata/"), "/nfsdata")
	assert.Equal(t, normalizeNfsPath("/nfsdata/"), "/nfsdata")
	assert.Equal(t, normalizeNfsPath("/nfsdata"), "/nfsdata")
	assert.Equal(t, normalizeNfsPath("nfsdata"), "/nfsdata")
	assert.Equal(t, normalizeNfsPath("/"), "/")
	assert.Equal(t, normalizeNfsPath(""), "")
}
