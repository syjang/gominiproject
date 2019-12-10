FROM ubuntu:18.04

RUN apt-get update && apt -y install alien wget
RUN wget https://download.oracle.com/otn_software/linux/instantclient/193000/oracle-instantclient19.3-basic-19.3.0.0.0-1.x86_64.rpm
RUN alien -i oracle-instantclient19.3-basic-19.3.0.0.0-1.x86_64.rpm

ENV GOLANG_VERSION 1.11.4
RUN wget -nv -O - https://storage.googleapis.com/golang/go${GOLANG_VERSION}.linux-amd64.tar.gz \
    | tar -C /usr/local -xz
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

COPY . /test
RUN cd test
