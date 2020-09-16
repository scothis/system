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

package v1alpha1

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/vmware-labs/reconciler-runtime/validation"
	corev1 "k8s.io/api/core/v1"
)

func TestValidateProcessor(t *testing.T) {
	for _, c := range []struct {
		name     string
		target   *Processor
		expected validation.FieldErrors
	}{{
		name:     "empty",
		target:   &Processor{},
		expected: validation.ErrMissingField("spec"),
	}, {
		name: "valid",
		target: &Processor{
			Spec: ProcessorSpec{
				Inputs: []InputStreamBinding{
					{Stream: "my-stream", Alias: "in"},
				},
				Template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "function",
								Image: "registry.example.com/my-func",
							},
						},
					},
				},
			},
		},
		expected: validation.FieldErrors{},
	}} {
		t.Run(c.name, func(t *testing.T) {
			actual := c.target.Validate()
			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("validateProcessor(%s) (-expected, +actual) = %v", c.name, diff)
			}
		})
	}
}

func TestValidateProcessorSpec(t *testing.T) {
	for _, c := range []struct {
		name     string
		target   *ProcessorSpec
		expected validation.FieldErrors
	}{{
		name:     "empty",
		target:   &ProcessorSpec{},
		expected: validation.ErrMissingField(validation.CurrentField),
	}, {
		name: "valid",
		target: &ProcessorSpec{
			Inputs: []InputStreamBinding{
				{Stream: "my-stream", Alias: "in"},
			},
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "function"},
					},
				},
			},
		},
		expected: validation.FieldErrors{},
	}, {
		name: "requires inputs",
		target: &ProcessorSpec{
			Inputs: nil,
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "function"},
					},
				},
			},
		},
		expected: validation.ErrMissingField("inputs"),
	}, {
		name: "empty input",
		target: &ProcessorSpec{
			Inputs: []InputStreamBinding{
				{},
			},
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "function"},
					},
				},
			},
		},
		expected: validation.FieldErrors{}.Also(
			validation.ErrMissingField("inputs[0].stream"),
			validation.ErrMissingField("inputs[0].alias"),
		),
	}, {
		name: "valid input",
		target: &ProcessorSpec{
			Inputs: []InputStreamBinding{
				{Stream: "my-stream", Alias: "in"},
			},
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "function"},
					},
				},
			},
		},
		expected: validation.FieldErrors{},
	}, {
		name: "empty output",
		target: &ProcessorSpec{
			Inputs: []InputStreamBinding{
				{Stream: "my-stream", Alias: "in"},
			},
			Outputs: []OutputStreamBinding{
				{},
			},
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "function"},
					},
				},
			},
		},
		expected: validation.FieldErrors{}.Also(
			validation.ErrMissingField("outputs[0].stream"),
			validation.ErrMissingField("outputs[0].alias"),
		),
	}, {
		name: "valid output",
		target: &ProcessorSpec{
			Inputs: []InputStreamBinding{
				{Stream: "my-stream", Alias: "my-input"},
			},
			Outputs: []OutputStreamBinding{
				{Stream: "my-stream", Alias: "my-output"},
			},
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "function"},
					},
				},
			},
		},
		expected: validation.FieldErrors{},
	}, {
		name: "valid offsets",
		target: &ProcessorSpec{
			Inputs: []InputStreamBinding{
				{Stream: "my-stream", Alias: "in1", StartOffset: Latest},
				{Stream: "my-stream", Alias: "in2", StartOffset: Earliest},
			},
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "function"},
					},
				},
			},
		},
		expected: validation.FieldErrors{},
	}, {
		name: "invalid offset",
		target: &ProcessorSpec{
			Inputs: []InputStreamBinding{
				{Stream: "my-stream", Alias: "my-input", StartOffset: "42"},
			},
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "function"},
					},
				},
			},
		},
		expected: validation.ErrInvalidValue("42", "inputs[0].startOffset"),
	}, {
		name: "input alias collision",
		target: &ProcessorSpec{
			Inputs: []InputStreamBinding{
				{Stream: "my-stream1", Alias: "my-input"},
				{Stream: "my-stream2", Alias: "my-input"},
			},
			Outputs: []OutputStreamBinding{},
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "function"},
					},
				},
			}},
		expected: validation.ErrDuplicateValue("my-input", "inputs[0].alias", "inputs[1].alias"),
	}, {
		name: "output alias collision",
		target: &ProcessorSpec{
			Inputs: []InputStreamBinding{
				{Stream: "my-stream1", Alias: "my-input"},
			},
			Outputs: []OutputStreamBinding{
				{Stream: "my-output-stream1", Alias: "my-output"},
				{Stream: "my-output-stream2", Alias: "my-output"},
			},
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "function"},
					},
				},
			}},
		expected: validation.ErrDuplicateValue("my-output", "outputs[0].alias", "outputs[1].alias"),
	}, {
		name: "allow duplicates across input/output aliases",
		target: &ProcessorSpec{
			Inputs: []InputStreamBinding{
				{Stream: "my-stream1", Alias: "duplicate-alias"},
			},
			Outputs: []OutputStreamBinding{
				{Stream: "my-stream2", Alias: "duplicate-alias"},
			},
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "function"},
					},
				},
			},
		},
		expected: validation.FieldErrors{},
	}, {
		name: "invalid container name",
		target: &ProcessorSpec{
			Inputs: []InputStreamBinding{
				{Stream: "my-stream", Alias: "my-input"},
			},
			Outputs: []OutputStreamBinding{
				{Stream: "my-stream", Alias: "my-output"},
			},
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "processor"},
					},
				},
			},
		},
		expected: validation.ErrInvalidValue("processor", "template.spec.containers[0].name"),
	}} {
		t.Run(c.name, func(t *testing.T) {
			actual := c.target.Validate()
			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("validateProcessorSpec(%s) (-expected, +actual) = %v", c.name, diff)
			}
		})
	}
}
