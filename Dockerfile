# Generate Thrift sources
FROM thrift:0.11 AS thrift
COPY idl/ /idl
RUN mkdir /gen_py && thrift -out /gen_py -gen py:dynamic idl/github.com/zedge/kubecd/kubecd.thrift

# Run tests and install package
FROM python:3.5 AS build
#RUN apk add --no-cache --update gcc python3-dev musl-dev
COPY *.py requirements*.txt /build/
COPY kubecd/ /build/kubecd/
WORKDIR /build
RUN pip install -r requirements.txt -r requirements-test.txt
COPY --from=thrift /gen_py/kubecd/gen_py/ /build/kubecd/gen_py/
RUN pytest
RUN python setup.py clean sdist

FROM python:3.5
COPY --from=build /build/dist/kubecd-*.tar.gz /tmp/
RUN pip install /tmp/kubecd-*.tar.gz
ENTRYPOINT ["/usr/local/bin/kcd"]
