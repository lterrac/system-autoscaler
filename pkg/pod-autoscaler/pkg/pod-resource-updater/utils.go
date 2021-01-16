package resourceupdater

import (
	"fmt"

	"github.com/lterrac/system-autoscaler/pkg/apis/systemautoscaler/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const containerPatchTemplate = `[
	{
		"op":"replace",
		"path":"/spec/containers/0/resources/limits/cpu",
		"value":"%dm"
	},
	{
		"op":"replace",
		"path":"/spec/containers/0/resources/limits/memory",
		"value":"%dMi"
	},
	{
		"op":"replace",
		"path":"/spec/containers/0/resources/requests/cpu",
		"value":"%dm"
	},
	{
		"op":"replace",
		"path":"/spec/containers/0/resources/requests/memory",
		"value":"%dMi"
	}]`

const podscalePatchTemplate = `[
	{
		"op":"replace",
		"path":"/spec/desired/cpu",
		"value":"%dm"
	},
	{
		"op":"replace",
		"path":"/spec/desired/memory",
		"value":"%dMi"
	},
	{
		"op":"replace",
		"path":"/status/actual/cpu",
		"value":"%dm"
	},
	{
		"op":"replace",
		"path":"/status/actual/memory",
		"value":"%dMi"
	}]
`

func syncPod(pod v1.Pod, podScale v1beta1.PodScale) (*v1.Pod, error) {
	newPod := pod.DeepCopy()

	// TODO: we should be handle pod with multiple containers
	if len(newPod.Spec.Containers) != 1 {
		return nil, fmt.Errorf("the pod must have only 1 container. containers: %v", newPod.Spec.Containers)
	}

	if newPod.Status.QOSClass != v1.PodQOSGuaranteed {
		return nil, fmt.Errorf("the pod has %v but it should have 'guaranteed' QOS class", newPod.Status.QOSClass)
	}

	if podScale.Status.ActualResources.Cpu().MilliValue() <= 0 {
		return nil, fmt.Errorf("pod scale must have positive cpu resource value, actual value: %v", podScale.Status.ActualResources.Cpu().ScaledValue(resource.Milli))
	}

	if podScale.Status.ActualResources.Memory().MilliValue() <= 0 {
		return nil, fmt.Errorf("pod scale must have positive memory resource value, actual value: %v", podScale.Status.ActualResources.Memory().ScaledValue(resource.Mega))
	}

	newPod.Spec.Containers[0].Resources.Requests = podScale.Status.ActualResources
	newPod.Spec.Containers[0].Resources.Limits = podScale.Status.ActualResources

	// TODO: I should check that the QOS class is 'GUARANTEED'
	return newPod, nil

}

// NewContainerResourcePatch generates the payload required to patch container resources. Since the
// patch is performed on the Pod it is mandatory to specify the container name.
func NewContainerResourcePatch(pod *v1.Pod) []byte {
	// var containersOrder []string

	// for _, c := range pod.Spec.Containers {
	// 	containersOrder = append(containersOrder, fmt.Sprintf(`{"name":"%s"}`, c.Name))
	// }

	payload := fmt.Sprintf(containerPatchTemplate,
		pod.Spec.Containers[0].Resources.Limits.Cpu().ScaledValue(resource.Milli),
		pod.Spec.Containers[0].Resources.Limits.Memory().ScaledValue(resource.Mega),
		pod.Spec.Containers[0].Resources.Requests.Cpu().ScaledValue(resource.Milli),
		pod.Spec.Containers[0].Resources.Requests.Memory().ScaledValue(resource.Mega),
	)

	return []byte(payload)
}

// NewPodScaleResourcePatch generates the payload required to patch container resources. Since the
// patch is performed on the Pod it is mandatory to specify the container position in the array.
func NewPodScaleResourcePatch(podscale *v1beta1.PodScale) []byte {
	payload := fmt.Sprintf(podscalePatchTemplate,
		podscale.Spec.DesiredResources.Cpu().ScaledValue(resource.Milli),
		podscale.Spec.DesiredResources.Memory().ScaledValue(resource.Mega),
		podscale.Status.ActualResources.Cpu().ScaledValue(resource.Milli),
		podscale.Status.ActualResources.Memory().ScaledValue(resource.Mega),
	)

	return []byte(payload)
}
