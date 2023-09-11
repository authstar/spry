build: godeps
	go build -o ./bin/spry
	cd ./cli && \
		go build -o ../bin/spry-cli && \
		cd ..

godeps:
	mkdir -p ./.test
	go get

golint:
	golangci-lint run --skip-dirs api,space

test: build
	go test -race -covermode=atomic \
		 -coverprofile=./.test/count.out -v \
		 -coverpkg=./... \
		 ./tests
	go tool cover -func=./.test/count.out
	go tool cover -html=./.test/count.out -o ./.test/cov.html

test-pg: build
	go test -race -covermode=atomic \
		 -coverprofile=./.test/pgcount.out -v \
		 -coverpkg=./... \
		 ./postgres/tests
	go tool cover -func=./.test/pgcount.out
	go tool cover -html=./.test/pgcount.out -o ./.test/pgcov.html

test-ci: build
	go test -v ./tests
	go test -race -covermode=atomic \
		 -coverprofile=./.test/coverage.out -v \
		 -coverpkg=./... \
		 ./postgres/tests

test-all: test test-pg

compose:
	cd ./.docker && \
		docker-compose up

decomp:
	cd ./.docker && \
		docker-compose down
	sudo rm -rf ./.docker/volumes/postgres/data