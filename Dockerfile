FROM golang:1.13-alpine3.10 AS builder
RUN apk add --no-cache make git
RUN go env -w GOPROXY=https://goproxy.io,direct
ADD https://github.com/swaggo/swag/archive/v1.6.7.tar.gz ./
RUN tar -xvf ./v1.6.7.tar.gz
RUN cd ./swag-1.6.7/cmd/swag && go install .
WORKDIR /src
COPY . .
RUN go get -v ./...
RUN make

FROM alpine:3.10
COPY --from=builder /src/build/outputs/ /opt/waves/
COPY --from=builder /src/conf /opt/waves/conf
WORKDIR /opt/waves
ENTRYPOINT ["./waves"]
