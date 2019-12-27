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

package factories

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/projectriff/system/pkg/apis"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	rtesting "github.com/projectriff/system/pkg/controllers/testing"
	"github.com/projectriff/system/pkg/refs"
)

type application struct {
	target *buildv1alpha1.Application
}

func Application(seed ...*buildv1alpha1.Application) *application {
	var target *buildv1alpha1.Application
	switch len(seed) {
	case 0:
		target = &buildv1alpha1.Application{}
	case 1:
		target = seed[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one seed, got %v", seed))
	}
	return &application{
		target: target,
	}
}

func (f *application) deepCopy() *application {
	return Application(f.target.DeepCopy())
}

func (f *application) Get() *buildv1alpha1.Application {
	return f.deepCopy().target
}

func (f *application) Mutate(m func(*buildv1alpha1.Application)) *application {
	f = f.deepCopy()
	m(f.target)
	return f
}

func (f *application) NamespaceName(namespace, name string) *application {
	return f.Mutate(func(app *buildv1alpha1.Application) {
		app.ObjectMeta.Namespace = namespace
		app.ObjectMeta.Name = name
	})
}

func (f *application) ObjectMeta(nf func(ObjectMeta)) *application {
	return f.Mutate(func(app *buildv1alpha1.Application) {
		omf := objectMeta(app.ObjectMeta)
		nf(omf)
		app.ObjectMeta = omf.Get()
	})
}

func (f *application) Image(format string, a ...interface{}) *application {
	return f.Mutate(func(app *buildv1alpha1.Application) {
		app.Spec.Image = fmt.Sprintf(format, a...)
	})
}

func (f *application) SourceGit(url string, revision string) *application {
	return f.Mutate(func(app *buildv1alpha1.Application) {
		if app.Spec.Source == nil {
			app.Spec.Source = &buildv1alpha1.Source{}
		}
		app.Spec.Source = &buildv1alpha1.Source{
			Git: &buildv1alpha1.Git{
				URL:      url,
				Revision: revision,
			},
			SubPath: app.Spec.Source.SubPath,
		}
	})
}

func (f *application) SourceSubPath(subpath string) *application {
	return f.Mutate(func(app *buildv1alpha1.Application) {
		if app.Spec.Source == nil {
			app.Spec.Source = &buildv1alpha1.Source{}
		}
		app.Spec.Source.SubPath = subpath
	})
}

func (f *application) BuildCache(quantity string) *application {
	return f.Mutate(func(app *buildv1alpha1.Application) {
		size, err := resource.ParseQuantity(quantity)
		if err != nil {
			panic(err)
		}
		app.Spec.CacheSize = &size
	})
}

func (f *application) StatusConditions(conditions ...apis.Condition) *application {
	return f.Mutate(func(app *buildv1alpha1.Application) {
		app.Status.Conditions = conditions
	})
}

func (f *application) StatusObservedGeneration(generation int64) *application {
	return f.Mutate(func(app *buildv1alpha1.Application) {
		app.Status.ObservedGeneration = generation
	})
}

func (f *application) StatusKpackImageRef(format string, a ...interface{}) *application {
	return f.Mutate(func(app *buildv1alpha1.Application) {
		app.Status.KpackImageRef = &refs.TypedLocalObjectReference{
			APIGroup: rtesting.StringPtr("build.pivotal.io"),
			Kind:     "Image",
			Name:     fmt.Sprintf(format, a...),
		}
	})
}

func (f *application) StatusBuildCacheRef(format string, a ...interface{}) *application {
	return f.Mutate(func(app *buildv1alpha1.Application) {
		app.Status.BuildCacheRef = &refs.TypedLocalObjectReference{
			Kind: "PersistentVolumeClaim",
			Name: fmt.Sprintf(format, a...),
		}
	})
}

func (f *application) StatusTargetImage(format string, a ...interface{}) *application {
	return f.Mutate(func(app *buildv1alpha1.Application) {
		app.Status.TargetImage = fmt.Sprintf(format, a...)
	})
}

func (f *application) StatusLatestImage(format string, a ...interface{}) *application {
	return f.Mutate(func(app *buildv1alpha1.Application) {
		app.Status.LatestImage = fmt.Sprintf(format, a...)
	})
}
