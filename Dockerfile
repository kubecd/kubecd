# Generate Thrift sources
FROM golang:1.12 AS build
RUN mkdir /src
COPY . /src/
RUN CGO_ENABLED=0 go build -o /tmp/kcd ./cmd/kcd

# Grab binary from build step and install in a clean Python image,
# along with kubectl, helm, gcloud, ssh and git
FROM debian:stretch-slim
ARG KUBECTL_VERSION=1.13.7
ARG HELM_VERSION=2.9.1
ARG GCLOUD_VERSION=258.0.0
COPY --from=build /tmp/kcd /usr/local/bin/kcd
RUN curl -Ls https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl \
  > /usr/local/bin/kubectl \
 && chmod +x /usr/local/bin/kubectl
RUN curl -Ls https://kubernetes-helm.storage.googleapis.com/helm-v${HELM_VERSION}-linux-amd64.tar.gz \
  | tar -xOzf - linux-amd64/helm > /usr/local/bin/helm \
 && chmod +x /usr/local/bin/helm
RUN curl https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GCLOUD_VERSION}-linux-x86_64.tar.gz \
  | tar -C /usr/local -xzf -
RUN apt-get update && apt-get install -y openssh-client git procps && apt-get clean
ENV PATH=/usr/local/bin:/usr/local/google-cloud-sdk/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin
ENTRYPOINT ["/usr/local/bin/kcd"]
