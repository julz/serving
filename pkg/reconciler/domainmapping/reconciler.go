package domainmapping

import (
	"context"

	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
	"knative.dev/serving/pkg/apis/serving/v1alpha1"
	domainmappingreconciler "knative.dev/serving/pkg/client/injection/reconciler/serving/v1alpha1/domainmapping"
)

type Reconciler struct{}

// Check that our Reconciler implements Interface
var _ domainmappingreconciler.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, dm *v1alpha1.DomainMapping) reconciler.Event {
	logger := logging.FromContext(ctx)
	logger.Debugf("Reconciling DomainMapping %s/%s", dm.Namespace, dm.Name)

	dm.Status.MarkIngressNotConfigured()

	url := &apis.URL{
		Scheme: "http",
		Host:   dm.Name,
	}

	url.Scheme = "https"
	dm.Status.URL = url
	dm.Status.Address = &duckv1.Addressable{URL: url}

	return nil
}
