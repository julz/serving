package resources

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"knative.dev/networking/pkg/apis/networking"
	netv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
	"knative.dev/pkg/kmeta"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
)

func MakeIngress(
	dm *servingv1alpha1.DomainMapping,
	ingressClass string,
) (*netv1alpha1.Ingress, error) {
	targetServiceName := dm.Spec.Ref.Name
	targetHostName := targetServiceName + "." + dm.Spec.Ref.Namespace + ".svc.cluster.local"

	return &netv1alpha1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kmeta.ChildName(dm.GetName(), ""),
			Namespace: dm.Namespace,
			Annotations: kmeta.FilterMap(kmeta.UnionMaps(map[string]string{
				networking.IngressClassAnnotationKey: ingressClass,
			}, dm.GetAnnotations()), func(key string) bool {
				return key == corev1.LastAppliedConfigAnnotation
			}),
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(dm)},
		},
		Spec: netv1alpha1.IngressSpec{
			Rules: []netv1alpha1.IngressRule{{
				Hosts:      []string{dm.Name},
				Visibility: netv1alpha1.IngressVisibilityExternalIP,
				HTTP: &netv1alpha1.HTTPIngressRuleValue{
					Paths: []netv1alpha1.HTTPIngressPath{{
						RewriteHost: targetHostName,
						Splits: []netv1alpha1.IngressBackendSplit{{
							IngressBackend: netv1alpha1.IngressBackend{
								ServiceName:      targetServiceName,
								ServiceNamespace: dm.Namespace,
								ServicePort:      intstr.FromInt(80),
							},
						}},
					}},
				},
			}},
		},
	}, nil
}
