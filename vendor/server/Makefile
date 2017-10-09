# -*-coding:utf-8-unix;-*-
#
cprof:
	go test -cpuprofile=cprof
	go tool pprof --text server.test cprof
#
clean:
	${RM} server.test cprof *~
