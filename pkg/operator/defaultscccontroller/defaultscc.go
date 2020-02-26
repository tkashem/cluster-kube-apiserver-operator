package defaultscccontroller

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"github.com/openshift/api"
	securityv1 "github.com/openshift/api/security/v1"

	assets "github.com/openshift/cluster-kube-apiserver-operator/pkg/operator/defaultscc_assets"
)

type DefaultSCC struct {
	list []*securityv1.SecurityContextConstraints
}

func (d *DefaultSCC) IsDefault(name string) bool {
	for _, scc := range d.list {
		if scc.GetName() == name {
			return true
		}
	}

	return false
}

func (d *DefaultSCC) List() []*securityv1.SecurityContextConstraints {
	return d.list
}

// Render renders the assets associated with the default set of SCC into a list
// of SecurityContextConstraints object(s).
func Render() (defaultSCC *DefaultSCC, err error) {
	decoder, decoderErr := decoder()
	if decoderErr != nil {
		err = fmt.Errorf("failed to create decoder - %s", decoderErr.Error())
		return
	}

	list := make([]*securityv1.SecurityContextConstraints, 0)
	set := make(map[string]*securityv1.SecurityContextConstraints)
	for _, name := range assets.AssetNames() {
		bytes, assetErr := assets.Asset(name)
		if assetErr != nil {
			return
		}

		object, _, decodeErr := decoder.Decode(bytes, nil, nil)
		if decodeErr != nil {
			err = fmt.Errorf("failed to decode SecurityContextConstraints from asset name=%s - %s", name, decodeErr.Error())
			return
		}

		scc, ok := object.(*securityv1.SecurityContextConstraints)
		if !ok {
			err = fmt.Errorf("obj is not SecurityContextConstraint type name=%s", name)
			return
		}

		_, exists := set[scc.GetName()]
		if exists {
			err = fmt.Errorf("SecurityContextConstraint already exists in set name=%s", scc.GetName())
			return
		}

		set[scc.GetName()] = scc
		list = append(list, scc)
	}

	defaultSCC = &DefaultSCC{
		list: list,
	}
	return
}

func IsSCCEqual(original, current *securityv1.SecurityContextConstraints) bool {
	copy := current.DeepCopy()

	copy.UID = ""
	copy.Generation = 0
	copy.ResourceVersion = ""
	copy.SelfLink = ""
	copy.CreationTimestamp = metav1.Time{}

	return equality.Semantic.DeepEqual(original, copy)
}

func decoder() (decoder runtime.Decoder, err error) {
	scheme := runtime.NewScheme()
	if err = api.InstallKube(scheme); err != nil {
		return
	}
	if err = api.Install(scheme); err != nil {
		return
	}

	factory := serializer.NewCodecFactory(scheme)

	decoder = factory.UniversalDeserializer()
	return
}
