#!/bin/bash

#chmod +x ab_load.sh

ulimit -n 65536

echo "Running A/B test for login API"
ab -n 1000000 -c 500 -T 'application/json' -p ab_loadtest_payload.json http://localhost:8888/api/login

echo "Running A/B test for resource API"
ab -n 1000000 -c 500 -H "X-ACCESS-TOKEN: 44ea030b0acbefdd12847090221c8b7ca93cd6235880.915f8d60e3461e932daf5f9dbde69ae4b09d3b72bc3ce7fb64f9fdbfbe3b9376" http://localhost:8888/api/resource
