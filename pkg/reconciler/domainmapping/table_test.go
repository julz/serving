/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package domainmapping

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	network "knative.dev/networking/pkg"
	netv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
	networkingclient "knative.dev/networking/pkg/client/injection/client/fake"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	pkgnet "knative.dev/pkg/network"
	pkgreconciler "knative.dev/pkg/reconciler"
	"knative.dev/serving/pkg/apis/serving/v1alpha1"
	servingclient "knative.dev/serving/pkg/client/injection/client/fake"
	domainmappingreconciler "knative.dev/serving/pkg/client/injection/reconciler/serving/v1alpha1/domainmapping"
	"knative.dev/serving/pkg/reconciler/domainmapping/config"
	"knative.dev/serving/pkg/reconciler/domainmapping/resources"

	. "knative.dev/pkg/reconciler/testing"
	. "knative.dev/serving/pkg/reconciler/testing/v1"
)

const TestIngressClass = "ingress-class-foo"

func TestReconcile(t *testing.T) {
	table := TableTest{{
		Name: "bad workqueue key",
		// Make sure Reconcile handles bad keys.
		Key: "too/many/parts",
	}, {
		Name: "key not found",
		// Make sure Reconcile handles good keys that don't exist.
		Key: "foo/not-found",
	}, {
		Name: "first reconcile",
		Key:  "default/first-reconcile.com",
		Objects: []runtime.Object{
			DomainMapping("default", "first-reconcile.com", WithRef("default", "target")),
		},
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: DomainMapping("default", "first-reconcile.com",
				WithRef("default", "target"),
				WithURL("https", "first-reconcile.com"),
				WithAddress("https", "first-reconcile.com"),
				WithInitDomainMappingConditions,
				WithIngressConditionNotConfigured,
			),
		}},
		WantCreates: []runtime.Object{
			ingress(t, DomainMapping("default", "first-reconcile.com", WithRef("default", "target")), TestIngressClass),
		},
		WantEvents: []string{
			Eventf(corev1.EventTypeNormal, "Created", "Created Ingress %q", "first-reconcile.com"),
		},
	}, {
		Name: "ingress becomes ready",
		Key:  "default/becomes-ready.org",
		Objects: []runtime.Object{
			DomainMapping("default", "becomes-ready.org", WithRef("default", "target")),
			ingress(t, DomainMapping("default", "becomes-ready.org", WithRef("default", "target")),
				TestIngressClass,
				WithIngressStatus(readyIngressStatus()),
			),
		},
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: DomainMapping("default", "becomes-ready.org",
				WithRef("default", "target"),
				WithURL("https", "becomes-ready.org"),
				WithAddress("https", "becomes-ready.org"),
				WithInitDomainMappingConditions,
				WithIngressConditionReady,
			),
		}},
	}}

	table.Test(t, MakeFactory(func(ctx context.Context, listers *Listers, cmw configmap.Watcher) controller.Reconciler {
		r := &Reconciler{
			netclient:     networkingclient.Get(ctx),
			ingressLister: listers.GetIngressLister(),
		}

		return domainmappingreconciler.NewReconciler(ctx, logging.FromContext(ctx), servingclient.Get(ctx),
			listers.GetDomainMappingLister(), controller.GetEventRecorder(ctx), r,
			controller.Options{ConfigStore: &testConfigStore{config: ReconcilerTestConfig()}})
	}))
}

type testConfigStore struct {
	config *config.Config
}

func (t *testConfigStore) ToContext(ctx context.Context) context.Context {
	return config.ToContext(ctx, t.config)
}

var _ pkgreconciler.ConfigStore = (*testConfigStore)(nil)

func ReconcilerTestConfig() *config.Config {
	return &config.Config{
		Network: &network.Config{
			DefaultIngressClass:     TestIngressClass,
			DefaultCertificateClass: network.CertManagerCertificateClassName,
			DomainTemplate:          network.DefaultDomainTemplate,
			TagTemplate:             network.DefaultTagTemplate,
			HTTPProtocol:            network.HTTPEnabled,
		},
	}
}

type DomainMappingOption func(dm *v1alpha1.DomainMapping)

func DomainMapping(namespace, name string, opt ...DomainMappingOption) *v1alpha1.DomainMapping {
	dm := &v1alpha1.DomainMapping{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
	for _, o := range opt {
		o(dm)
	}
	return dm
}

func WithRef(namespace, name string) DomainMappingOption {
	return func(dm *v1alpha1.DomainMapping) {
		dm.Spec.Ref.Namespace = namespace
		dm.Spec.Ref.Name = name
	}
}

func WithURL(scheme, host string) DomainMappingOption {
	return func(dm *v1alpha1.DomainMapping) {
		dm.Status.URL = &apis.URL{Scheme: scheme, Host: host}
	}
}

func WithAddress(scheme, host string) DomainMappingOption {
	return func(dm *v1alpha1.DomainMapping) {
		dm.Status.Address = &duckv1.Addressable{URL: &apis.URL{
			Scheme: scheme,
			Host:   host,
		}}
	}
}

func WithInitDomainMappingConditions(dm *v1alpha1.DomainMapping) {
	dm.Status.InitializeConditions()
}

func WithIngressConditionNotConfigured(dm *v1alpha1.DomainMapping) {
	dm.Status.MarkIngressNotConfigured()
}

func WithIngressConditionReady(dm *v1alpha1.DomainMapping) {
	dm.Status.MarkIngressReady()
}

type IngressOption func(*netv1alpha1.Ingress)

func ingress(t *testing.T, dm *v1alpha1.DomainMapping, ingressClass string, io ...IngressOption) *netv1alpha1.Ingress {
	ing, err := resources.MakeIngress(dm, ingressClass)
	if err != nil {
		t.Fatalf("MakeIngress() = %v, expected no error", err)
	}

	for _, o := range io {
		o(ing)
	}

	return ing
}

func WithIngressStatus(status netv1alpha1.IngressStatus) IngressOption {
	return func(ing *netv1alpha1.Ingress) {
		ing.Status = status
	}
}

func readyIngressStatus() netv1alpha1.IngressStatus {
	status := netv1alpha1.IngressStatus{}
	status.InitializeConditions()
	status.MarkNetworkConfigured()
	status.MarkLoadBalancerReady(
		[]netv1alpha1.LoadBalancerIngressStatus{{
			DomainInternal: pkgnet.GetServiceHostname("istio-ingressgateway", "istio-system"),
		}},
		[]netv1alpha1.LoadBalancerIngressStatus{{
			DomainInternal: pkgnet.GetServiceHostname("private-istio-ingressgateway", "istio-system"),
		}},
	)

	return status
}
