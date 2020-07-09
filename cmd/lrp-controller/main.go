package main

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/eirini"
	"code.cloudfoundry.org/eirini/bifrost"
	cmdcommons "code.cloudfoundry.org/eirini/cmd"
	"code.cloudfoundry.org/eirini/models/cf"
	"code.cloudfoundry.org/lager"
	"github.com/jessevdk/go-flags"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	eiriniv1 "code.cloudfoundry.org/eirini/pkg/apis/lrp/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type options struct {
	ConfigFile string `short:"c" long:"config" description:"Config for running lrp-controller"`
}

func main() {
	var opts options
	_, err := flags.ParseArgs(&opts, os.Args)
	cmdcommons.ExitIfError(err)

	eiriniCfg, err := readConfigFile(opts.ConfigFile)
	cmdcommons.ExitIfError(err)

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", eiriniCfg.Properties.ConfigPath)
	cmdcommons.ExitIfError(err)

	client, err := client.New(kubeConfig, client.Options{})
	cmdcommons.ExitIfError(err)

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	cmdcommons.ExitIfError(err)

	logger := lager.NewLogger("lrp-informer")
	logger.RegisterSink(lager.NewPrettySink(os.Stdout, lager.DEBUG))

	bifrost := cmdcommons.InitLRPBifrost(clientset, eiriniCfg)

	lrpReconciler := NewLRPReconciler(client, bifrost)
	// lrpController := controller.NewRestLrp(httpClient, eiriniURI)
	// informer := lrp.NewInformer(logger, lrpClientset, lrpController)
	// informer.Start()

	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	cmdcommons.ExitIfError(err)

	err = builder.
		ControllerManagedBy(mgr).
		For(&eiriniv1.LRP{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(lrpReconciler)
	cmdcommons.ExitIfError(err)
}

func NewLRPReconciler(client client.Client, lrpBifrost *bifrost.LRP) *LRPReconciler {
	return &LRPReconciler{
		client:     client,
		lrpBifrost: lrpBifrost,
	}
}

type LRPReconciler struct {
	client     client.Client
	lrpBifrost *bifrost.LRP
}

func (c *LRPReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	lrp := &eiriniv1.LRP{}
	err := c.client.Get(context.Background(), request.NamespacedName, lrp)
	if err != nil {
		return reconcile.Result{}, err
	}

	var createReq cf.DesireLRPRequest
	if err := copier.Copy(&createReq, &lrp.Spec); err != nil {
		return reconcile.Result{}, err
	}

	createReq.Namespace = lrp.Namespace

	err = c.lrpBifrost.Reconcile(context.Background(), createReq, lrp.Spec.LastUpdated)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func readConfigFile(path string) (*eirini.Config, error) {
	fileBytes, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	var conf eirini.Config
	err = yaml.Unmarshal(fileBytes, &conf)
	return &conf, errors.Wrap(err, "failed to unmarshal yaml")
}
