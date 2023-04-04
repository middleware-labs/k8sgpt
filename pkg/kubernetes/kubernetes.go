package kubernetes

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	Client kubernetes.Interface
}

func (c *Client) GetClient() kubernetes.Interface {
	return c.Client
}

func NewClient(kubecontext string, kubeconfig string) (*Client, error) {

	// Running as In-Cluster config if kubecontext and kubeconfig are empty
	if kubecontext == "" && kubeconfig == "" {
		// create the clientset config using incluster config
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}

		// create the clientset
		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, err
		}

		return &Client{
			Client: clientSet,
		}, nil
	} else {
		config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
			&clientcmd.ConfigOverrides{
				CurrentContext: kubecontext,
			})
		// create the clientset
		c, err := config.ClientConfig()
		if err != nil {
			return nil, err
		}
		clientSet, err := kubernetes.NewForConfig(c)
		if err != nil {
			return nil, err
		}

	return &Client{
		Client: clientSet,
	}, nil
}
