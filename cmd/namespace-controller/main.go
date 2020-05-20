package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
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

type templateConfig struct {
	ReleaseNamespace string
	LrpNamespace     string
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

	startLrpNamespaceInformer(cfg, kubeClient, lrpnsClient)
	go runNamespaceReconciler(cfg, kubeClient, lrpnsClient)

	<-wait.NeverStop
}

func startLrpNamespaceInformer(cfg *eirini.NamespaceControllerConfig, kubeClient kubernetes.Interface, lrpnsClient lrpnsclientset.Interface) {
	informerFactory := informers.NewSharedInformerFactory(lrpnsClient, 0)
	informer := informerFactory.Lrpnamespace().V1().LrpNamespaces().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(object interface{}) {
			createNamespace(cfg, kubeClient, object.(*lrpns.LrpNamespace))
		},
		DeleteFunc: func(object interface{}) {
			deleteNamespace(kubeClient, object.(*lrpns.LrpNamespace))
		},
	})
	informerFactory.Start(wait.NeverStop)
}

func runNamespaceReconciler(cfg *eirini.NamespaceControllerConfig, kubeClient kubernetes.Interface, lrpnsClient lrpnsclientset.Interface) {
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
				createNamespace(cfg, kubeClient, &lrpns)
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

func createNamespace(cfg *eirini.NamespaceControllerConfig, kubeClient kubernetes.Interface, lrpNamespace *lrpns.LrpNamespace) {
	klog.Infof("creating namespace %q", lrpNamespace.Name)

	kubeNamespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: lrpNamespace.Name,
		},
	}
	_, err := kubeClient.CoreV1().Namespaces().Create(&kubeNamespace)
	if err != nil {
		klog.Errorf("failed to create namespace %q: %v", lrpNamespace.Name, err)
		return
	}

	templateCfg := templateConfig{
		ReleaseNamespace: cfg.ReleaseNamespace,
		LrpNamespace:     lrpNamespace.Name,
	}
	if err := createTemplatedObjects(templateCfg, cfg.TemplatesPath); err != nil {
		klog.Errorf("failed to create templated objects: %v", err)
		return
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

func readConfigFile(path string) (*eirini.NamespaceControllerConfig, error) {
	fileBytes, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	var conf eirini.NamespaceControllerConfig
	err = yaml.Unmarshal(fileBytes, &conf)
	return &conf, errors.Wrap(err, "failed to unmarshal yaml")
}

func createTemplatedObjects(templateCfg templateConfig, templatePath string) error {
	return filepath.Walk(templatePath, func(path string, fInfo os.FileInfo, errIn error) error {
		if errIn != nil {
			return errIn
		}

		// TODO: Think about a better way to filter out non template entries, such a dirs, random metadata files, etc.
		if filepath.Ext(fInfo.Name()) != ".yml" {
			klog.Infof("will not generate objects for file %s", filepath.Join(templatePath, fInfo.Name()))
			return nil
		}

		if err := createTemplatedObject(filepath.Join(templatePath, fInfo.Name()), templateCfg); err != nil {
			return fmt.Errorf("failed to create templated object for template file %s: %v", fInfo.Name(), err)
		}
		return nil
	})
}

func createTemplatedObject(templateFilePath string, templateCfg templateConfig) error {
	tmpl, err := template.New(templateFilePath).ParseFiles(templateFilePath)
	if err != nil {
		return err
	}

	var buffer strings.Builder
	if err := tmpl.Execute(&buffer, templateCfg); err != nil {
		return err
	}

	klog.Infof("I would have created the following object: \n%s\n", buffer.String())
	return nil
}
