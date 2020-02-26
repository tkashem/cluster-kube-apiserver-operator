package defaultscccontroller

import (
	"fmt"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
)

// OperatorStatusUpdater is responsible for updating the status block of the operator.
type OperatorStatusUpdater func(condition operatorv1.OperatorCondition) error

func (u OperatorStatusUpdater) UpdateStatus(condition operatorv1.OperatorCondition) error {
	return u(condition)
}

// NewOperatorStatusUpdater returns a OperatorStatusUpdater that can be used to update
// the status block.
//
// If the desired condition already exists in status and matches then update call
// is skipped.
// Two conditions are deemed equal if the corresponding Type, Status, Reason and
// Message fields match. Note that LastTransitionTime is ignored.
func NewOperatorStatusUpdater(client v1helpers.OperatorClient) OperatorStatusUpdater {
	return func(desired operatorv1.OperatorCondition) error {
		_, status, _, err := client.GetOperatorState()
		if err != nil {
			return err
		}

		if current := find(status, &desired); current != nil && isConditionEqual(&desired, current) {
			return nil
		}

		if _, _, updateError := v1helpers.UpdateStatus(client, v1helpers.UpdateConditionFn(desired)); updateError != nil {
			return updateError
		}

		return nil
	}
}

func NewCondition(mutated []string) operatorv1.OperatorCondition {
	status := operatorv1.ConditionTrue
	message := ""
	reason := "AsExpected"
	if len(mutated) > 0 {
		reason = "Mutated"
		status = operatorv1.ConditionFalse
		message = fmt.Sprintf("Default SecurityContextConstraints object(s) have mutated %s", mutated)
	}

	return operatorv1.OperatorCondition{
		Type:    "DefaultSecurityContextConstraintsUpgradeable",
		Reason:  reason,
		Status:  status,
		Message: message,
	}
}

func find(status *operatorv1.OperatorStatus, desired *operatorv1.OperatorCondition) *operatorv1.OperatorCondition {
	if status == nil {
		return nil
	}

	for i, _ := range status.Conditions {
		if desired.Type == status.Conditions[i].Type {
			return &status.Conditions[i]
		}
	}

	return nil
}

func isConditionEqual(desired, current *operatorv1.OperatorCondition) bool {
	if desired.Type == current.Type &&
		desired.Reason == current.Reason &&
		desired.Status == current.Status &&
		desired.Message == current.Message {
		return true
	}

	return false
}
