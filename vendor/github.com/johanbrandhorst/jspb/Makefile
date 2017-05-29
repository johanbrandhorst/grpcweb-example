all: checkout build

checkout:
	npm install google-protobuf webpack

build:
	./node_modules/.bin/webpack

clean:
	rm grpc.inc.js jspb.inc.js
	rm -rf node_modules
