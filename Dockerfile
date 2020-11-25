FROM debian:buster-slim
ARG KUBECTL_VERSION=1.19.4
ARG HELM_VERSION=3.4.1
RUN apt-get update && apt-get install -y openssh-client git procps curl && apt-get clean
RUN curl -Ls -o /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl \
 && chmod +x /usr/local/bin/kubectl
RUN curl -Ls https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz \
  | tar -C /usr/local/bin --strip-components=1 -xvzf - linux-amd64/helm
ENV PATH=/usr/local/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin
COPY kcd /usr/local/bin/kcd
ENTRYPOINT ["/usr/local/bin/kcd"]
