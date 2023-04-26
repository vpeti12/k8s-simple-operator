/*
Copyright 2023 vpeti12.

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

package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	myappv1alpha1 "github.com/vpeti12/k8s-simple-operator/api/v1alpha1"
)

// SimpleAppReconciler reconciles a SimpleApp object
type SimpleAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=myapp.peter.gyuk,resources=simpleapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=myapp.peter.gyuk,resources=simpleapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=myapp.peter.gyuk,resources=simpleapps/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups="cert-manager.io",resources=issuers,verbs=get;list;watch;create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// Business logic: The reconciliation includes creating a new Deployment,
// Service, Issuer and Ingress based on the SimpleApp API instance if it
// doesn't already exist or update the specific mentioned resource if it
// already exists.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *SimpleAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var instance myappv1alpha1.SimpleApp
	r.Client.Get(ctx, req.NamespacedName, &instance)

	newDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-deployment",
			Namespace: instance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "SimpleApp",
				},
			},
			Replicas: instance.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "SimpleApp",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  instance.Name,
							Image: instance.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	newService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-service",
			Namespace: instance.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "SimpleApp",
			},
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
			},
		},
	}

	newIssuer := &certmanagerv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "selfsigned-issuer",
			Namespace: instance.Namespace,
		},
		Spec: certmanagerv1.IssuerSpec{
			IssuerConfig: certmanagerv1.IssuerConfig{
				SelfSigned: &certmanagerv1.SelfSignedIssuer{},
			},
		},
	}

	newIngress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-ingress",
			Namespace: instance.Namespace,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class": "nginx",
				"cert-manager.io/issuer":      "selfsigned-issuer",
			},
		},
		Spec: networkingv1.IngressSpec{
			TLS: []networkingv1.IngressTLS{
				{
					Hosts: []string{
						instance.Spec.Host,
					},
					SecretName: "tls-secret",
				},
			},
			Rules: []networkingv1.IngressRule{
				{
					Host: instance.Spec.Host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									PathType: func(pt networkingv1.PathType) *networkingv1.PathType { return &pt }(networkingv1.PathTypePrefix),
									Path:     "/",
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: instance.Name + "-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	oldDeployment := appsv1.Deployment{}
	oldService := corev1.Service{}
	oldIngress := networkingv1.Ingress{}
	oldIssuer := certmanagerv1.Issuer{}

	err := r.Client.Get(ctx, types.NamespacedName{Name: newDeployment.Name, Namespace: newDeployment.Namespace}, &oldDeployment)
	if err != nil && errors.IsNotFound(err) {
		r.Client.Create(ctx, newDeployment)

	} else {
		oldDeployment.Spec = newDeployment.Spec
		r.Client.Update(ctx, &oldDeployment)
	}

	err = r.Client.Get(ctx, types.NamespacedName{Name: newService.Name, Namespace: newService.Namespace}, &oldService)
	if err != nil && errors.IsNotFound(err) {
		r.Client.Create(ctx, newService)
	} else {
		oldService.Spec = newService.Spec
		r.Client.Update(ctx, &oldService)
	}

	err = r.Client.Get(ctx, types.NamespacedName{Name: newIssuer.Name, Namespace: newIssuer.Namespace}, &oldIssuer)
	if err != nil && errors.IsNotFound(err) {
		r.Client.Create(ctx, newIssuer)
	} else {
		oldIssuer.Spec = newIssuer.Spec
		r.Client.Update(ctx, &oldIssuer)
	}

	err = r.Client.Get(ctx, types.NamespacedName{Name: newIngress.Name, Namespace: newIngress.Namespace}, &oldIngress)
	if err != nil && errors.IsNotFound(err) {
		r.Client.Create(ctx, newIngress)
	} else {
		oldIngress.Spec = newIngress.Spec
		r.Client.Update(ctx, &oldIngress)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SimpleAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&myappv1alpha1.SimpleApp{}).
		Complete(r)
}
