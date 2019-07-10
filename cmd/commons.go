package cmd

import (
	"flag"
	"os"

	eiriniclientset "code.cloudfoundry.org/eirini/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"

	// Kubernetes has a tricky way to add authentication
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func CreateEiriniClient(kubeConfigPath string) eiriniclientset.Interface {
	klog.SetOutput(os.Stdout)
	klog.SetOutputBySeverity("Fatal", os.Stderr)
	klog.SetOutputBySeverity("Info", os.Stderr)
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	ExitWithError(err)

	client, err := eiriniclientset.NewForConfig(config)
	ExitWithError(err)

	return client
}

func CreateMetricsClient(kubeConfigPath string) metricsclientset.Interface {
	klog.SetOutput(os.Stdout)
	klog.SetOutputBySeverity("Fatal", os.Stderr)
	klog.SetOutputBySeverity("Info", os.Stderr)
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	ExitWithError(err)

	metricsClient, err := metricsclientset.NewForConfig(config)
	ExitWithError(err)

	return metricsClient
}

func CreateKubeClient(kubeConfigPath string) kubernetes.Interface {
	klog.SetOutput(os.Stdout)
	klog.SetOutputBySeverity("Fatal", os.Stderr)
	klog.SetOutputBySeverity("Info", os.Stderr)
	klog.InitFlags(nil)
	flag.Set("v", "10")
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	ExitWithError(err)

	clientset, err := kubernetes.NewForConfig(config)
	ExitWithError(err)

	return clientset
}

func ExitWithError(err error) {
	if err != nil {
		panic(err)
	}
}
