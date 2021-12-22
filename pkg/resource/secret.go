package resource

import (
	"k8s-event-listener/pkg/eventlistener"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	v1 "k8s.io/api/core/v1"
)

func init() {
	Resources = append(Resources, getSecret())
}

func getSecret() resourceType {
	return resourceType{
		Name: []string{"secret", "secrets"},
		Fn: func(callback string) (r *eventlistener.Resource, e error) {
			r = &eventlistener.Resource{}
			r.ResourceName = "secrets"
			r.RestClient = func(clientset *kubernetes.Clientset) rest.Interface {
				return clientset.CoreV1().RESTClient()
			}
			r.ResourceType = &v1.Secret{}
			r.Callback = createCallbackFn(
				callback,
				r.ResourceName,
				func(obj interface{}, meta *callBackMeta) {
					if obj != nil {
						objType := obj.(*v1.Secret)
						meta.namespace = objType.GetNamespace()
						meta.name = objType.GetName()
					}
				},
			)

			return
		},
	}
}
