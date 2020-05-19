package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/eirini"
	cmdcommons "code.cloudfoundry.org/eirini/cmd"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	nsclientset "code.cloudfoundry.org/eirini/pkg/generated/clientset/versioned"
	informers "code.cloudfoundry.org/eirini/pkg/generated/informers/externalversions"
)

type options struct {
	ConfigFile string `short:"c" long:"config" description:"Config for running staging-reporter"`
}

func main() {
	klog.SetOutput(os.Stdout)
	klog.SetOutputBySeverity("Fatal", os.Stderr)

	var opts options
	_, err := flags.ParseArgs(&opts, os.Args)
	cmdcommons.ExitIfError(err)

	cfg, err := readConfigFile(opts.ConfigFile)
	cmdcommons.ExitIfError(err)

	/*kubeClient,*/
	_, nsClient := getKubernetesClient(cfg.ConfigPath)

	informerFactory := informers.NewSharedInformerFactory(nsClient, time.Second)
	informer := informerFactory.Lrpnamespace().V1().LrpNamespaces().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(object interface{}) {
			klog.Infof("[NSController] Added: %v", object)
		},
		UpdateFunc: func(oldObject, newObject interface{}) {
			klog.Infof("[NSController] Updated: %v", newObject)
		},
		DeleteFunc: func(object interface{}) {
			klog.Infof("[NSController] Deleted: %v", object)
		},
	})

	informerFactory.Start(wait.NeverStop)

}

func getKubernetesClient(kubeConfigPath string) (kubernetes.Interface, nsclientset.Interface) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	cmdcommons.ExitIfError(err)

	namespaceClient, err := nsclientset.NewForConfig(config)
	if err != nil {
		klog.Fatalf("getClusterConfig: %v", err)
	}

	klog.Info("Successfully constructed k8s client")
	return cmdcommons.CreateKubeClient(kubeConfigPath), namespaceClient
}

func readConfigFile(path string) (*eirini.ReporterConfig, error) {
	fileBytes, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	var conf eirini.ReporterConfig
	err = yaml.Unmarshal(fileBytes, &conf)
	return &conf, errors.Wrap(err, "failed to unmarshal yaml")
}
