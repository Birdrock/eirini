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
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

const (
	eiriniConfigmapLabelSelector = "cloudfoundry.org/owner=eirini"
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

	configmaps, err := clientset.CoreV1().ConfigMaps("").List(metav1.ListOptions{
		LabelSelector: eiriniConfigmapLabelSelector,
	})
	cmdcommons.ExitIfError(err)

	if len(configmaps.Items) == 0 {
		panic("eirini config not found")
	}

	eiriniNs := configmaps.Items[0].Namespace
	if err := replicateSecret(clientset,
		eiriniNs, cfg.Properties.RegistrySecretName,
		cfg.Properties.Namespace, k8s.ImagePullSecretName); err != nil {
		klog.Warningf("could not replicate app registry secret: %s", err.Error())
	}

	cancel := make(<-chan struct{})
	informer.Run(cancel)
}

func tweakListOpts(opts *metav1.ListOptions) {
	opts.LabelSelector = eiriniConfigmapLabelSelector
}

func updateConfig(clientset kubernetes.Interface, oldConfig, updatedConfig *corev1.ConfigMap) {
	oldCfg := unmarshalConfig([]byte(oldConfig.Data["opi.yml"]))
	updatedCfg := unmarshalConfig([]byte(updatedConfig.Data["opi.yml"]))

	if updatedCfg.Properties.RegistrySecretName != oldCfg.Properties.RegistrySecretName {
		srcNs := updatedConfig.Namespace
		srcSecret := updatedCfg.Properties.RegistrySecretName
		dstNs := updatedCfg.Properties.Namespace

		if err := replicateSecret(clientset, srcNs, srcSecret, dstNs, k8s.ImagePullSecretName); err != nil {
			klog.Errorf("failed to replicate registry secret: %s/%s to %s/%s: %s", srcNs, srcSecret, dstNs, k8s.ImagePullSecretName, err)
			return
		}
	}
}

func replicateSecret(clientset kubernetes.Interface, sourceNamespace, sourceSecretName, destNamespace, destSecretName string) error {
	sourceSecretsClient := clientset.CoreV1().Secrets(sourceNamespace)
	destSecretsClient := clientset.CoreV1().Secrets(destNamespace)

	sourceSecret, err := sourceSecretsClient.Get(sourceSecretName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "replicate secret: failed to get source secret: %s/%s", sourceNamespace, sourceSecretName)
	}

	destSecret := corev1.Secret{}
	destSecret.Name = k8s.ImagePullSecretName
	destSecret.Data = sourceSecret.Data
	destSecret.Type = sourceSecret.Type
	destSecret.Namespace = destNamespace
	if _, err := destSecretsClient.Create(&destSecret); err != nil {
		if k8s_errors.IsAlreadyExists(err) {
			existingSecret, err := destSecretsClient.Get(destSecretName, metav1.GetOptions{})
			if err != nil {
				return errors.Wrapf(err, "replicate secret: failed to get existing destination secret: %s/%s", sourceNamespace, sourceSecretName)
			}
			existingSecret.Data = sourceSecret.Data
			if _, err := destSecretsClient.Update(existingSecret); err != nil {
				return err
			}
			return nil
		}
		return errors.Wrapf(err, "replicate secret: failed to create new destination secret: %s/%s", destNamespace, destSecretName)
	}
	return nil
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
