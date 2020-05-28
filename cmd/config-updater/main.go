package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/eirini"
	cmdcommons "code.cloudfoundry.org/eirini/cmd"
	"code.cloudfoundry.org/eirini/k8s"
	"code.cloudfoundry.org/lager"
	"github.com/jessevdk/go-flags"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

type options struct {
	ConfigFile string `short:"c" long:"config" description:"Config for running route-collector" required:"true"`
}

func main() {
	var opts options
	_, err := flags.ParseArgs(&opts, os.Args)
	cmdcommons.ExitIfError(err)

	cfg := loadConfigFromFile(opts.ConfigFile)

	logger := lager.NewLogger("configmap-informer")
	logger.RegisterSink(lager.NewPrettySink(os.Stdout, lager.DEBUG))

	clientset := cmdcommons.CreateKubeClient(cfg.Properties.ConfigPath)

	factory := informers.NewSharedInformerFactoryWithOptions(clientset,
		0,
		informers.WithTweakListOptions(tweakListOpts))

	informer := factory.Core().V1().ConfigMaps().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, updatedObj interface{}) {
			oldConfig := oldObj.(*corev1.ConfigMap)
			updatedConfig := updatedObj.(*corev1.ConfigMap)
			updateConfig(clientset, oldConfig, updatedConfig)
		},
	})

	cancel := make(<-chan struct{})
	informer.Run(cancel)
}

func tweakListOpts(opts *metav1.ListOptions) {
	opts.LabelSelector = "cloudfoundry.org/owner=eirini"
}

func updateConfig(clientset kubernetes.Interface, oldConfig, updatedConfig *corev1.ConfigMap) {
	oldCfg := unmarshalConfig([]byte(oldConfig.Data["opi.yml"]))
	updatedCfg := unmarshalConfig([]byte(updatedConfig.Data["opi.yml"]))

	if updatedCfg.Properties.RegistrySecretName != oldCfg.Properties.RegistrySecretName {
		sourceSecretsClient := clientset.CoreV1().Secrets(updatedConfig.Namespace)
		destSecretsClient := clientset.CoreV1().Secrets(updatedCfg.Properties.Namespace)

		secret, err := sourceSecretsClient.Get(updatedCfg.Properties.RegistrySecretName, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("failed to get registry secret: %s", err)
			return
		}

		registrySecret := corev1.Secret{}
		registrySecret.Name = k8s.ImagePullSecretName
		registrySecret.Data = secret.Data
		registrySecret.Type = secret.Type
		registrySecret.Namespace = updatedCfg.Properties.Namespace
		if _, err := destSecretsClient.Create(&registrySecret); err != nil {
			klog.Errorf("failed to create registry secret: %s", err)
			return
		}
	}

}

func loadConfigFromFile(path string) *eirini.Config {
	fileBytes, err := ioutil.ReadFile(filepath.Clean(path))
	cmdcommons.ExitIfError(err)

	return unmarshalConfig(fileBytes)
}

func unmarshalConfig(data []byte) *eirini.Config {
	var conf eirini.Config
	conf.Properties.DiskLimitMB = 2048
	err := yaml.Unmarshal(data, &conf)
	cmdcommons.ExitIfError(err)

	return &conf
}
