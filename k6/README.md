[Install K6](https://grafana.com/docs/k6/latest/get-started/installation/)

How to use?

Linux/Mac/Windows:
```
k6 run loadtest-script.js
```

Docker:
```
docker run --rm -i grafana/k6 run - <loadtest-script.js
```

>When using the `k6` docker image, you can't just give the script name since the script file will not be available to the container as it runs. Instead you must tell k6 to read `stdin` by passing the file name as `-`. Then you pipe the actual file into the container with `<` or equivalent. This will cause the file to be redirected into the container and be read by k6.