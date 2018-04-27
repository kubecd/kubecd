# Generate Thrift sources
FROM thrift:0.11 AS thrift
COPY idl/ /idl
RUN mkdir /gen_py && thrift -out /gen_py -gen py:dynamic idl/github.com/zedge/kubecd/kubecd.thrift

# Run tests and install package
FROM python:3.5 AS build
#RUN apk add --no-cache --update gcc python3-dev musl-dev
COPY *.py requirements*.txt README.md /build/
COPY kubecd/ /build/kubecd/
WORKDIR /build
RUN pip install -r requirements.txt -r requirements-test.txt
COPY --from=thrift /gen_py/kubecd/gen_py/ /build/kubecd/gen_py/
RUN pytest
RUN python setup.py clean sdist

FROM python:3.5
ARG KUBECTL_VERSION=1.8.6
ARG HELM_VERSION=2.8.2
ARG GCLOUD_VERSION=199.0.0
COPY --from=build /build/dist/kubecd-*.tar.gz /tmp/
RUN pip install /tmp/kubecd-*.tar.gz
RUN curl https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl > /usr/local/bin/kubectl \
 && chmod +x /usr/local/bin/kubectl
RUN curl https://kubernetes-helm.storage.googleapis.com/helm-v${HELM_VERSION}-linux-amd64.tar.gz|tar -xOzf - linux-amd64/helm > /usr/local/bin/helm \
 && chmod +x /usr/local/bin/helm
RUN curl https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GCLOUD_VERSION}-linux-x86_64.tar.gz | tar -C /usr/local -xzf -
ENV PATH=/usr/local/bin:/usr/local/google-cloud-sdk/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ENTRYPOINT ["/usr/local/bin/kcd"]
