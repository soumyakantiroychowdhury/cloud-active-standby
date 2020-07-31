# cloud-active-standby

## How to build

```
# Build the application
$ go build -tags netgo -o ./app ./server/server.go

# Create Docker image
$ docker build . --tag act_stby_app
Sending build context to Docker daemon 7.381 MB
Step 1/3 : FROM scratch
 --->
Step 2/3 : COPY app /bin/app
 ---> 595a8d52381c
Removing intermediate container 762677cfc878
Step 3/3 : ENTRYPOINT /bin/app
 ---> Running in afbd71c200d9
 ---> ae29545a7bc2
Removing intermediate container afbd71c200d9

# Check if image is successfully created
$ docker images -a
REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
act_stby_app        latest              957a288cd94b        3 days ago          27.7 MB
<none>              <none>              ba582198f623        3 days ago          27.7 MB
```

## How to run

Run two instances of the application.

- instance 1 (Active):

  `$ ./app -c -a <IP address or FQDN of instance 2>`

- instance 2 (Standby):

  `$ ./app -a <IP address or FQDN of instance 1>`

By default the server instance runs on port `8090`.

The command-line parameters are given below:

```
Usage: app [-ch] [-a value] [-p value] [parameters ...]
 -a, --peerAddr=value
                   IP address or FQDN of the peer instance. Mandatory
 -c, --active      If set, the instance will be active instance in the cluster
 -h, --help        Help
 -p, --port=value  Server Port. Peer instance must run on this port. Optional
```

### Run as container

```
# Run the image (Note the argument to the container inside "")
$ sudo docker run -itd --network=host 957a288cd94b
d2fe417e5eb2e3772da97091eadf912f05bca471755b7973546b03206233c404

$ docker exec -it 54f981d30ea5 /bin/app -c -a 192.168.137.1
2020/07/31 11:15:40 Starting Active instance. Peer instance address 192.168.137.1.
...

# From another shell, test if the app is running using "curl"
$ curl localhost:8090/ready
hello world

# See logs from the container
$ docker logs d2fe417e5eb2e3772da97091eadf912f05bca471755b7973546b03206233c404

# Stop the container
$ sudo docker stop d2fe417e5eb2e3772da97091eadf912f05bca471755b7973546b03206233c404
d2fe417e5eb2e3772da97091eadf912f05bca471755b7973546b03206233c404
```

### Run in Kubernetes
```
$ kubectl apply -f app.yaml

$ kubectl get pods -o wide
NAME         READY   STATUS    RESTARTS   AGE   IP           NODE          NOMINATED NODE   READINESS GATES
instance-0   1/1     Running   0          10s   10.36.0.30   slave-node1   <none>           <none>
instance-1   0/1     Running   0          10s   10.36.0.31   slave-node1   <none>           <none>
```
