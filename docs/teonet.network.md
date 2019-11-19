# Teonet (fisical) network organisation

## Using webdav file system on Storage Box

Install webdav filesystem:

    apt-get install davfs2

Add to ```vim /etc/davfs2/secrets```:

    https://u205450.your-storagebox.de u205450 3Er3ZtTWUMRtaJrd

Mount manually:

    mkdir /mnt/storage
    mount -t davfs https://u205450.your-storagebox.de /mnt/storage

Use ```vim /etc/fstab``` to mount at startup:

    https://u205450.your-storagebox.de /mnt/storage davfs rw,uid=root,gid=root,file_mode=0660,dir_mode=0770 0 0

## Using Swarm to manage docker containers

***Note: remove ```172.16.0.0/12 via x.x.x.x``` network path during docker and docker swarm setup

### Docker Swarm init

    docker swarm init --advertise-addr 192.168.106.5

### Create worker tocken (and use it to join swarm claster)

    docker swarm join-token worker

Or create master tocken:

    docker swarm join-token master

### Some usefull docker swarm commands

Create container 'ubuntu-sw-04' on the 'gt8' node from image 'ubuntu':

    docker service create -t --name ubuntu-sw-03 --constraint 'node.hostname == gt8' ubuntu

Get all services on all nodes:

    docker node ps $(docker node ls -q)

### Create owerlay network

    docker network create --driver overlay --subnet 10.0.10.0/24 teo-overlay

Create test containers in overlay network:

    docker service create --constraint 'node.hostname == teonet' --network teo-overlay --name ubuntu-overlay-01 -t ubuntu

    docker service create --constraint 'node.hostname == dev-ks-2' --network teo-overlay --name ubuntu-overlay-02 -t ubuntu

    # Remove it after test finished
    docker service rm ubuntu-overlay-01 ubuntu-overlay-02  

### Create local registry to store created docker images and run it in swarm claster

    docker service create --name registry --publish published=5000,target=5000 registry:2

Mark this registry as unsecury in all hosts which used this repository. Edit
the ```daemon.json``` file, whose default location
is ```/etc/docker/daemon.json```. If the daemon.json file does not exist,
create it. Assuming there are no other settings in the file, it should have the
following contents:

    {
        "insecure-registries" : ["192.168.106.5:5000"]
    }

Push image to local repository:

    docker build -t teonet-go .
    docker tag teonet-go 192.168.106.5:5000/teonet-go
    docker push 192.168.106.5:5000/teonet-go

Run in swarm claster:

    docker service create --constraint 'node.hostname == teonet' --network teo-overlay --name teonet-go -t 192.168.106.5:5000/teonet-go teonet -a 5.63.158.100 -r 9010 -n teonet teo-go-01
