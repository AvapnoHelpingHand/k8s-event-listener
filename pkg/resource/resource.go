package resource

import (
	"fmt"
	"k8s-event-listener/pkg/eventlistener"
	"k8s-event-listener/pkg/resource/internal"
	"log"
	"strings"

	"github.com/go-cmd/cmd"
)

type callBackMeta struct {
	callback, resourceName, action, namespace, name string
}

type resourceType struct {
	Name []string
	Fn   func(string) (*eventlistener.Resource, error)
}

var Resources []resourceType

// NewResource returns pointer to Resource and/or error
func NewResource(resourceName, callback string) (*eventlistener.Resource, error) {
	resourceName = strings.ToLower(resourceName)
	for _, resource := range Resources {
		if internal.Contains(resource.Name, resourceName) {
			return resource.Fn(callback)
		}
	}

	return nil, fmt.Errorf("unknown resource %s", resourceName)
}

func createCallbackFn(callback, resourceName string, closure func(interface{}, *callBackMeta)) eventlistener.CallbackFn {
	return func(event eventlistener.Event, i interface{}) error {
		cm := callBackMeta{
			callback:     callback,
			resourceName: resourceName,
			name:         event.Key,
			action:       event.Action,
		}

		closure(i, &cm)

		return createCallback(cm)
	}
}

func createCallback(meta callBackMeta) error {
	findCmd := cmd.NewCmd(meta.callback, meta.resourceName, meta.action, meta.namespace, meta.name)
	statusChan := findCmd.Start()
	status := <-statusChan

	log.Printf("stdout: %s\n", status.Stdout)
	log.Printf("stderr: %s\n", status.Stderr)

	return status.Error
}
