package defaultscccontroller

import (
	"testing"

	"github.com/stretchr/testify/require"

	securityv1 "github.com/openshift/api/security/v1"
)

func TestRender(t *testing.T) {
	defaultSCCGot, errGot := Render()

	require.NoError(t, errGot)
	require.NotNil(t, defaultSCCGot)

	defaultSCCWant := []string{
		"anyuid",
		"hostaccess",
		"hostmount-anyuid",
		"hostnetwork",
		"nonroot",
		"privileged",
		"restricted",
	}

	for _, nameWant := range defaultSCCWant {
		require.True(t, defaultSCCGot.IsDefault(nameWant))

		sccGot := helperFind(nameWant, defaultSCCGot.List())
		require.NotNil(t, sccGot)
		require.Equal(t, nameWant, sccGot.GetName())
	}
}

func helperFind(name string, list []*securityv1.SecurityContextConstraints) *securityv1.SecurityContextConstraints {
	for i, _ := range list {
		if name == list[i].GetName() {
			return list[i]
		}
	}

	return nil
}
