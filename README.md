
## Running the demo

* Create a new RSA key using `ssh-keygen`
* Create a new etcd discovery token at: https://discovery.etcd.io/new
* Copy yml file: `cp launch-user-data.tmpl.yml launch-user-data.yml`
* Edit `launch-user-data.yml`
  * Paste your RSA private key in (get the indentation right)
  * Paste your etcd discovery token (see: "PUT-YOUR-TOKEN-HERE")
* Edit launch.sh 
  * You'll need VPC subnet id and a security group id
* Install the AWS python command line tools (pip install awscli)
* Run launch.sh
* ssh into the machines and run:  `docker run -d --net=host coopernurse/iris-bench:v1`
* on one machine run: `docker run --net=host coopernurse/iris-bench:v1 /usr/local/bin/iris-bench -m echo -s 10 -c 15`

## Interesting commands

Create image

    docker build -no-cache -t="coopernurse/iris-bench:v1" .

Upload image to docker registry

    docker push coopernurse/iris-bench:v1

On CoreOS host, run server:

    # --net=host is needed to bridge the host Iris process to 
    # the container.  it's not the best thing security-wise..
    docker run -d --net=host coopernurse/iris-bench:v1

On CoreOS host, run client benchmarker:

    # echo 
    docker run --net=host coopernurse/iris-bench:v1 /usr/local/bin/iris-bench -m echo -s 10 -c 15

    # add
    docker run --net=host coopernurse/iris-bench:v1 /usr/local/bin/iris-bench -m add -s 10 -c 15

    docker run -e GOMAXPROCS=2 --net=host coopernurse/iris-bench:v1 /usr/local/bin/iris-bench -m add -s 10 -c 15

