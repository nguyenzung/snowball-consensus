binary=snowball-consensus

.PHONY: demo all clean kill

log:
	mkdir log

demo: $(binary) log
	for i in `seq 0 199`; do (./$< $$i > log/log$$i.txt &); done;

all: $(binary)
	@echo $<
	./$< &

clean:
	go clean && rm -rf log/

$(binary): main.go servicenode/node.go
	go build

kill:
	ps aux | grep snowball | awk '{print $$2}' | xargs kill -9


