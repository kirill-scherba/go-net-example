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

Swarm init :

    docker swarm init --advertise-addr 192.168.106.5

Create worker tocken (and use it to join swarm claster):

    docker swarm join-token worker

Or create master tocken:

    docker swarm join-token master

Create container 'ubuntu-sw-04' on the 'gt8' node from image 'ubuntu':

    docker service create -t --name ubuntu-sw-03 --constraint 'node.hostname == gt8' ubuntu

Get all services on all nodes:

    docker node ps $(docker node ls -q)
