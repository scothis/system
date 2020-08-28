/*
Copyright 2019 the original author or authors.

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

package build_test

import (
	"testing"

	"github.com/vmware-labs/reconciler-runtime/reconcilers"
	rtesting "github.com/vmware-labs/reconciler-runtime/testing"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kpackbuildv1alpha1 "github.com/projectriff/system/pkg/apis/thirdparty/kpack/build/v1alpha1"
	"github.com/projectriff/system/pkg/controllers/build"
	"github.com/projectriff/system/pkg/controllers/testing/factories"
)

func TestClusterBuildersReconciler(t *testing.T) {
	testNamespace := "riff-system"
	testName := "builders"
	testKey := types.NamespacedName{Namespace: testNamespace, Name: testName}
	testApplicationTag := "projectriff/builder:application"
	testFunctionTag := "projectriff/builder:function"

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = kpackbuildv1alpha1.AddToScheme(scheme)

	testApplicationBuilder := factories.KpackClusterBuilder().
		NamespaceName("", "riff-application").
		Tag(testApplicationTag)
	testApplicationBuilderReady := testApplicationBuilder.
		StatusReady().
		StatusLatestImage(testApplicationTag)
	testFunctionBuilder := factories.KpackClusterBuilder().
		NamespaceName("", "riff-function").
		Tag(testFunctionTag)
	testFunctionBuilderReady := testFunctionBuilder.
		StatusReady().
		StatusLatestImage(testFunctionTag)

	testBuilders := factories.ConfigMap().
		NamespaceName(testNamespace, testName)

	rts := rtesting.ReconcilerTestSuite{{
		Name: "builders configmap does not exist",
		Key:  testKey,
		ExpectCreates: []rtesting.Factory{
			testBuilders,
		},
	}, {
		Name: "builders configmap unchanged",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testBuilders,
		},
	}, {
		Name: "ignore deleted builders configmap",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testBuilders.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.Deleted(1)
				}),
			testApplicationBuilder,
			testFunctionBuilder,
		},
	}, {
		Name: "ignore other configmaps in the correct namespace",
		Key:  types.NamespacedName{Namespace: testNamespace, Name: "not-builders"},
		GivenObjects: []rtesting.Factory{
			testBuilders.
				NamespaceName(testNamespace, "not-builders"),
			testApplicationBuilder,
			testFunctionBuilder,
		},
	}, {
		Name: "ignore other configmaps in the wrong namespace",
		Key:  types.NamespacedName{Namespace: "not-riff-system", Name: testName},
		GivenObjects: []rtesting.Factory{
			testBuilders.
				NamespaceName("not-riff-system", testName),
			testApplicationBuilder,
			testFunctionBuilder,
		},
	}, {
		Name: "create builders configmap, not ready",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testApplicationBuilder,
			testFunctionBuilder,
		},
		ExpectCreates: []rtesting.Factory{
			testBuilders.
				AddData("riff-application", "").
				AddData("riff-function", ""),
		},
	}, {
		Name: "create builders configmap, ready",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testApplicationBuilderReady,
			testFunctionBuilderReady,
		},
		ExpectCreates: []rtesting.Factory{
			testBuilders.
				AddData("riff-application", testApplicationTag).
				AddData("riff-function", testFunctionTag),
		},
	}, {
		Name: "create builders configmap, error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("create", "ConfigMap"),
		},
		GivenObjects: []rtesting.Factory{
			testApplicationBuilder,
			testFunctionBuilder,
		},
		ShouldErr: true,
		ExpectCreates: []rtesting.Factory{
			testBuilders.
				AddData("riff-application", "").
				AddData("riff-function", ""),
		},
	}, {
		Name: "update builders configmap",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testBuilders,
			testApplicationBuilderReady,
			testFunctionBuilderReady,
		},
		ExpectUpdates: []rtesting.Factory{
			testBuilders.
				AddData("riff-application", testApplicationTag).
				AddData("riff-function", testFunctionTag),
		},
	}, {
		Name: "update builders configmap, error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("update", "ConfigMap"),
		},
		GivenObjects: []rtesting.Factory{
			testBuilders,
			testApplicationBuilderReady,
			testFunctionBuilderReady,
		},
		ShouldErr: true,
		ExpectUpdates: []rtesting.Factory{
			testBuilders.
				AddData("riff-application", testApplicationTag).
				AddData("riff-function", testFunctionTag),
		},
	}, {
		Name: "get builders configmap error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "ConfigMap"),
		},
		GivenObjects: []rtesting.Factory{
			testBuilders,
			testApplicationBuilder,
			testFunctionBuilder,
		},
		ShouldErr: true,
	}, {
		Name: "list builders error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("list", "ClusterBuilderList"),
		},
		GivenObjects: []rtesting.Factory{
			testBuilders,
			testApplicationBuilder,
			testFunctionBuilder,
		},
		ShouldErr: true,
	}}

	rts.Test(t, scheme, func(t *testing.T, rtc *rtesting.ReconcilerTestCase, c reconcilers.Config) reconcile.Reconciler {
		return &build.ClusterBuilderReconciler{
			Client:    c.Client,
			Recorder:  c.Recorder,
			Log:       c.Log,
			Namespace: testNamespace,
		}
	})
}
