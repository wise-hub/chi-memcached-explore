exploratory project for server-side in-memory session persistence

```

ab -n 100000 -c 1000 -T 'application/json' -p ab_loadtest_payload.json http://localhost:8080/api/login


ab -n 100000 -c 1000 -H "X-ACCESS-TOKEN: 9a65c1fb8bda77dfaf8fa1cc12e3016828fd67d082c9.38d084b546a8878058320f16f8a6a13044ad8ad9ce90781ebfbaef3e8a4946be" http://localhost:8080/api/resource


```


