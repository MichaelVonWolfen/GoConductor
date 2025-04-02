all: clean build run_empty

hello:
	echo "hello"

build:
	go build -o bin/test/GoConductor ./GoConductor.go
	cp -r examples/timer/* bin/test/
	cd bin/test/producer && go build producer.go
	cd bin/test/timer && go build timer.go
run:
	cd ./bin/test/ && ./GoConductor -configPath=/Intel/GfxCPLBatchFiles/

run_empty:
	cd ./bin/test/ && ./GoConductor


clean:
	rm -rf bin/
