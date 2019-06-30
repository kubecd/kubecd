package model

type GkeProvider struct {
	Project     string  `json:"project"`
	ClusterName string  `json:"clusterName"`
	Zone        *string `json:"zone"`
	Region      *string `json:"region"`
}

type AksProvider struct {
	ResourceGroup string `json:"resourceGroup"`
	ClusterName   string `json:"clusterName"`
}

type MinikubeProvider struct {
}

type DockerForDesktopProvider struct {
}

type ExistingContextProvider struct {
	ContextName string `json:"contextName"`
}

type Provider struct {
	GKE              *GkeProvider              `json:"gke,omitempty"`
	Minikube         *MinikubeProvider         `json:"minikube,omitempty"`
	AKS              *AksProvider              `json:"aks,omitempty"`
	DockerForDesktop *DockerForDesktopProvider `json:"dockerForDesktop,omitempty"`
	ExistingContext  *ExistingContextProvider  `json:"existingContext,omitempty"`
}
