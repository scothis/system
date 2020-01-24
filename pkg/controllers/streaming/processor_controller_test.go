/*
Copyright 2020 the original author or authors.

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

package streaming_test

import (
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	streamingv1alpha1 "github.com/projectriff/system/pkg/apis/streaming/v1alpha1"
	"github.com/projectriff/system/pkg/controllers"
	"github.com/projectriff/system/pkg/controllers/streaming"
	rtesting "github.com/projectriff/system/pkg/controllers/testing"
	"github.com/projectriff/system/pkg/controllers/testing/factories"
	"github.com/projectriff/system/pkg/tracker"
)

func TestProcessorReconciler(t *testing.T) {
	testNamespace := "test-namespace"
	testSystemNamespace := "system-namespace"
	testName := "test-processor"
	testKey := types.NamespacedName{Namespace: testNamespace, Name: testName}
	testImagePrefix := "example.com/repo"
	testSha256 := "cf8b4c69d5460f88530e1c80b8856a70801f31c50b191c8413043ba9b160a43e"
	testImage := fmt.Sprintf("%s@sha256:%s", testImagePrefix, testSha256)

	testFunction := factories.Function().
		NamespaceName(testNamespace, "my-function").
		StatusLatestImage(testImage)
	testContainer := factories.Container().
		NamespaceName(testNamespace, "my-container").
		StatusLatestImage(testImage)

	processorConditionDeploymentReady := factories.Condition().Type(streamingv1alpha1.ProcessorConditionDeploymentReady)
	processorConditionReady := factories.Condition().Type(streamingv1alpha1.ProcessorConditionReady)
	processorConditionScaledObjectReady := factories.Condition().Type(streamingv1alpha1.ProcessorConditionScaledObjectReady)
	processorConditionStreamsReady := factories.Condition().Type(streamingv1alpha1.ProcessorConditionStreamsReady)
	deploymentConditionAvailable := factories.Condition().Type("Available")
	deploymentConditionProgressing := factories.Condition().Type("Progressing")
	scaledObjectConditionReady := factories.Condition().Type("Ready")
	streamConditionAvailable := factories.Condition().Type(streamingv1alpha1.StreamConditionReady)

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = streamingv1alpha1.AddToScheme(scheme)

	processorMinimal := factories.Processor().
		NamespaceName(testNamespace, testName).
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Created(1)
			om.Generation(1)
		})
	processor := processorMinimal.
		StatusLatestImage(testImage).
		StatusDeploymentRef("%s-processor-000", testName).
		StatusScaledObjectRef("%s-processor-000", testName)
	processorReady := processor.
		StatusObservedGeneration(1).
		StatusConditions(
			processorConditionDeploymentReady.True(),
			processorConditionReady.True(),
			processorConditionScaledObjectReady.True(),
			processorConditionStreamsReady.True(),
		)

	deploymentCreate := factories.Deployment().
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Namespace(testNamespace)
			om.GenerateName("%s-processor-", testName)
			om.AddLabel(streamingv1alpha1.ProcessorLabelKey, testName)
			om.ControlledBy(processor, scheme)
		}).
		AddSelectorLabel(streamingv1alpha1.ProcessorLabelKey, testName).
		PodTemplateSpec(func(pts factories.PodTemplateSpec) {
			pts.ContainerNamed("test", func(c *corev1.Container) {
				c.Image = "scratch"
			})
		})
	deploymentGiven := deploymentCreate.
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Name("%s-processor-000", testName)
			om.Created(1)
		})

	table := rtesting.Table{{
		Name: "processor does not exist",
		Key:  testKey,
	}, {
		Name: "ignore deleted processor",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			processor.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.Deleted(1)
				}),
		},
	}, {
		Name: "error fetching processor",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "Processor"),
		},
		GivenObjects: []rtesting.Factory{
			processor,
		},
		ShouldErr: true,
	}}

	table.Test(t, scheme, func(t *testing.T, row *rtesting.Testcase, client client.Client, tracker tracker.Tracker, recorder record.EventRecorder, log logr.Logger) reconcile.Reconciler {
		return streaming.ProcessorReconciler(
			controllers.Config{
				Client:   client,
				Recorder: recorder,
				Log:      log,
				Scheme:   scheme,
				Tracker:  tracker,
			},
			testSystemNamespace,
		)
	})

	_ = testFunction
	_ = testContainer
	_ = processorReady
	_ = deploymentConditionAvailable
	_ = deploymentConditionProgressing
	_ = scaledObjectConditionReady
	_ = streamConditionAvailable
	_ = deploymentGiven
}
