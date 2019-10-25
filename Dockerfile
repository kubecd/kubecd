FROM debian:buster
ARG KUBECTL_VERSION=1.16.2
ARG HELM_VERSION=2.9.1
ARG GCLOUD_VERSION=268.0.0
RUN apt-get update && apt-get install -y openssh-client git procps curl && apt-get clean
RUN curl -Ls -o /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl \
 && chmod +x /usr/local/bin/kubectl
RUN curl -Ls https://kubernetes-helm.storage.googleapis.com/helm-v${HELM_VERSION}-linux-amd64.tar.gz \
  | tar -C /usr/local/bin --strip-components=1 -xvzf - linux-amd64/helm
RUN curl https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GCLOUD_VERSION}-linux-x86_64.tar.gz \
  | tar -C /usr/local -xzf -
ENV PATH=/usr/local/bin:/usr/local/google-cloud-sdk/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin
COPY kcd /usr/local/bin/kcd
ENTRYPOINT ["/usr/local/bin/kcd"]
