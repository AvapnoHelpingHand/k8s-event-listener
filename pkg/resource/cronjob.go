package resource

import (
	"k8s-event-listener/pkg/eventlistener"

	"k8s.io/api/batch/v1beta1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func init() {
	Resources = append(Resources, getCronjob())
}

func getCronjob() resourceType {
	return resourceType{
		Name: []string{"cj", "cronjob", "cronjobs"},
		Fn: func(callback string) (r *eventlistener.Resource, e error) {
			r = &eventlistener.Resource{}
			r.ResourceName = "cronjobs"
			r.RestClient = func(clientset *kubernetes.Clientset) rest.Interface {
				return clientset.BatchV1beta1().RESTClient()
			}
			r.ResourceType = &v1beta1.CronJob{}
			r.Callback = createCallbackFn(
				callback,
				r.ResourceName,
				func(obj interface{}, meta *callBackMeta) {
					if obj != nil {
						objType := obj.(*v1beta1.CronJob)
						meta.namespace = objType.GetNamespace()
						meta.name = objType.GetName()
					}
				},
			)

			return
		},
	}
}
