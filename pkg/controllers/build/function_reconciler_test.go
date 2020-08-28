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

	"github.com/vmware-labs/reconciler-runtime/apis"
	"github.com/vmware-labs/reconciler-runtime/reconcilers"
	rtesting "github.com/vmware-labs/reconciler-runtime/testing"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	kpackbuildv1alpha1 "github.com/projectriff/system/pkg/apis/thirdparty/kpack/build/v1alpha1"
	"github.com/projectriff/system/pkg/controllers/build"
	"github.com/projectriff/system/pkg/controllers/testing/factories"
)

func TestFunctionReconciler(t *testing.T) {
	testNamespace := "test-namespace"
	testName := "test-function"
	testKey := types.NamespacedName{Namespace: testNamespace, Name: testName}
	testImagePrefix := "example.com/repo"
	testGitUrl := "git@example.com:repo.git"
	testGitRevision := "main"
	testSha256 := "cf8b4c69d5460f88530e1c80b8856a70801f31c50b191c8413043ba9b160a43e"
	testConditionReason := "TestReason"
	testConditionMessage := "meaningful, yet concise"
	testLabelKey := "test-label-key"
	testLabelValue := "test-label-value"
	testBuildCacheName := "test-build-cache-000"
	testArtifact := "test-fn-artifact"
	testHandler := "test-fn-handler"
	testInvoker := "test-fn-invoker"

	functionConditionImageResolved := factories.Condition().Type(buildv1alpha1.FunctionConditionImageResolved)
	functionConditionKpackImageReady := factories.Condition().Type(buildv1alpha1.FunctionConditionKpackImageReady)
	functionConditionReady := factories.Condition().Type(buildv1alpha1.FunctionConditionReady)

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = kpackbuildv1alpha1.AddToScheme(scheme)
	_ = buildv1alpha1.AddToScheme(scheme)

	funcMinimal := factories.Function().
		NamespaceName(testNamespace, testName).
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Created(1)
		})
	funcValid := funcMinimal.
		Image("%s/%s", testImagePrefix, testName).
		SourceGit(testGitUrl, testGitRevision)

	kpackImageCreate := factories.KpackImage().
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Namespace(testNamespace).
				GenerateName("%s-function-", testName).
				AddLabel(buildv1alpha1.FunctionLabelKey, testName).
				ControlledBy(funcMinimal, scheme)
		}).
		Tag("%s/%s", testImagePrefix, testName).
		FunctionBuilder("", "", "").
		SourceGit(testGitUrl, testGitRevision)
	kpackImageGiven := kpackImageCreate.
		ObjectMeta(func(om factories.ObjectMeta) {
			om.
				Name("%s-function-001", testName).
				Created(1).
				Generation(1)
		}).
		StatusObservedGeneration(1)

	cmImagePrefix := factories.ConfigMap().
		NamespaceName(testNamespace, "riff-build").
		AddData("default-image-prefix", "")

	rts := rtesting.ReconcilerTestSuite{{
		Name: "function does not exist",
		Key:  testKey,
	}, {
		Name: "ignore deleted function",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			funcValid.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.Deleted(1)
				}),
		},
	}, {
		Name: "function get error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "Function"),
		},
		ShouldErr: true,
	}, {
		Name: "create kpack image",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			funcValid,
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "Created",
				`Created Image "%s-function-001"`, testName),
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			kpackImageCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcMinimal.
				StatusConditions(
					functionConditionImageResolved.True(),
					functionConditionKpackImageReady.Unknown(),
					functionConditionReady.Unknown(),
				).
				StatusKpackImageRef("%s-function-001", testName).
				StatusTargetImage("%s/%s", testImagePrefix, testName),
		},
	}, {
		Name: "create kpack image, function properties",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			funcValid.
				Artifact(testArtifact).
				Handler(testHandler).
				Invoker(testInvoker),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "Created",
				`Created Image "%s-function-001"`, testName),
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			kpackImageCreate.
				FunctionBuilder(testArtifact, testHandler, testInvoker),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcMinimal.
				StatusConditions(
					functionConditionImageResolved.True(),
					functionConditionKpackImageReady.Unknown(),
					functionConditionReady.Unknown(),
				).
				StatusKpackImageRef("%s-function-001", testName).
				StatusTargetImage("%s/%s", testImagePrefix, testName),
		},
	}, {
		Name: "create kpack image, build cache",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			funcValid.
				BuildCache("1Gi"),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "Created",
				`Created Image "%s-function-001"`, testName),
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			kpackImageCreate.
				BuildCache("1Gi"),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcMinimal.
				StatusConditions(
					functionConditionImageResolved.True(),
					functionConditionKpackImageReady.Unknown(),
					functionConditionReady.Unknown(),
				).
				StatusKpackImageRef("%s-function-001", testName).
				StatusTargetImage("%s/%s", testImagePrefix, testName),
		},
	}, {
		Name: "create kpack image, propagating labels",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			funcValid.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddLabel(testLabelKey, testLabelValue)
				}),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "Created",
				`Created Image "%s-function-001"`, testName),
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			kpackImageCreate.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddLabel(testLabelKey, testLabelValue)
				}),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcMinimal.
				StatusConditions(
					functionConditionImageResolved.True(),
					functionConditionKpackImageReady.Unknown(),
					functionConditionReady.Unknown(),
				).
				StatusKpackImageRef("%s-function-001", testName).
				StatusTargetImage("%s/%s", testImagePrefix, testName),
		},
	}, {
		Name: "default image",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			cmImagePrefix.
				AddData("default-image-prefix", testImagePrefix),
			funcMinimal.
				SourceGit(testGitUrl, testGitRevision),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "Created",
				`Created Image "%s-function-001"`, testName),
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			kpackImageCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcMinimal.
				StatusConditions(
					functionConditionImageResolved.True(),
					functionConditionKpackImageReady.Unknown(),
					functionConditionReady.Unknown(),
				).
				StatusKpackImageRef("%s-function-001", testName).
				StatusTargetImage("%s/%s", testImagePrefix, testName),
		},
	}, {
		Name: "default image, missing",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			funcMinimal.
				SourceGit(testGitUrl, testGitRevision),
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcMinimal.
				StatusConditions(
					functionConditionImageResolved.False().Reason("DefaultImagePrefixMissing", "missing default image prefix"),
					functionConditionKpackImageReady.Unknown(),
					functionConditionReady.False().Reason("DefaultImagePrefixMissing", "missing default image prefix"),
				),
		},
	}, {
		Name: "default image, undefined",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			cmImagePrefix,
			funcMinimal.
				SourceGit(testGitUrl, testGitRevision),
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcMinimal.
				StatusConditions(
					functionConditionImageResolved.False().Reason("DefaultImagePrefixMissing", "missing default image prefix"),
					functionConditionKpackImageReady.Unknown(),
					functionConditionReady.False().Reason("DefaultImagePrefixMissing", "missing default image prefix"),
				),
		},
	}, {
		Name: "default image, error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "ConfigMap"),
		},
		GivenObjects: []rtesting.Factory{
			cmImagePrefix,
			funcMinimal.
				SourceGit(testGitUrl, testGitRevision),
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcMinimal.
				StatusConditions(
					functionConditionImageResolved.False().Reason("ImageInvalid", "inducing failure for get ConfigMap"),
					functionConditionKpackImageReady.Unknown(),
					functionConditionReady.False().Reason("ImageInvalid", "inducing failure for get ConfigMap"),
				),
		},
	}, {
		Name: "kpack image ready",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			funcValid,
			kpackImageGiven.
				StatusReady().
				StatusLatestImage("%s/%s@sha256:%s", testImagePrefix, testName, testSha256),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcMinimal.
				StatusConditions(
					functionConditionImageResolved.True(),
					functionConditionKpackImageReady.True(),
					functionConditionReady.True(),
				).
				StatusKpackImageRef(kpackImageGiven.Create().GetName()).
				StatusTargetImage("%s/%s", testImagePrefix, testName).
				StatusLatestImage("%s/%s@sha256:%s", testImagePrefix, testName, testSha256),
		},
	}, {
		Name: "kpack image ready, build cache",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			funcValid,
			kpackImageGiven.
				StatusReady().
				StatusBuildCacheName(testBuildCacheName).
				StatusLatestImage("%s/%s@sha256:%s", testImagePrefix, testName, testSha256),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcMinimal.
				StatusConditions(
					functionConditionImageResolved.True(),
					functionConditionKpackImageReady.True(),
					functionConditionReady.True(),
				).
				StatusKpackImageRef(kpackImageGiven.Create().GetName()).
				StatusBuildCacheRef(testBuildCacheName).
				StatusTargetImage("%s/%s", testImagePrefix, testName).
				StatusLatestImage("%s/%s@sha256:%s", testImagePrefix, testName, testSha256),
		},
	}, {
		Name: "kpack image not-ready",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			funcValid,
			kpackImageGiven.
				StatusConditions(
					factories.Condition().Type(apis.ConditionReady).False().Reason(testConditionReason, testConditionMessage),
				).
				StatusLatestImage("%s/%s@sha256:%s", testImagePrefix, testName, testSha256),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcMinimal.
				StatusConditions(
					functionConditionImageResolved.True(),
					functionConditionKpackImageReady.False().Reason(testConditionReason, testConditionMessage),
					functionConditionReady.False().Reason(testConditionReason, testConditionMessage),
				).
				StatusKpackImageRef(kpackImageGiven.Create().GetName()).
				StatusTargetImage("%s/%s", testImagePrefix, testName).
				StatusLatestImage("%s/%s@sha256:%s", testImagePrefix, testName, testSha256),
		},
	}, {
		Name: "kpack image create error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("create", "Image"),
		},
		GivenObjects: []rtesting.Factory{
			funcValid,
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeWarning, "CreationFailed",
				`Failed to create Image "": inducing failure for create Image`),
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			kpackImageCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcValid.
				StatusConditions(
					functionConditionImageResolved.True(),
					functionConditionKpackImageReady.Unknown(),
					functionConditionReady.Unknown(),
				).
				StatusTargetImage("%s/%s", testImagePrefix, testName),
		},
	}, {
		Name: "kpack image update, spec",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			funcValid,
			kpackImageGiven.
				SourceGit(testGitUrl, "bogus"),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "Updated",
				`Updated Image "%s"`, kpackImageGiven.Create().GetName()),
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectUpdates: []rtesting.Factory{
			kpackImageGiven,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcValid.
				StatusConditions(
					functionConditionImageResolved.True(),
					functionConditionKpackImageReady.Unknown(),
					functionConditionReady.Unknown(),
				).
				StatusKpackImageRef(kpackImageGiven.Create().GetName()).
				StatusTargetImage("%s/%s", testImagePrefix, testName),
		},
	}, {
		Name: "kpack image update, labels",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			funcValid.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddLabel(testLabelKey, testLabelValue)
				}),
			kpackImageGiven,
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "Updated",
				`Updated Image "%s"`, kpackImageGiven.Create().GetName()),
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectUpdates: []rtesting.Factory{
			kpackImageGiven.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddLabel(testLabelKey, testLabelValue)
				}),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcValid.
				StatusConditions(
					functionConditionImageResolved.True(),
					functionConditionKpackImageReady.Unknown(),
					functionConditionReady.Unknown(),
				).
				StatusKpackImageRef(kpackImageGiven.Create().GetName()).
				StatusTargetImage("%s/%s", testImagePrefix, testName),
		},
	}, {
		Name: "kpack image update, fails",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("update", "Image"),
		},
		GivenObjects: []rtesting.Factory{
			funcValid,
			kpackImageGiven.
				SourceGit(testGitUrl, "bogus"),
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeWarning, "UpdateFailed",
				`Failed to update Image "%s": inducing failure for update Image`, kpackImageGiven.Create().GetName()),
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectUpdates: []rtesting.Factory{
			kpackImageGiven,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcValid.
				StatusConditions(
					functionConditionImageResolved.True(),
					functionConditionKpackImageReady.Unknown(),
					functionConditionReady.Unknown(),
				).
				StatusTargetImage("%s/%s", testImagePrefix, testName),
		},
	}, {
		Name: "kpack image list error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("list", "ImageList"),
		},
		GivenObjects: []rtesting.Factory{
			funcValid,
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcValid.
				StatusConditions(
					functionConditionImageResolved.True(),
					functionConditionKpackImageReady.Unknown(),
					functionConditionReady.Unknown(),
				).
				StatusTargetImage("%s/%s", testImagePrefix, testName),
		},
	}, {
		Name: "function status update error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("update", "Function"),
		},
		GivenObjects: []rtesting.Factory{
			funcValid,
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "Created",
				`Created Image "%s-function-001"`, testName),
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeWarning, "StatusUpdateFailed",
				`Failed to update status: inducing failure for update Function`),
		},
		ExpectCreates: []rtesting.Factory{
			kpackImageCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcMinimal.
				StatusConditions(
					functionConditionImageResolved.True(),
					functionConditionKpackImageReady.Unknown(),
					functionConditionReady.Unknown(),
				).
				StatusKpackImageRef("%s-function-001", testName).
				StatusTargetImage("%s/%s", testImagePrefix, testName),
		},
		ShouldErr: true,
	}, {
		Name: "delete extra kpack image",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			funcValid,
			kpackImageGiven.
				NamespaceName(testNamespace, "extra1"),
			kpackImageGiven.
				NamespaceName(testNamespace, "extra2"),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "Deleted",
				`Deleted Image "extra1"`),
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "Deleted",
				`Deleted Image "extra2"`),
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "Created",
				`Created Image "%s-function-001"`, testName),
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectDeletes: []rtesting.DeleteRef{
			{Group: "kpack.io", Kind: "Image", Namespace: testNamespace, Name: "extra1"},
			{Group: "kpack.io", Kind: "Image", Namespace: testNamespace, Name: "extra2"},
		},
		ExpectCreates: []rtesting.Factory{
			kpackImageCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcMinimal.
				StatusConditions(
					functionConditionImageResolved.True(),
					functionConditionKpackImageReady.Unknown(),
					functionConditionReady.Unknown(),
				).
				StatusKpackImageRef("%s-function-001", testName).
				StatusTargetImage("%s/%s", testImagePrefix, testName),
		},
	}, {
		Name: "delete extra kpack image, fails",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("delete", "Image"),
		},
		GivenObjects: []rtesting.Factory{
			funcValid,
			kpackImageGiven.
				NamespaceName(testNamespace, "extra1"),
			kpackImageGiven.
				NamespaceName(testNamespace, "extra2"),
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeWarning, "DeleteFailed",
				`Failed to delete Image "extra1": inducing failure for delete Image`),
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectDeletes: []rtesting.DeleteRef{
			{Group: "kpack.io", Kind: "Image", Namespace: testNamespace, Name: "extra1"},
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcMinimal.
				StatusConditions(
					functionConditionImageResolved.True(),
					functionConditionKpackImageReady.Unknown(),
					functionConditionReady.Unknown(),
				).
				StatusTargetImage("%s/%s", testImagePrefix, testName),
		},
	}, {
		Name: "local build",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			funcMinimal.
				Image("%s/%s", testImagePrefix, testName),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcMinimal.
				StatusConditions(
					functionConditionImageResolved.True(),
					functionConditionKpackImageReady.True(),
					functionConditionReady.True(),
				).
				StatusTargetImage("%s/%s", testImagePrefix, testName).
				// TODO resolve to a digest
				StatusLatestImage("%s/%s", testImagePrefix, testName),
		},
	}, {
		Name: "local build, removes existing build",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			funcMinimal.
				Image("%s/%s", testImagePrefix, testName),
			kpackImageGiven,
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "Deleted",
				`Deleted Image "%s"`, kpackImageGiven.Create().GetName()),
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectDeletes: []rtesting.DeleteRef{
			{Group: "kpack.io", Kind: "Image", Namespace: kpackImageGiven.Create().GetNamespace(), Name: kpackImageGiven.Create().GetName()},
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcMinimal.
				StatusConditions(
					functionConditionImageResolved.True(),
					functionConditionKpackImageReady.True(),
					functionConditionReady.True(),
				).
				StatusTargetImage("%s/%s", testImagePrefix, testName).
				// TODO resolve to a digest
				StatusLatestImage("%s/%s", testImagePrefix, testName),
		},
	}, {
		Name: "local build, removes existing build, error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("delete", "Image"),
		},
		GivenObjects: []rtesting.Factory{
			funcMinimal.
				Image("%s/%s", testImagePrefix, testName),
			kpackImageGiven,
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeWarning, "DeleteFailed",
				`Failed to delete Image "%s": inducing failure for delete Image`, kpackImageGiven.Create().GetName()),
			rtesting.NewEvent(funcValid, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectDeletes: []rtesting.DeleteRef{
			{Group: "kpack.io", Kind: "Image", Namespace: kpackImageGiven.Create().GetNamespace(), Name: kpackImageGiven.Create().GetName()},
		},
		ExpectStatusUpdates: []rtesting.Factory{
			funcMinimal.
				StatusConditions(
					functionConditionImageResolved.True(),
					functionConditionKpackImageReady.Unknown(),
					functionConditionReady.Unknown(),
				).
				StatusTargetImage("%s/%s", testImagePrefix, testName),
		},
	}}

	rts.Test(t, scheme, func(t *testing.T, rtc *rtesting.ReconcilerTestCase, c reconcilers.Config) reconcile.Reconciler {
		return build.FunctionReconciler(reconcilers.Config{
			Client:    c.Client,
			APIReader: c.APIReader,
			Recorder:  c.Recorder,
			Scheme:    c.Scheme,
			Log:       c.Log,
		})
	})
}
