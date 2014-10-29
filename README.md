
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
* Add your ssh key to ssh-agent so that fleetctl can access it
* Install the iris-bench server on each client. `--tunnel` should point at 
  * `fleetctl --tunnel=change-me.compute.amazonaws.com load fleet/iris-bench.service`
  * `fleetctl --tunnel=change-me.compute.amazonaws.com start iris-bench.service`
  * Verify that it worked: `fleetctl --tunnel=change-me.compute.amazonaws.com list-units`
    * You should see that `iris-bench.service` is active on all nodes
* on one machine run: `docker run --net=host coopernurse/iris-bench:v3 /usr/local/bin/iris-bench -m echo -s 10 -c 15 -r 3`

Cool things to try:

* Run `launch.sh` again to add more machines to the cluster
* Wait a few minutes, then run: `fleetctl --tunnel=change-me.compute.amazonaws.com list-units`
  * You should see that the new machines automatically joined the cluster, and ran the 
    `iris-bench.service` because it's configured as a 'global' service that should run on all nodes
* Re-run the benchmark, but increase the `-r` param to the # of machines in the cluster. 
  You should see performance increase as you add machines

## Interesting commands

Create image

    docker build -no-cache -t="coopernurse/iris-bench:v3" .

Upload image to docker registry

    docker push coopernurse/iris-bench:v3

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

