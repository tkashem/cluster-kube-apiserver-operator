package defaultscccontroller

import (
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"

	securityv1listers "github.com/openshift/client-go/security/listers/security/v1"
	"github.com/openshift/library-go/pkg/operator/events"
)

type Syncer struct {
	lister        securityv1listers.SecurityContextConstraintsLister
	recorder      events.Recorder
	updater       OperatorStatusUpdater
	defaultSCCSet *DefaultSCC
}

func (s *Syncer) Sync(key types.NamespacedName) error {
	// If it's not to do with the default SCC, we don't care.
	if !s.defaultSCCSet.IsDefault(key.Name) {
		return nil
	}

	mutated := make([]string, 0)
	for _, original := range s.defaultSCCSet.List() {
		current, err := s.lister.Get(original.GetName())
		if err != nil {
			if k8serrors.IsNotFound(err) {
				klog.Infof("[%s] name=%s scc has been deleted - %s", ControllerName, original.GetName(), err.Error())
				continue
			}

			return err
		}

		if !IsSCCEqual(original, current) {
			mutated = append(mutated, current.GetName())
		}
	}

	if len(mutated) > 0 {
		klog.Infof("[%s] default scc has mutated %s", ControllerName, mutated)
	}

	condition := NewCondition(mutated)
	return s.updater.UpdateStatus(condition)
}
