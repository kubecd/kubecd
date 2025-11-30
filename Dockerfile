# Build stage
FROM --platform=$BUILDPLATFORM golang:1.23-bookworm AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags "-w -s" -o /out/kcd ./cmd/kcd

# Final stage
FROM debian:bookworm-slim
ARG TARGETARCH
ARG KUBECTL_VERSION=1.19.4
ARG HELM_VERSION=3.4.1
ARG KUSTOMIZE_VERSION=5.4.3
RUN apt-get update && apt-get install -y openssh-client git procps curl && apt-get clean
RUN curl -Ls -o /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/${TARGETARCH}/kubectl \
 && chmod +x /usr/local/bin/kubectl
RUN curl -Ls https://get.helm.sh/helm-v${HELM_VERSION}-linux-${TARGETARCH}.tar.gz \
  | tar -C /usr/local/bin --strip-components=1 -xvzf - linux-${TARGETARCH}/helm
RUN curl -Ls https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_linux_${TARGETARCH}.tar.gz \
  | tar -C /usr/local/bin -xvzf -
ENV PATH=/usr/local/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin
COPY --from=builder /out/kcd /usr/local/bin/kcd
ENTRYPOINT ["/usr/local/bin/kcd"]
