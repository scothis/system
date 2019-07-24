/*
Copyright 2018 The Knative Authors

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

package testing

import (
	knbuildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	fakeknbuildclientset "github.com/knative/build/pkg/client/clientset/versioned/fake"
	knbuildv1alpha1listers "github.com/knative/build/pkg/client/listers/build/v1alpha1"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	requestv1alpha1 "github.com/projectriff/system/pkg/apis/request/v1alpha1"
	streamv1alpha1 "github.com/projectriff/system/pkg/apis/stream/v1alpha1"
	fakeprojectriffclientset "github.com/projectriff/system/pkg/client/clientset/versioned/fake"
	buildv1alpha1listers "github.com/projectriff/system/pkg/client/listers/build/v1alpha1"
	requestv1alpha1listers "github.com/projectriff/system/pkg/client/listers/request/v1alpha1"
	streamv1alpha1listers "github.com/projectriff/system/pkg/client/listers/stream/v1alpha1"
	"github.com/projectriff/system/pkg/reconciler/testing"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

var clientSetSchemes = []func(*runtime.Scheme) error{
	fakekubeclientset.AddToScheme,
	fakeprojectriffclientset.AddToScheme,
	fakeknbuildclientset.AddToScheme,
}

type Listers struct {
	sorter testing.ObjectSorter
}

func NewListers(objs []runtime.Object) Listers {
	scheme := runtime.NewScheme()

	for _, addTo := range clientSetSchemes {
		addTo(scheme)
	}

	ls := Listers{
		sorter: testing.NewObjectSorter(scheme),
	}

	ls.sorter.AddObjects(objs...)

	return ls
}

func (l *Listers) indexerFor(obj runtime.Object) cache.Indexer {
	return l.sorter.IndexerForObjectType(obj)
}

func (l *Listers) GetProjectriffObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakeprojectriffclientset.AddToScheme)
}

func (l *Listers) GetKnBuildObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakeknbuildclientset.AddToScheme)
}

func (l *Listers) GetKubeObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakekubeclientset.AddToScheme)
}

func (l *Listers) GetApplicationLister() buildv1alpha1listers.ApplicationLister {
	return buildv1alpha1listers.NewApplicationLister(l.indexerFor(&buildv1alpha1.Application{}))
}

func (l *Listers) GetFunctionLister() buildv1alpha1listers.FunctionLister {
	return buildv1alpha1listers.NewFunctionLister(l.indexerFor(&buildv1alpha1.Function{}))
}

func (l *Listers) GetHandlerLister() requestv1alpha1listers.HandlerLister {
	return requestv1alpha1listers.NewHandlerLister(l.indexerFor(&requestv1alpha1.Handler{}))
}

func (l *Listers) GetStreamLister() streamv1alpha1listers.StreamLister {
	return streamv1alpha1listers.NewStreamLister(l.indexerFor(&streamv1alpha1.Stream{}))
}

func (l *Listers) GetProcessorLister() streamv1alpha1listers.ProcessorLister {
	return streamv1alpha1listers.NewProcessorLister(l.indexerFor(&streamv1alpha1.Processor{}))
}

func (l *Listers) GetKnBuildLister() knbuildv1alpha1listers.BuildLister {
	return knbuildv1alpha1listers.NewBuildLister(l.indexerFor(&knbuildv1alpha1.Build{}))
}

func (l *Listers) GetKnClusterBuildTemplateLister() knbuildv1alpha1listers.ClusterBuildTemplateLister {
	return knbuildv1alpha1listers.NewClusterBuildTemplateLister(l.indexerFor(&knbuildv1alpha1.ClusterBuildTemplate{}))
}

func (l *Listers) GetPersistentVolumeClaimLister() corev1listers.PersistentVolumeClaimLister {
	return corev1listers.NewPersistentVolumeClaimLister(l.indexerFor(&corev1.PersistentVolumeClaim{}))
}

func (l *Listers) GetConfigMapLister() corev1listers.ConfigMapLister {
	return corev1listers.NewConfigMapLister(l.indexerFor(&corev1.ConfigMap{}))
}

func (l *Listers) GetDeploymentLister() appsv1listers.DeploymentLister {
	return appsv1listers.NewDeploymentLister(l.indexerFor(&appsv1.Deployment{}))
}

func (l *Listers) GetServiceLister() corev1listers.ServiceLister {
	return corev1listers.NewServiceLister(l.indexerFor(&corev1.Service{}))
}

func (l *Listers) GetSecretLister() corev1listers.SecretLister {
	return corev1listers.NewSecretLister(l.indexerFor(&corev1.Secret{}))
}

func (l *Listers) GetServiceAccountLister() corev1listers.ServiceAccountLister {
	return corev1listers.NewServiceAccountLister(l.indexerFor(&corev1.ServiceAccount{}))
}
