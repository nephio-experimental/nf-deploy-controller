/*
Copyright 2022-2023 The Nephio Authors.

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
	"fmt"
	"net"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/nephio-project/common-lib/edge/approve"
	"github.com/nephio-project/common-lib/edge/porch"
	edgewatcher "github.com/nephio-project/edge-watcher"
	"google.golang.org/grpc"
	"k8s.io/client-go/dynamic"

	crdreader "github.com/nephio-project/nf-deploy-controller/crd-reader"
	deployment "github.com/nephio-project/nf-deploy-controller/deployment"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	nfdeployv1alpha1 "github.com/nephio-project/nf-deploy-controller/api/v1alpha1"
	"github.com/nephio-project/nf-deploy-controller/controllers"
	"github.com/nephio-project/nf-deploy-controller/hydration"
	packageservice "github.com/nephio-project/nf-deploy-controller/packageservice"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(nfdeployv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

const (
	CRD_DIRECTORY = "CRD-DIRECTORY"
)

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(
		&metricsAddr, "metrics-bind-address", ":8080",
		"The address the metric endpoint binds to.",
	)
	flag.StringVar(
		&probeAddr, "health-probe-bind-address", ":8081",
		"The address the probe endpoint binds to.",
	)
	flag.BoolVar(
		&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.",
	)
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(
		ctrl.GetConfigOrDie(), ctrl.Options{
			Scheme:                 scheme,
			MetricsBindAddress:     metricsAddr,
			Port:                   10250,
			HealthProbeBindAddress: probeAddr,
			LeaderElection:         enableLeaderElection,
			LeaderElectionID:       "leader.nephio.org",
			// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
			// when the Manager ends. This requires the binary to immediately end when the
			// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
			// speeds up voluntary leader transitions as the new leader don't have to wait
			// LeaseDuration time first.
			//
			// In the default scaffold provided, the program ends immediately after
			// the manager stops, so would be fine to enable this option. However,
			// if you are doing or is intended to do any operation such as perform cleanups
			// after the manager stops then its usage might be unsafe.
			// LeaderElectionReleaseOnCancel: true,
		},
	)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	crdDir := os.Getenv(CRD_DIRECTORY)
	setupLog.V(1).Info("reading capacity and interface profiles",
		"crdDir", crdDir)

	var crdReader crdreader.CRDReader = &crdreader.CRDSet{}
	err = crdReader.ReadCRDFiles(crdDir)
	if err != nil {
		setupLog.Error(err, "CRD Reader initialisation failed")
		os.Exit(1)
	}

	setupLog.V(1).Info("creating porch package service")

	porchClient, err := packageservice.NewPorchClient(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create porch client")
		os.Exit(1)
	}
	ps := &packageservice.PorchPackageService{
		Client: porchClient,
		Log:    ctrl.Log.WithName("PorchPackageService"),
	}
	h := &hydration.Hydration{
		PS:  ps,
		Log: ctrl.Log.WithName("Hydration"),
	}

	setupLog.V(1).Info("creating k8s rest client")

	k8sRestClient, err := approve.NewK8sRestClient(mgr.GetConfig(), porchClient.Scheme())
	if err != nil {
		setupLog.Error(err, "unable to get a rest client for porch")
		os.Exit(1)
	}

	setupLog.V(1).Info("reading edgewatcher configuration")

	var edgeWatcherConfig edgewatcher.Params
	err = envconfig.Process("", &edgeWatcherConfig)
	if err != nil {
		setupLog.Error(err, "unable to instantiate edgeWatcherConfig")
		os.Exit(1)
	}

	setupLog.V(1).Info("creating k8s dynamic client")

	edgeWatcherConfig.K8sDynamicClient, err = dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to get k8s dynamic client")
		os.Exit(1)
	}

	edgeWatcherConfig.PorchClient = porch.NewClient(ctrl.Log.WithName("PorchClient"),
		ps, k8sRestClient)

	setupLog.V(1).Info("staring edgewatcher")

	var grpcOpts []grpc.ServerOption
	edgeWatcherConfig.GRPCServer = grpc.NewServer(grpcOpts...)

	ctx := ctrl.SetupSignalHandler()
	edgewatcherLogger := ctrl.Log.WithName("EdgeWatcher")
	eventPublisher, err := edgewatcher.New(
		ctx, edgewatcherLogger, edgeWatcherConfig,
	)
	subscriberChan := eventPublisher.Subscribe()
	cancellationChan := eventPublisher.Cancel()
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", edgeWatcherConfig.Port))
	go func() {
		err := edgeWatcherConfig.GRPCServer.Serve(lis)
		if err != nil {
			setupLog.Error(err, "starting grpc server failed")
			os.Exit(1)
		}
	}()

	setupLog.V(1).Info("starting deployment")

	var deploy deployment.DeploymentManager = deployment.NewDeploymentManager(
		crdReader, subscriberChan, cancellationChan, mgr.GetClient(),
		mgr.GetClient().Status(), ctrl.Log.WithName("Deployment"),
	)

	if err = (&controllers.NfDeployReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		DeploymentManager: deploy,
		Log:               ctrl.Log.WithName("controllers").WithName("NfDeploy"),
		Hydration:         h,
		PS:                ps,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NfDeploy")
		os.Exit(1)
	}
	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		if err = (&nfdeployv1alpha1.NfDeploy{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "NfDeploy")
			os.Exit(1)
		}
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}

}
