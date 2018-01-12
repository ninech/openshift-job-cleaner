FROM ubuntu:xenial

LABEL maintainer="engineering@nine.ch"

ARG OC_VERSION
ARG OC_TAG
ENV GOROOT="/usr/lib/go-1.9" \
    GOPATH=/go \
    GOBIN=/go/bin
ENV PATH=$GOROOT/bin:$GOBIN:$PATH

# oc install
RUN apt-get -qq update && \
    apt-get -y install wget ca-certificates --no-install-recommends && \
    wget https://github.com/openshift/origin/releases/download/"${OC_VERSION}"/openshift-origin-client-tools-"${OC_VERSION}"-"${OC_TAG}"-linux-64bit.tar.gz -O /tmp/oc.tar.gz && \
    apt-get -y remove wget && \
    apt-get clean && \
    tar zxvf /tmp/oc.tar.gz -C /tmp/ && \
    mv /tmp/openshift-origin-client-tools-"${OC_VERSION}"-"${OC_TAG}"-linux-64bit/oc /usr/bin/ && \
    chown root:root /usr/bin/oc && \
    chmod +x /usr/bin/oc && \
    rm -rf /tmp/oc.tar.gz /tmp/openshift-origin-client-tools-"${OC_VERSION}"-"${OC_TAG}"-linux-64bit/ && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* && \
    oc version && \
    mkdir -p /go/src/app && \
    mkdir -p /.kube && \
    touch /.kube/config

# application compile and install
COPY ./ /go/src/app
WORKDIR /go/src/app
RUN apt-get -qq update && \
    apt-get -y install software-properties-common && \
    add-apt-repository ppa:gophers/archive && \
    apt-get -qq update && \
    apt-get -y install git golang-1.9-go --no-install-recommends && \
    go get -u github.com/golang/dep/cmd/dep && \
    dep ensure && \
    go install && \
    apt-get -y remove software-properties-common git golang-1.9-go && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* && \
    rm -rf /go/pkg && \
    rm -rf /go/src && \
    rm /go/bin/dep && \
    chown -R 1001:0 /go && \
    touch /.kube/config && \
    chmod -R 0775 /.kube && \
    chown -R 1001:0 /.kube

USER 1001
ENTRYPOINT ["/go/bin/app"]
