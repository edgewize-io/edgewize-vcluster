package csistoragecapacities

import (
	"context"
	"fmt"

	"github.com/loft-sh/vcluster/pkg/constants"
	"github.com/loft-sh/vcluster/pkg/controllers/syncer"
	synccontext "github.com/loft-sh/vcluster/pkg/controllers/syncer/context"
	"github.com/loft-sh/vcluster/pkg/controllers/syncer/translator"
	"github.com/loft-sh/vcluster/pkg/util/clienthelper"
	"github.com/loft-sh/vcluster/pkg/util/translate"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ syncer.Syncer = &csistoragecapacitySyncer{}

func (s *csistoragecapacitySyncer) Name() string {
	return "csistoragecapacity"
}

func (s *csistoragecapacitySyncer) Resource() client.Object {
	return &storagev1.CSIStorageCapacity{}
}

func (s *csistoragecapacitySyncer) IsManaged(obj client.Object) (bool, error) {
	return true, nil
}

func (s *csistoragecapacitySyncer) RegisterIndices(ctx *synccontext.RegisterContext) error {
	return ctx.PhysicalManager.GetFieldIndexer().IndexField(ctx.Context, &storagev1.CSIStorageCapacity{}, constants.IndexByVirtualName, func(rawObj client.Object) []string {
		return []string{s.PhysicalToVirtual(rawObj).Name}
	})
}

// translate namespace
func (s *csistoragecapacitySyncer) PhysicalToVirtual(pObj client.Object) types.NamespacedName {
	return types.NamespacedName{Name: translate.SafeConcatName(pObj.GetName(), "x", pObj.GetNamespace()), Namespace: "kube-system"}
}
func (s *csistoragecapacitySyncer) VirtualToPhysical(req types.NamespacedName, vObj client.Object) types.NamespacedName {

	// if the virtual object is annotated with the physical name and namespace, return that
	if vObj != nil {
		vAnnotations := vObj.GetAnnotations()
		if vAnnotations != nil && vAnnotations[translator.NameAnnotation] != "" {
			return types.NamespacedName{
				Namespace: vAnnotations[translator.NamespaceAnnotation],
				Name:      vAnnotations[translator.NameAnnotation],
			}
		}
	}

	sc := &storagev1.CSIStorageCapacity{}
	pObj := sc.DeepCopyObject().(client.Object)
	err := clienthelper.GetByIndex(context.Background(), s.physicalClient, pObj, constants.IndexByVirtualName, req.Name)
	if err != nil {
		return types.NamespacedName{}
	}

	return types.NamespacedName{
		Namespace: pObj.GetNamespace(),
		Name:      pObj.GetName(),
	}
}

// TranslateMetadata translates the object's metadata
func (s *csistoragecapacitySyncer) TranslateMetadata(pObj client.Object) (client.Object, error) {
	name := s.PhysicalToVirtual(pObj)
	pObjCopy := pObj.DeepCopyObject()
	vObj, ok := pObjCopy.(client.Object)
	if !ok {
		return nil, fmt.Errorf("%q not a metadata object: %+v", pObj.GetName(), pObjCopy)
	}
	translator.ResetObjectMetadata(vObj)
	vObj.SetName(name.Name)
	vObj.SetNamespace(name.Namespace)
	vObj.SetLabels(translator.TranslateLabels(pObj, nil, []string{}))
	vObj.SetAnnotations(translator.TranslateAnnotations(pObj, nil, []string{}))
	return vObj, nil
}

// TranslateMetadataUpdate translates the object's metadata annotations and labels and determines
// if they have changed between the physical and virtual object
func (s *csistoragecapacitySyncer) TranslateMetadataUpdate(vObj client.Object, pObj client.Object) (changed bool, annotations map[string]string, labels map[string]string) {
	updatedAnnotations := translator.TranslateAnnotations(pObj, vObj, []string{})
	updatedLabels := translator.TranslateLabels(pObj, vObj, []string{})
	return !equality.Semantic.DeepEqual(updatedAnnotations, vObj.GetAnnotations()) || !equality.Semantic.DeepEqual(updatedLabels, vObj.GetLabels()), updatedAnnotations, updatedLabels
}
