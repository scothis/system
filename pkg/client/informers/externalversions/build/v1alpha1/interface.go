/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package v1alpha1

import (
	internalinterfaces "github.com/projectriff/system/pkg/client/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// Applications returns a ApplicationInformer.
	Applications() ApplicationInformer
	// Containers returns a ContainerInformer.
	Containers() ContainerInformer
	// Functions returns a FunctionInformer.
	Functions() FunctionInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// Applications returns a ApplicationInformer.
func (v *version) Applications() ApplicationInformer {
	return &applicationInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Containers returns a ContainerInformer.
func (v *version) Containers() ContainerInformer {
	return &containerInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Functions returns a FunctionInformer.
func (v *version) Functions() FunctionInformer {
	return &functionInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
