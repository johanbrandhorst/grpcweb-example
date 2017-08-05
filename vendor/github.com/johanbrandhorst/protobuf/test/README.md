## GopherJS Protobuf binding automated tests

### Synopsis
We have Qunit tests written in the GopherJS client in `./client/`.
We have a server running the gRPC backend and hosting the GopherJS client in `./server/`.
We have a Chromedriver driven browser integration testing setup
that loads the GopherJS client page and checks that there were 0 failures,
courtesy of `ginkgo`, `gomega` and `agouti`.
Simply running `ginkgo .` starts the server hosting the GopherJS unit test client,
loads the page in Chromedriver using `agouti`, parses the number of failures and
if it's anything other than 0, the entire test fails.

### Requirements
* Chromedriver
* Google Chrome
* ginkgo (`govendor install +program,vendor`)

### Running
```
$ make test
```
