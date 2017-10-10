# -*-coding:utf-8-unix;-*-
all: dep tunnel-server tunnel-client
#
run-client: tunnel-client
	./tunnel-client -logtostderr
#
run-server: tunnel-server
	./tunnel-server -logtostderr
#
tunnel-server:
	go build -o tunnel-server -ldflags '-w -s' server/server.go
#
tunnel-client:
	go build -o tunnel-client -ldflags '-w -s' client/client.go
#
dep:
	go get -v -u github.com/golang/glog
	./gen
#
clean:
	${RM} *~ tunnel-server tunnel-client

