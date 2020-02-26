package defaultscccontroller

import (
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

// NewResourceHandler returns a cache.ResourceEventHandler appropriate for
// reconciliation of SecurityContextConstraints object(s).
func NewResourceHandler(queue workqueue.RateLimitingInterface) ResourceHandler {
	return ResourceHandler{
		queue: queue,
	}
}

var _ cache.ResourceEventHandler = ResourceHandler{}

type ResourceHandler struct {
	// The underlying work queue where the keys are added for sync.
	queue workqueue.RateLimitingInterface
}

func (e ResourceHandler) OnAdd(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Errorf("[%s] OnAdd: could not extract key, type=%T- %s", ControllerName, obj, err.Error())
		return
	}

	e.add(key, e.queue)
}

// OnUpdate creates UpdateEvent and calls Update on EventHandler
func (e ResourceHandler) OnUpdate(oldObj, newObj interface{}) {
	// We don't distinguish between an add and update.
	e.OnAdd(newObj)
}

func (e ResourceHandler) OnDelete(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Errorf("[%s] OnDelete: could not extract key, type=%T - %s", ControllerName, obj, err.Error())
		return
	}

	e.add(key, e.queue)
}

func (e ResourceHandler) add(key string, queue workqueue.RateLimitingInterface) {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return
	}

	queue.Add(types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	})
}
