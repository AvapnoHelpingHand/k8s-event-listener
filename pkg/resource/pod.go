package resource

import (
	"k8s-event-listener/pkg/eventlistener"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	v1 "k8s.io/api/core/v1"
)

func init() {
	Resources = append(Resources, getPod())
}

func getPod() resourceType {
	return resourceType{
		Name: []string{"p", "pod", "pods"},
		Fn: func(callback string) (r *eventlistener.Resource, e error) {
			r = &eventlistener.Resource{}
			r.ResourceName = "pods"
			r.RestClient = func(clientset *kubernetes.Clientset) rest.Interface {
				return clientset.CoreV1().RESTClient()
			}
			r.ResourceType = &v1.Pod{}
			r.Callback = createCallbackFn(
				callback,
				r.ResourceName,
				func(obj interface{}, meta *callBackMeta) {
					if obj != nil {
						objType := obj.(*v1.Pod)
						meta.namespace = objType.GetNamespace()
						meta.name = objType.GetName()
					}
				},
			)

			return
		},
	}
}
