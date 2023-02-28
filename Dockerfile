FROM golang:1.19.3-alpine

WORKDIR /opt/ddupdate/
COPY go.mod /opt/ddupdate/
COPY go.sum /opt/ddupdate/
RUN go mod download
COPY cffunc/cffunc.go /opt/ddupdate/cffunc/
COPY main.go /opt/ddupdate/
COPY dnslist /opt/ddupdate/
#COPY README /opt/ddupdate/

RUN go build /opt/ddupdate/main.go
