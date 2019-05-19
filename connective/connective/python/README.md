## Heartbeat example testing

0. Make sure `elconn.so` is the latest binary
1. Run `heartbeat_example_server.py`. The alive status will display as "true"
   briefly.
2. Once the alive status is "false", run `heartbeat_example_client.py`
3. Observe the output of `heartbeat_example_server`. It should report that the
   client is alive for several seconds, then hear for a short period of time,
   then dead for a longer period of time, finally sending a final heartbeat
   before terminating.
