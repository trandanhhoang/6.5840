# RUN Sequential
cd src/main
go build -buildmode=plugin ../mrapps/wc.go
rm mr-out*
go run mrsequential.go wc.so pg*.txt
more mr-out-0
A 509
ABOUT 2
ACT 8
...


# RUN Coordinate + Worker
cd src/main
go build -buildmode=plugin ../mrapps/wc.go
go run mrcoordinator.go pg-*.txt
rm mr-out*
go run mrcoordinator.go pg-*.txt
go run mrworker.go wc.so

# when you run finish, check by command
cat mr-out-* | sort | more
A 509
ABOUT 2
ACT 8
...

# Test
- main/test-mr.sh
- how to run
    $ cd ~/6.5840/src/main
    $ bash test-mr.sh
    *** Starting wc test.
- run with ret := false
    $ bash test-mr.sh
    *** Starting wc test.
    sort: No such file or directory
    cmp: EOF on mr-wc-all
    --- wc output is not the same as mr-correct-wc.txt
    --- wc test: FAIL
- When you've finished, the test script output should look like this:
```
$ bash test-mr.sh
    *** Starting wc test.
    --- wc test: PASS
    *** Starting indexer test.
    --- indexer test: PASS
    *** Starting map parallelism test.
    --- map parallelism test: PASS
    *** Starting reduce parallelism test.
    --- reduce parallelism test: PASS
    *** Starting job count test.
    --- job count test: PASS
    *** Starting early exit test.
    --- early exit test: PASS
    *** Starting crash test.
    --- crash test: PASS
    *** PASSED ALL TESTS
```

# Note
- You may see some errors from the Go RPC package that look like
    - 2019/12/16 13:27:09 rpc.Register: method "Done" has 1 input parameters; needs exactly three -> IGNORE 
