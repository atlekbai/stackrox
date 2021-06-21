package reconciler

import (
	pkgReconciler "github.com/joelanford/helm-operator/pkg/reconciler"
	"github.com/stackrox/rox/image"
	centralV1Alpha1 "github.com/stackrox/rox/operator/api/central/v1alpha1"
	"github.com/stackrox/rox/operator/pkg/central/extensions"
	"github.com/stackrox/rox/operator/pkg/central/values/translation"
	"github.com/stackrox/rox/operator/pkg/reconciler"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

// RegisterNewReconciler registers a new helm reconciler in the given k8s controller manager
func RegisterNewReconciler(mgr ctrl.Manager) error {
	// TODO(ROX-7415): Use a single client.
	return reconciler.SetupReconcilerWithManager(
		mgr, centralV1Alpha1.CentralGVK, image.CentralServicesChartPrefix, translation.Translator{Config: mgr.GetConfig()},
		pkgReconciler.WithPreExtension(extensions.ReconcileCentralTLSExtensions(kubernetes.NewForConfigOrDie(mgr.GetConfig()))))
}