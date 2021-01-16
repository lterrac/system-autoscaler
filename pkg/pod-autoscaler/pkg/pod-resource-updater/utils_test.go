package resourceupdater

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/lterrac/system-autoscaler/pkg/apis/systemautoscaler/v1beta1"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSyncPod(t *testing.T) {

	// TODO: The test case should be modified in future in order to handle more granularity.
	// Instead of pod resource values, we should insert cpu and mem values for each container.
	testcases := []struct {
		description            string
		podQOS                 corev1.PodQOSClass
		podNumberOfContainers  int64
		podCPUValue            int64
		podMemValue            int64
		podScaleCPUActualValue int64
		podScaleMemActualValue int64
		success                bool
	}{
		{
			description:            "successfully increased the resources of a pod",
			podQOS:                 corev1.PodQOSGuaranteed,
			podNumberOfContainers:  1,
			podCPUValue:            100,
			podMemValue:            100,
			podScaleCPUActualValue: 1000,
			podScaleMemActualValue: 1000,
			success:                true,
		},
		{
			description:            "successfully decreased the resources of a pod",
			podQOS:                 corev1.PodQOSGuaranteed,
			podNumberOfContainers:  1,
			podCPUValue:            100,
			podMemValue:            100,
			podScaleCPUActualValue: 1000,
			podScaleMemActualValue: 1000,
			success:                true,
		},
		{
			description:            "fail to update a pod with negative cpu resource value",
			podQOS:                 corev1.PodQOSGuaranteed,
			podNumberOfContainers:  1,
			podCPUValue:            100,
			podMemValue:            100,
			podScaleCPUActualValue: -1,
			podScaleMemActualValue: 1000,
			success:                false,
		},
		{
			description:            "fail to update a pod with negative memory resource value",
			podQOS:                 corev1.PodQOSGuaranteed,
			podNumberOfContainers:  1,
			podCPUValue:            100,
			podMemValue:            100,
			podScaleCPUActualValue: 1000,
			podScaleMemActualValue: -1,
			success:                false,
		},
		{
			description:            "fail to update a pod that has BE QOS",
			podQOS:                 corev1.PodQOSBestEffort,
			podNumberOfContainers:  1,
			podCPUValue:            100,
			podMemValue:            100,
			podScaleCPUActualValue: 1000,
			podScaleMemActualValue: 1000,
			success:                false,
		},
		{
			description:            "fail to update a pod that has BU QOS",
			podQOS:                 corev1.PodQOSBurstable,
			podNumberOfContainers:  1,
			podCPUValue:            100,
			podMemValue:            100,
			podScaleCPUActualValue: 1000,
			podScaleMemActualValue: 1000,
			success:                false,
		},
		{
			// TODO: this test should be changed once we are able to update multiple containers
			description:            "fail to update a pod that has multiple containers",
			podQOS:                 corev1.PodQOSGuaranteed,
			podNumberOfContainers:  2,
			podCPUValue:            100,
			podMemValue:            100,
			podScaleCPUActualValue: 1000,
			podScaleMemActualValue: 1000,
			success:                false,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.description, func(t *testing.T) {
			// Instantiate the containers
			containers := make([]corev1.Container, 0)
			for i := 0; i < int(tt.podNumberOfContainers); i++ {
				container := corev1.Container{
					Name:  fmt.Sprint("container-n-", i),
					Image: "gcr.io/distroless/static:nonroot",
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    *resource.NewScaledQuantity(tt.podCPUValue, resource.Milli),
							corev1.ResourceMemory: *resource.NewScaledQuantity(tt.podMemValue, resource.Mega),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    *resource.NewScaledQuantity(tt.podCPUValue, resource.Milli),
							corev1.ResourceMemory: *resource.NewScaledQuantity(tt.podMemValue, resource.Mega),
						},
					},
				}
				containers = append(containers, container)
			}
			// Instantiate the pod
			pod := corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: corev1.SchemeGroupVersion.String(),
					Kind:       "pods",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-name",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: containers,
				},
				Status: corev1.PodStatus{
					QOSClass: tt.podQOS,
				},
			}
			// Instantiate the pod scale
			podScale := v1beta1.PodScale{
				TypeMeta: metav1.TypeMeta{
					Kind:       "podscales",
					APIVersion: v1beta1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "podscale-name",
					Namespace: "default",
				},
				Spec: v1beta1.PodScaleSpec{
					PodRef: v1beta1.PodRef{
						Name:      "pod-name",
						Namespace: "default",
					},
					DesiredResources: corev1.ResourceList{
						corev1.ResourceCPU:    *resource.NewScaledQuantity(tt.podScaleCPUActualValue, resource.Milli),
						corev1.ResourceMemory: *resource.NewScaledQuantity(tt.podScaleMemActualValue, resource.Mega),
					},
				},
				Status: v1beta1.PodScaleStatus{
					ActualResources: corev1.ResourceList{
						corev1.ResourceCPU:    *resource.NewScaledQuantity(tt.podScaleCPUActualValue, resource.Milli),
						corev1.ResourceMemory: *resource.NewScaledQuantity(tt.podScaleMemActualValue, resource.Mega),
					},
				},
			}
			newPod, err := syncPod(pod, podScale)
			if tt.success {
				require.Nil(t, err, "Do not expect error")
				require.Equal(t, newPod.Spec.Containers[0].Resources.Limits.Cpu().ScaledValue(resource.Milli), tt.podScaleCPUActualValue)
				require.Equal(t, newPod.Spec.Containers[0].Resources.Requests.Cpu().ScaledValue(resource.Milli), tt.podScaleCPUActualValue)
				require.Equal(t, newPod.Spec.Containers[0].Resources.Limits.Memory().ScaledValue(resource.Mega), tt.podScaleMemActualValue)
				require.Equal(t, newPod.Spec.Containers[0].Resources.Requests.Memory().ScaledValue(resource.Mega), tt.podScaleMemActualValue)
				require.Equal(t, newPod.Status.QOSClass, corev1.PodQOSGuaranteed)
			} else {
				require.Error(t, err, "expected error")
			}
		})
	}
}
func TestNewContainerResourcePatch(t *testing.T) {
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "foo",
					Resources: corev1.ResourceRequirements{
						Limits: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    *resource.NewScaledQuantity(50, resource.Milli),
							corev1.ResourceMemory: *resource.NewScaledQuantity(50, resource.Mega),
						},
						Requests: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    *resource.NewScaledQuantity(50, resource.Milli),
							corev1.ResourceMemory: *resource.NewScaledQuantity(50, resource.Mega),
						},
					},
				},
				{
					Name: "bar",
					Resources: corev1.ResourceRequirements{
						Limits: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    *resource.NewScaledQuantity(50, resource.Milli),
							corev1.ResourceMemory: *resource.NewScaledQuantity(50, resource.Mega),
						},
						Requests: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    *resource.NewScaledQuantity(50, resource.Milli),
							corev1.ResourceMemory: *resource.NewScaledQuantity(50, resource.Mega),
						},
					},
				},
			},
		},
	}
	// "$setElementOrder/containers":[{"name":"foo"},{"name":"bar"}],

	payload := fmt.Sprint(`{"kind":"Pod","spec":{"containers":[{"name":"foo","resources":{"limits":{"cpu":"100m","memory":"100Mi"},"requests":{"cpu":"100m","memory":"100Mi"}}}]}}`)

	// resources := corev1.ResourceList{
	// 	corev1.ResourceCPU:    *resource.NewScaledQuantity(100, resource.Milli),
	// 	corev1.ResourceMemory: *resource.NewScaledQuantity(100, resource.Mega),
	// }

	actual := NewContainerResourcePatch(pod)

	actualString, err := strconv.Unquote(string(actual))
	require.Nil(t, err)

	require.Equal(t, payload, actualString)

	podunm := &corev1.Pod{}
	err = json.Unmarshal([]byte(`{"kind":"Pod","spec":{"containers":[{"name":"foo","resources":{"limits":{"cpu":"100m","memory":"100Mi"},"requests":{"cpu":"100m","memory":"100Mi"}}}]}}`), podunm)

	require.Nil(t, err)

}
