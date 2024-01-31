### Load testing with Grafana K6
[Install K6](https://grafana.com/docs/k6/latest/get-started/installation/)

### How to use?

Linux/Mac/Windows:
```
k6 run loadtest-script.js
```

Docker:
```
docker run --rm -i grafana/k6 run - <loadtest-script.js
```

>When using the `k6` docker image, you can't just give the script name since the script file will not be available to the container as it runs. Instead you must tell k6 to read `stdin` by passing the file name as `-`. Then you pipe the actual file into the container with `<` or equivalent. This will cause the file to be redirected into the container and be read by k6.

### Improve performance of GET API
Varnish (HTTP Cache) is used. It can perfectly launch the GET API's performance to 100K RPS. For example:

```yaml
kind: ConfigMap
apiVersion: v1
metadata:
  name: varnish-config
data:
  default.vcl: |
    vcl 4.1;

    import dynamic;
    import std;

    backend default {
      .host = "127.0.0.1";
      .port = "80";
    }

    sub vcl_recv {
      set req.url = std.querysort(req.url);

      # Only cache GET and HEAD requests
      if (req.method != "GET" && req.method != "HEAD") {
        return (pass);
      }

      if (req.http.Authorization) {
        return (pass);
      }

      return (hash);
    }

    sub vcl_hash {
      if (req.http.Cookie ~ "XSRF-TOKEN=") {
        set req.http.X-TMP = regsub(req.http.Cookie, "^.*?XSRF-TOKEN=([^;]+);*.*$", "\1");
        hash_data(req.http.X-TMP);
        unset req.http.X-TMP;
      }

      hash_data(req.url);
      if (req.http.host) {
        hash_data(req.http.host);
      } else {
        hash_data(server.ip);
      }
    }

    sub vcl_backend_response {
      if (beresp.http.Cache-Control) {
        set beresp.ttl = 1h;
        set beresp.http.X-Cacheable = "YES:Forced";
        set beresp.http.Cache-Control = "max-age=120";
      }
    }

    sub vcl_deliver {
      # Debug header
      if(req.http.X-Cacheable) {
        set resp.http.X-Cacheable = req.http.X-Cacheable;    
      } elseif(obj.uncacheable) {
        if(!resp.http.X-Cacheable) {
          set resp.http.X-Cacheable = "NO:UNCACHEABLE";        
        }
      } elseif(!resp.http.X-Cacheable) {
        set resp.http.X-Cacheable = "YES";
      }
    }

---
apiVersion: v1
kind: Service
metadata:
  name: app-with-varnish
spec:
  selector:
    app: app-with-varnish
  ports:
  - protocol: TCP
    port: 8080

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-with-varnish
  labels:
    app: app-with-varnish
spec:
  selector:
    matchLabels:
      app: app-with-varnish
  template:
    metadata:
      labels:
        app: app-with-varnish
    spec:
      containers:
      - name: app-with-varnish
        image: app
        resources:
          requests:
            cpu: 2
            memory: 4Gi
          limits:
            cpu: 2
            memory: 4Gi
        ports:
        - containerPort: 80
          name: app
      - name: varnish
        image: varnish
        ports:
        - containerPort: 8080
          name: varnish
        env:
        - name: VARNISH_HTTP_PORT
          value: "8080"
        - name: VARNISH_SIZE
          value: 2G
        resources:
          requests:
            cpu: 2
            memory: 4Gi
          limits:
            cpu: 2
            memory: 4Gi
        volumeMounts:
        - name: varnish-config
          mountPath: /etc/varnish/default.vcl
          subPath: default.vcl
        securityContext:
          runAsUser: 0
      volumes:
      - name: varnish-config
        configMap:
          name: varnish-config
          items:
            - key: default.vcl
              path: default.vcl
```