# cloud-active-standby

## How to build

```
# Build the application
$ go build -tags netgo -o ./app ./server/server.go

# Create Docker image
$ docker build . --tag act_stby_eg_1
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
<none>              <none>              595a8d52381c        2 seconds ago       7.32 MB
act_stby_eg_1       latest              ae29545a7bc2        2 seconds ago       7.32 MB

# Run the image
$ sudo docker run -itd --network=host ae29545a7bc2
d2fe417e5eb2e3772da97091eadf912f05bca471755b7973546b03206233c404

# From another shell, test if the app is running using "curl"
$ curl localhost:8090/hello
hello world

# Stop the container
$ sudo docker stop d2fe417e5eb2e3772da97091eadf912f05bca471755b7973546b03206233c404
d2fe417e5eb2e3772da97091eadf912f05bca471755b7973546b03206233c404
```

## How to run

Run two instances of the application. The instances will talk to each other using `/hello` API.

- instance 1:

  `$ ./app -a <IP address or FQDN of instance 2>`

- instance 2:

  `$ ./app -a <IP address or FQDN of instance 1>`

By default the server instance runs on port `8090`.

The command-line parameters are given below:

```
Usage: app [-h] [-a value] [-p value] [parameters ...]
 -a, --peerAddr=value
                   IP address or FQDN of the peer instance
 -h, --help        Help
 -p, --port=value  Server Port. Peer instance must run on this port
```

