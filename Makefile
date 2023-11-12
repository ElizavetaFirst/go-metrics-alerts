LAST := 14
NUMBERS := $(shell seq 1 ${LAST})
ITERS := $(addprefix iter,${NUMBERS})

.DEFAULT_GOAL := build
.SILENT: ${ITERS}

METRICSTEST_ARGS = -test.v -source-path=.

.PHONY:build
build:
	go build -C ./cmd/agent/ -o agent
	go build -C ./cmd/server/ -o server

.PHONY:clean
clean:
	-rm -f ./cmd/agent/agent
	-rm -f ./cmd/server/server

.PHONY:statictest
statictest:
	go vet -vettool=$$(which statictest) ./...

.PHONY: ${ITERS}
${ITERS}: iter%: build statictest;
	for i in $(shell seq 1 $*) ; do \
		METRICSTEST_ARGS="${METRICSTEST_ARGS} -test.run=TestIteration$$i[AB]?$$" ; \
		if [ $$i -eq 1 ]; then \
			metricstest $$METRICSTEST_ARGS \
						-binary-path=cmd/server/server ; \
		elif [ $$i -eq 2 ]; then \
			metricstest $$METRICSTEST_ARGS  \
						-agent-binary-path=cmd/agent/agent ; \
		elif [ $$i -eq 3 ]; then \
			metricstest $$METRICSTEST_ARGS \
						-binary-path=cmd/server/server \
						-agent-binary-path=cmd/agent/agent ; \
		elif [ $$i -ge 4 ] && [ $$i -le 7 ]; then \
			SERVER_PORT=$$(random unused-port) ; \
			ADDRESS="localhost:$${SERVER_PORT}" ; \
			TEMP_FILE=$$(random tempfile) ; \
			metricstest $$METRICSTEST_ARGS \
						-binary-path=cmd/server/server \
						-agent-binary-path=cmd/agent/agent \
						-server-port=$$SERVER_PORT ; \
		elif [ $$i -ge 8 ] && [ $$i -le 9 ]; then \
			SERVER_PORT=$$(random unused-port) ; \
			ADDRESS="localhost:$${SERVER_PORT}" ; \
			TEMP_FILE=$$(random tempfile) ; \
			metricstest $$METRICSTEST_ARGS \
						-binary-path=cmd/server/server \
						-agent-binary-path=cmd/agent/agent \
						-server-port=$$SERVER_PORT \
						-file-storage-path=$$TEMP_FILE ; \
		elif [ $$i -ge 10 ] && [ $$i -le 13 ]; then \
			SERVER_PORT=$$(random unused-port) ; \
			ADDRESS="localhost:$${SERVER_PORT}" ; \
			TEMP_FILE=$$(random tempfile) ; \
			metricstest $$METRICSTEST_ARGS \
						-binary-path=cmd/server/server \
						-agent-binary-path=cmd/agent/agent \
						-server-port=$$SERVER_PORT \
						-database-dsn='postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable' ; \
		elif [ $$i -eq 14 ]; then \
			SERVER_PORT=$$(random unused-port) ; \
			ADDRESS="localhost:$${SERVER_PORT}" ; \
			TEMP_FILE=$$(random tempfile) ; \
			metricstest $$METRICSTEST_ARGS \
						-binary-path=cmd/server/server \
						-agent-binary-path=cmd/agent/agent \
						-server-port=$$SERVER_PORT \
						-database-dsn='postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable' \
						-key="$$TEMP_FILE" ; \
			go test -v -race ./... ; \
		fi ; \
		if [ $$? -eq 1 ]; then \
			break ; \
		fi ; \
    done