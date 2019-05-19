#!/bin/bash

(cd ../sharedlib && ./build.sh) && (rm ./elconn.so; cp ../sharedlib/elconn.so ./)
