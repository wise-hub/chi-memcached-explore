exploratory project

```
cd mamcached/
docker build -t memcached-img . 
docker run -d --name memchached-1 -p 11211:11211 --restart always memcached-img


cd ab/
ab -n 10000 -c 1000 -T 'application/json' -p payload.json http://localhost:8080/api/login/ > results.txt

```


