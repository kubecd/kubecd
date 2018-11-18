FROM golang:1.11 AS build
COPY *.mod *.go Makefile /go/src/
RUN cd /go/src; CGO_ENABLED=0 go build -o demo-app main.go

FROM scratch
COPY --from=build /go/src/demo-app /demo-app
ENTRYPOINT ["/demo-app"]
