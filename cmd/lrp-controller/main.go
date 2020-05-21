package main

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/eirini"
	"code.cloudfoundry.org/eirini/bifrost"
	cmdcommons "code.cloudfoundry.org/eirini/cmd"
	"code.cloudfoundry.org/eirini/k8s"
	"code.cloudfoundry.org/eirini/models/cf"
	"code.cloudfoundry.org/eirini/stager/docker"
	"code.cloudfoundry.org/lager"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	lrpv1 "code.cloudfoundry.org/eirini/pkg/apis/lrp/v1"
	lrpnsclientset "code.cloudfoundry.org/eirini/pkg/generated/clientset/versioned"

	informers "code.cloudfoundry.org/eirini/pkg/generated/informers/externalversions"
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

	kubeClient, lrpnsClient := getKubernetesClients(cfg.Properties.ConfigPath)

	startInformer(cfg, kubeClient, lrpnsClient)
	<-wait.NeverStop
}

func startInformer(cfg *eirini.Config, kubeClient kubernetes.Interface, lrpClient lrpnsclientset.Interface) {
	informerFactory := informers.NewSharedInformerFactory(lrpClient, 0)
	informer := informerFactory.Eirini().V1().LRPs().Informer()

	bifrost := initLRPBifrost(kubeClient, cfg)
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(object interface{}) {
			createLRP(*bifrost, object.(*lrpv1.LRP))
		},
		DeleteFunc: func(object interface{}) {
		},
	})
	informerFactory.Start(wait.NeverStop)
}

func createLRP(bifrost bifrost.LRP, lrp *lrpv1.LRP) {
	err := bifrost.Transfer(context.Background(), cf.DesireLRPRequest{
		Namespace:               lrp.Namespace,
		GUID:                    lrp.Spec.GUID,
		Version:                 lrp.Spec.Version,
		ProcessGUID:             lrp.Spec.ProcessGUID,
		ProcessType:             lrp.Spec.ProcessType,
		AppGUID:                 lrp.Spec.AppGUID,
		AppName:                 lrp.Spec.AppName,
		SpaceGUID:               lrp.Spec.SpaceGUID,
		SpaceName:               lrp.Spec.SpaceName,
		OrganizationGUID:        lrp.Spec.OrganizationGUID,
		OrganizationName:        lrp.Spec.OrganizationName,
		PlacementTags:           lrp.Spec.PlacementTags,
		Ports:                   lrp.Spec.Ports,
		Routes:                  lrp.Spec.Routes,
		Environment:             lrp.Spec.Environment,
		EgressRules:             lrp.Spec.EgressRules,
		NumInstances:            lrp.Spec.NumInstances,
		LastUpdated:             lrp.Spec.LastUpdated,
		HealthCheckType:         lrp.Spec.HealthCheckType,
		HealthCheckHTTPEndpoint: lrp.Spec.HealthCheckHTTPEndpoint,
		HealthCheckTimeoutMs:    lrp.Spec.HealthCheckTimeoutMs,
		StartTimeoutMs:          lrp.Spec.StartTimeoutMs,
		MemoryMB:                lrp.Spec.MemoryMB,
		DiskMB:                  lrp.Spec.DiskMB,
		CPUWeight:               lrp.Spec.CPUWeight,
		VolumeMounts:            []cf.VolumeMount{},
		Lifecycle: cf.Lifecycle{
			DockerLifecycle: &cf.DockerLifecycle{
				Image:   lrp.Spec.Lifecycle.DockerLifecycle.Image,
				Command: lrp.Spec.Lifecycle.DockerLifecycle.Command,
			},
		},
		DropletHash:            lrp.Spec.DropletHash,
		DropletGUID:            lrp.Spec.DropletGUID,
		StartCommand:           lrp.Spec.StartCommand,
		UserDefinedAnnotations: lrp.Spec.UserDefinedAnnotations,
	})

	if err != nil {
		klog.Info("Failed to transfer app", err)
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

func initLRPBifrost(clientset kubernetes.Interface, cfg *eirini.Config) *bifrost.LRP {
	kubeNamespace := "delete this"
	desireLogger := lager.NewLogger("desirer")
	desireLogger.RegisterSink(lager.NewPrettySink(os.Stdout, lager.DEBUG))
	desirer := k8s.NewStatefulSetDesirer(
		clientset,
		kubeNamespace,
		cfg.Properties.RegistrySecretName,
		cfg.Properties.RootfsVersion,
		cfg.Properties.ApplicationServiceAccount,
		desireLogger,
	)
	converter := initConverter(cfg)

	return &bifrost.LRP{
		Converter: converter,
		Desirer:   desirer,
	}
}

func initConverter(cfg *eirini.Config) *bifrost.OPIConverter {
	convertLogger := lager.NewLogger("convert")
	convertLogger.RegisterSink(lager.NewPrettySink(os.Stdout, lager.DEBUG))

	stagerCfg := eirini.StagerConfig{
		EiriniAddress:   cfg.Properties.EiriniAddress,
		DownloaderImage: cfg.Properties.DownloaderImage,
		UploaderImage:   cfg.Properties.UploaderImage,
		ExecutorImage:   cfg.Properties.ExecutorImage,
	}
	return bifrost.NewOPIConverter(
		convertLogger,
		cfg.Properties.RegistryAddress,
		cfg.Properties.DiskLimitMB,
		docker.Fetch,
		docker.Parse,
		cfg.Properties.AllowRunImageAsRoot,
		stagerCfg,
	)
}

func setConfigFromFile(path string) *eirini.Config {
	fileBytes, err := ioutil.ReadFile(filepath.Clean(path))
	cmdcommons.ExitIfError(err)

	var conf eirini.Config
	conf.Properties.DiskLimitMB = 2048
	err = yaml.Unmarshal(fileBytes, &conf)
	cmdcommons.ExitIfError(err)

	return &conf
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
