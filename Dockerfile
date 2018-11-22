FROM golang:1.11

ARG pkgpath=/go/src/github.com/Hrubon/mutexdash
ARG runpath=/opt/showmax/mutexdash
ENV bin=$runpath/mutexdash

WORKDIR $runpath

COPY . $pkgpath

RUN dep ensure
RUN go build $pkgpath
RUN cp -r $pkgpath/templates $runpath

ENTRYPOINT [ "/bin/bash", "-c", "$bin -e $ETCD_EPS -t $ETCD_TO -n $ETCD_NS -l $HTTP_LISTEN_ON" ]
