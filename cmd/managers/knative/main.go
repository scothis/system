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

package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	knativev1alpha1 "github.com/projectriff/system/pkg/apis/knative/v1alpha1"
	servingv1 "github.com/projectriff/system/pkg/apis/thirdparty/knative/serving/v1"
	controllers "github.com/projectriff/system/pkg/controllers/knative"
	"github.com/projectriff/system/pkg/tracker"
	// +kubebuilder:scaffold:imports
)

var (
	scheme     = runtime.NewScheme()
	setupLog   = ctrl.Log.WithName("setup")
	syncPeriod = 10 * time.Hour
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = knativev1alpha1.AddToScheme(scheme)
	_ = buildv1alpha1.AddToScheme(scheme)
	_ = servingv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var probesAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probesAddr, "probes-addr", ":8081", "The address health probes bind to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.Logger(true))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		HealthProbeBindAddress: probesAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "controller-leader-election-helper-knative",
		SyncPeriod:             &syncPeriod,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.AdapterReconciler{
		Client:   mgr.GetClient(),
		Recorder: mgr.GetEventRecorderFor("Adapter"),
		Log:      ctrl.Log.WithName("controllers").WithName("Adapter"),
		Scheme:   mgr.GetScheme(),
		Tracker:  tracker.New(syncPeriod, ctrl.Log.WithName("controllers").WithName("Adapter").WithName("tracker")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Adapter")
		os.Exit(1)
	}
	if err = ctrl.NewWebhookManagedBy(mgr).For(&knativev1alpha1.Adapter{}).Complete(); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Adapter")
		os.Exit(1)
	}
	if err = (&controllers.DeployerReconciler{
		Client:   mgr.GetClient(),
		Recorder: mgr.GetEventRecorderFor("Deployer"),
		Log:      ctrl.Log.WithName("controllers").WithName("Deployer"),
		Scheme:   mgr.GetScheme(),
		Tracker:  tracker.New(syncPeriod, ctrl.Log.WithName("controllers").WithName("Deployer").WithName("tracker")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Deployer")
		os.Exit(1)
	}
	if err = ctrl.NewWebhookManagedBy(mgr).For(&knativev1alpha1.Deployer{}).Complete(); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Deployer")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("default", func(_ *http.Request) error { return nil }); err != nil {
		setupLog.Error(err, "unable to create health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("default", func(_ *http.Request) error { return nil }); err != nil {
		setupLog.Error(err, "unable to create ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
