# KubeCD Demo

## GKE Example Setup

This demo uses a GKE cluster created like this:

```
gcloud beta container clusters create demo-cluster \
    --machine-type n1-standard-2 \
	--num-nodes=2 \
	--zone=europe-north1-a
gcloud compute addresses create demo-nginx-ingress --region europe-north1
```

## Helm Setup

Typical Helm setup:

```
kubectl -n kube-system create serviceaccount helm-tiller
kubectl -n kube-system create clusterrolebinding helm-cluster-admin \
    --clusterrole cluster-admin --serviceaccount=kube-system:helm-tiller
helm init --service-account helm-tiller
```

## Cleanup

```
gcloud container clusters delete demo-cluster
gcloud compute addresses delete demo-nginx-ingress --region europe-north1
```
