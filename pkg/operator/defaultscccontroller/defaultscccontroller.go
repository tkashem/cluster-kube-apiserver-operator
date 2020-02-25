package defaultscccontroller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	"github.com/openshift/library-go/pkg/operator/events"

	securityv1 "github.com/openshift/api/security/v1"
	securityv1client "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	securityv1listers "github.com/openshift/client-go/security/listers/security/v1"
)

const (
	ControllerName = "default-scc-upgradeable"
)

type Options struct {
	Config       *rest.Config
	Recorder     events.Recorder
	ResyncPeriod time.Duration
}

func NewDefaultSCCController(options *Options) (controller *DefaultSCCController, err error) {
	client, clientErr := securityv1client.NewForConfig(options.Config)
	if clientErr != nil {
		err = fmt.Errorf("[%s] failed to create client for security/v1 - %s", ControllerName, clientErr.Error())
		return
	}

	// Create a new SecurityContextConstraints watcher
	watcher := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return client.SecurityContextConstraints().List(options)
		},

		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return client.SecurityContextConstraints().Watch(options)
		},
	}

	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "DefaultSCCController")

	indexer, informer := cache.NewIndexerInformer(watcher, &securityv1.SecurityContextConstraints{}, options.ResyncPeriod, NewResourceHandler(queue), cache.Indexers{})
	lister := securityv1listers.NewSecurityContextConstraintsLister(indexer)
	recorder := options.Recorder.WithComponentSuffix(ControllerName)

	controller = &DefaultSCCController{
		queue:    queue,
		informer: informer,
		syncer: &Syncer{
			lister:   lister,
			recorder: recorder,
		},
		recorder: recorder,
	}

	return
}

type DefaultSCCController struct {
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
	recorder events.Recorder
	syncer   *Syncer
}

func (c *DefaultSCCController) Run(ctx context.Context) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Infof("[%s] Starting DefaultSCCController", ControllerName)
	defer klog.Infof("[%s] Shutting down DefaultSCCController", ControllerName)

	go c.informer.Run(ctx.Done())

	if !cache.WaitForCacheSync(ctx.Done(), c.informer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("[%s] cache for DefaultSCCController did not sync", ControllerName))
		return
	}

	go c.runWorker()
	<-ctx.Done()
}

func (c *DefaultSCCController) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *DefaultSCCController) processNextWorkItem() bool {
	key, shutdown := c.queue.Get()

	if shutdown {
		return false
	}

	defer c.queue.Done(key)

	request, ok := key.(types.NamespacedName)
	if !ok {
		// As the item in the workqueue is actually invalid, we call Forget here else
		// we'd go into a loop of attempting to process a work item that is invalid.
		c.queue.Forget(key)

		utilruntime.HandleError(fmt.Errorf("[%s] expected types.NamespacedName in workqueue but got %#v", ControllerName, key))
		return true
	}

	if err := c.syncer.Sync(request); err != nil {
		// Put the item back on the workqueue to handle any transient errors.
		c.queue.AddRateLimited(key)

		utilruntime.HandleError(fmt.Errorf("[%s] key=%s error syncing, requeuing - %s", ControllerName, request, err.Error()))
		return true
	}

	c.queue.Forget(key)
	return true
}
