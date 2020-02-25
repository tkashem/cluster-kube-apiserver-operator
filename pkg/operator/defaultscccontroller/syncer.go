package defaultscccontroller

import (
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"

	securityv1listers "github.com/openshift/client-go/security/listers/security/v1"
	"github.com/openshift/library-go/pkg/operator/events"
)

type Syncer struct {
	lister   securityv1listers.SecurityContextConstraintsLister
	recorder events.Recorder
}

func (s *Syncer) Sync(key types.NamespacedName) error {
	current, err := s.lister.Get(key.Name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.Infof("[%s] key=%s scc has been deleted - %s", ControllerName, key.Name, err.Error())
			return nil
		}

		return err
	}

	klog.Infof("[%s] key=%s scc has been updated", ControllerName, current.Name)
	return nil
}
