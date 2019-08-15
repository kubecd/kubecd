# Generate Thrift sources
FROM thrift:0.11 AS thrift
COPY idl/ /idl
RUN mkdir /gen_py && thrift -out /gen_py -gen py:dynamic idl/github.com/zedge/kubecd/kubecd.thrift

# Run tests and install package
FROM python:3.7 AS build
COPY *.py requirements*.txt README.md /build/
COPY kubecd/ /build/kubecd/
WORKDIR /build
RUN pip install -r requirements.txt -r requirements-dev.txt
COPY --from=thrift /gen_py/kubecd/gen_py/ /build/kubecd/gen_py/
RUN pytest
RUN python setup.py clean sdist

# Grab package from build step and install in a clean Python image,
# along with kubectl, helm, gcloud, ssh and git
FROM python:3.7
ARG KUBECTL_VERSION=1.13.7
ARG HELM_VERSION=2.9.1
ARG GCLOUD_VERSION=258.0.0
ARG HUB_VERSION=2.12.3
COPY --from=build /build/dist/kubecd-*.tar.gz /tmp/
RUN pip install /tmp/kubecd-*.tar.gz
RUN curl -Ls https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl \
  > /usr/local/bin/kubectl \
 && chmod +x /usr/local/bin/kubectl
RUN curl -Ls https://kubernetes-helm.storage.googleapis.com/helm-v${HELM_VERSION}-linux-amd64.tar.gz \
  | tar -xOzf - linux-amd64/helm > /usr/local/bin/helm \
 && chmod +x /usr/local/bin/helm
RUN curl -Ls https://github.com/github/hub/releases/download/v${HUB_VERSION}/hub-linux-amd64-${HUB_VERSION}.tgz \
  | tar --wildcards -xOzf - '*/hub' > /usr/local/bin/hub \
 && chmod +x /usr/local/bin/hub
RUN curl https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GCLOUD_VERSION}-linux-x86_64.tar.gz \
  | tar -C /usr/local -xzf -
RUN apt-get update && apt-get install -y openssh-client git procps && apt-get clean
ENV PATH=/usr/local/bin:/usr/local/google-cloud-sdk/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin
ENTRYPOINT ["/usr/local/bin/kcd"]
