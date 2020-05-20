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

	lrpns "code.cloudfoundry.org/eirini/pkg/apis/lrpnamespace/v1"
	lrpnsclientset "code.cloudfoundry.org/eirini/pkg/generated/clientset/versioned"

	informers "code.cloudfoundry.org/eirini/pkg/generated/informers/externalversions"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	kubeClient, lrpnsClient := getKubernetesClients(cfg.ConfigPath)

	startLrpNamespaceInformer(kubeClient, lrpnsClient)
	go startNamespaceReconciler(kubeClient, lrpnsClient)

	<-wait.NeverStop
}

func startLrpNamespaceInformer(kubeClient kubernetes.Interface, lrpnsClient lrpnsclientset.Interface) {
	informerFactory := informers.NewSharedInformerFactory(lrpnsClient, 0)
	informer := informerFactory.Lrpnamespace().V1().LrpNamespaces().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(object interface{}) {
			createNamespace(kubeClient, object.(*lrpns.LrpNamespace))
		},
		DeleteFunc: func(object interface{}) {
			deleteNamespace(kubeClient, object.(*lrpns.LrpNamespace))
		},
	})
	informerFactory.Start(wait.NeverStop)
}

func startNamespaceReconciler(kubeClient kubernetes.Interface, lrpnsClient lrpnsclientset.Interface) {
	for range time.Tick(5 * time.Second) {
		lrpNamespaces, err := lrpnsClient.LrpnamespaceV1().LrpNamespaces().List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("failed to list lrpnamespaces: %v", err)
			continue
		}
		for _, lrpns := range lrpNamespaces.Items {
			exists, err := kubeNsExists(kubeClient, lrpns.Name)
			if err != nil {
				klog.Errorf("failed to check whether kube namespace %q exists: %v", lrpns.Name, err)
				continue
			}

			if !exists {
				createNamespace(kubeClient, &lrpns)
				continue
			}
		}
	}
}

func kubeNsExists(kubeClient kubernetes.Interface, namespace string) (bool, error) {
	kubens, err := kubeClient.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
	if err == nil {
		return kubens != nil, nil
	}
	var statusErr k8serrors.StatusError
	if errors.As(err, statusErr) && statusErr.Status().Reason == metav1.StatusReasonNotFound {
		return false, nil
	}

	return err == nil, err
}

func createNamespace(kubeClient kubernetes.Interface, lrpNamespace *lrpns.LrpNamespace) {
	klog.Infof("creating namespace %q", lrpNamespace.Name)

	kubeNamespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: lrpNamespace.Name,
		},
	}
	_, err := kubeClient.CoreV1().Namespaces().Create(&kubeNamespace)
	if err != nil {
		klog.Errorf("failed to create namespace %q: %v", lrpNamespace.Name, err)
	}
}

func deleteNamespace(kubeClient kubernetes.Interface, lrpNamespace *lrpns.LrpNamespace) {
	klog.Infof("deleting namespace %q", lrpNamespace.Name)

	if err := kubeClient.CoreV1().Namespaces().Delete(lrpNamespace.Name, &metav1.DeleteOptions{}); err != nil {
		klog.Errorf("failed to delete namespace %q: %v", lrpNamespace.Name, err)
	}
}

func getKubernetesClients(kubeConfigPath string) (kubernetes.Interface, lrpnsclientset.Interface) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	cmdcommons.ExitIfError(err)

	namespaceClient, err := lrpnsclientset.NewForConfig(config)
	if err != nil {
		klog.Fatalf("getClusterConfig: %v", err)
	}

	klog.Info("Successfully constructed lrpnamespace client")

	kubeClient := cmdcommons.CreateKubeClient(kubeConfigPath)
	klog.Info("Successfully constructed kube client")

	return kubeClient, namespaceClient
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
