DaoliNet User Guide
=========

Suppose you have installed DaoliNet by following the instructions in [DaoliNet Installation Guide](DaoliNetInstallGuide-en.md). Also make sure that all services listed there have been started-up properly.

DaoliNet provides a set of CLI (Command Line Interface) commands to:

* Group connect/disconnect CIDR (Classless Inter-Domain Routing) subnets which have been created by Docker network with DaoliNet driver plugin. Containers in a group of CIDR subnets are connected whether or not they are on the same Docker host.
* Connect/disconnect any two containers. These containers can be on any Docker hosts.
* Set firewall policy controls for a container.

#### 1. Docker Network with DaoliNet Driver

DaoliNet makes use of docker network to create and manage container subnets by plugging DaoliNet driver.

1.1. Create a Network

Use 'docker network' CLI to create container networks. The following two docker CLIs exemplify creating two docker networks in CIDR subnets. The parameter --driver=daolinet specifies that these subnets will have DaoliNet networking properties.

    docker -H :3380 network create --subnet=10.1.0.0/24 --gateway=10.1.0.1 --driver=daolinet dnet1
    docker -H :3380 network create --subnet=192.168.0.0/24 --gateway=192.168.0.1 --driver=daolinet dnet2

The above CLI commands are executed on a Docker Swarm Manager node which is also a DaoliNet API Service Manager (see Section 2.1.4 of DaoliNet Installation Guide, "DaoliNet API Service", we always install DaoliNet API Service Manager on a Docker Swarm Manager node). If a CLI command is executed on a non-Swarm Manage node, then you must specify the IP address of the Docker Swarm Manager node in -H parameter. For example:

	docker -H <SWARN-MANAGER-IP>:3380 network create --subnet=10.1.0.0/24 --gateway=10.1.0.1 --driver=daolinet dnet1

1.2. Launch a Container

Use 'docker run' CLI to launch a container in an CIDR subnet; the subnet has been created using the 'docker network' CLI （see Section 1.1.); you should spcidfy the name of the subnet using --net parameter:

    # Launch containers in 10.1.0.0/24 subnet
    docker -H :3380 run -ti -d --net=dnet1 --name test1 centos
    # Suppose that container test1 has IP address 10.1.0.2 (You may view the actual IP address using 'docker inspect test1')
    docker -H :3380 run -ti -d --net=dnet1 --name test2 centos
    # Suppose that container test2 has IP address 10.1.0.3 (You may view the actual IP using 'docker inspect test2')

    # Launch containers in 192.168.0.0/24 subnet
    docker -H :3380 run -ti -d --net=dnet2 --name test3 centos
    # Suppose test3 has IP 192.168.0.2
    docker -H :3380 run -ti -d --net=dnet2 --name test4 centos
    # Suppose test4 has IP 192.168.0.3

1.3. Test Container Network

The default networking rule of DaoliNet: ***The workloads in the same subnet are connected, and those in different subnets are not connected***

    # Enter container test1
    docker -H :3380 attach test1

    # In test1, ping test2, connected
    >> ping 10.1.0.3

    # In test1, ping test3 or test4, not connected
    >> ping 192.168.0.2
    >> ping 192.168.0.3

#### 2. DaoliNet Network Control and Management

As we have seen in Section 1, containers in different subnets are not connected. DaoliNet can connect different subnets by lacing them in a network group.

2.1. Create a Network Group

    # Create a network group
    daolictl group create G1
    # G1 is the name of a newly created network group

    # Add a subnet into a network group
    daolictl member add --group G1 dnet1
    daolictl member add --group G1 dnet2
    # You can view members in a network group
    daolictl group show G1

    # Now in container test1, ping container test3 and test4, connected
    >> ping 192.168.0.2
    >> ping 192.168.0.3

    # Now remove a subnet from the network group
    daolictl member rm --group G1 dnet2
    # Now in container test1 ping container test3 or test4, not connected
    >> ping 192.168.0.2
    >> ping 192.168.0.3

2.2. Fine Granular Control

DaoliNet can control connection between any two containers

    # Disconnect two containers
    daolictl disconnect test1:test2
    # Now in container test1, ping container test2 (suppose its IP is 10.1.0.3)，not connected
    >> ping 10.1.0.3

    # Resume connection
    daolictl connect test1:test2
    # In test1 ping test2, connected
    >> ping 10.1.0.3

2.3. Set Firewall Policy

If a container has launched a service, you can map the service port number to make the service usable externally

> **Note, please login an agent node to add the service image**
>
> For example, login agent-node and download ssh service and apache service images:
>
>       ssh agent-node
>       docker pull daolicloud/centos6.6-ssh
>       docker pull daolicloud/centos6.6-apache

    # First, launch containers to test firewall rules
    docker -H :3380 run -ti -d --net=dnet1 --name testssh daolicloud/centos6.6-ssh
    docker -H :3380 run -ti -d --net=dnet1 --name testweb daolicloud/centos6.6-apache

    # Create an ssh firewall rule named fw-ssh for container testssh, map the ssh port 22 in the container to the server port 20022
    daolictl firewall create --container testssh --rule 20022:22 fw-ssh
    # Now access the ssh service of the container, <GATEWAY IP> is the ip address of the hosting server
    daolictl firewall show testssh
    ssh <GATEWAY IP> -p 20022

    # Create a apache firewall rule named fw-web to container testweb, map the apache port 80 in the container to the server port 20080
    daolictl firewall create --container testweb --rule 20080:80 fw-web
    # Now access the apache service of the container, <GATEWAY IP> is the ip address of the hosting server
    daolictl firewall show testweb
    curl -L http://<GATEWAY IP>:20080

    # Delete a named firewall rule
    daolictl firewall delete fw-ssh fw-web


