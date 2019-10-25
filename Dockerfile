# Generate Thrift sources
FROM golang:1.13 AS build
RUN mkdir /src
COPY . /src/
RUN cd /src; CGO_ENABLED=0 make build

# Grab binary from build step and install in a clean Python image,
# along with kubectl, helm, gcloud, ssh and git
FROM debian:buster
ARG KUBECTL_VERSION=1.16.2
ARG HELM_VERSION=2.9.1
ARG GCLOUD_VERSION=268.0.0
COPY --from=build /src/kcd /usr/local/bin/kcd
RUN apt-get update && apt-get install -y openssh-client git procps curl && apt-get clean
RUN curl -Ls https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl \
  > /usr/local/bin/kubectl \
 && chmod +x /usr/local/bin/kubectl
RUN curl -Ls https://kubernetes-helm.storage.googleapis.com/helm-v${HELM_VERSION}-linux-amd64.tar.gz \
  | tar -xOzf - linux-amd64/helm > /usr/local/bin/helm \
 && chmod +x /usr/local/bin/helm
RUN curl https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GCLOUD_VERSION}-linux-x86_64.tar.gz \
  | tar -C /usr/local -xzf -
ENV PATH=/usr/local/bin:/usr/local/google-cloud-sdk/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin
ENTRYPOINT ["/usr/local/bin/kcd"]
