# Generate Thrift sources
FROM thrift:0.11 AS thrift
COPY idl/ /idl
RUN mkdir /gen_py && thrift -out /gen_py -gen py:dynamic idl/github.com/zedge/kubecd/kubecd.thrift
ARG PYPI_UPLOAD_URL=""

# Run tests and install package
FROM python:3.5-alpine
COPY *.py requirements*.txt /tmp/kubecd/
COPY kubecd/ /tmp/kubecd/kubecd/
COPY --from=thrift /gen_py/kubecd/gen_py/ /tmp/kubecd/kubecd/gen_py/
RUN cd /tmp/kubecd \
 && apk add --no-cache --update gcc python3-dev musl-dev \
 && pip install -r requirements.txt -r requirements-test.txt \
 && apk del gcc python3-dev musl-dev \
 && pytest \
 && pip uninstall -y -r requirements-test.txt \
 && python setup.py sdist \
 && ver=`python setup.py --version` \
 && pip install dist/kubecd-$ver.tar.gz \
 && if [ -n "${PYPI_UPLOAD_URL}" ]; then python setup.py upload -r "${PYPI_UPLOAD_URL}"; fi \
 && cd / \
 && rm -rf /tmp/kubecd

ENTRYPOINT ["/usr/local/bin/kcd"]
