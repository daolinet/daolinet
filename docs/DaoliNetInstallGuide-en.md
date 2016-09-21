DaoliNet Installation Guide
==========

This document provides the installation steps for DaoliNet over the latest Linux distributions of CentOS7 or Fedora. For Ubuntu or Debian distributios, replace all 'yum' in this guide with 'apt-get'.

1. Environment Preparation
----------

DaoliNet depends on the following software environment:

* Docker version 1.9 or later
* Golang version 1.5 or later
* Git
* Python 2.7

DaoliNet provides network connection between containers which may run on a cluster of docker hosts, and makes use of official docker swarm cluster tools to distribute and manage containers on such a cluster. Therefore the preparation of the above software environment must be repeated on each server node in the cluster.

Each step in the preparation of the above software environment is detailed below.

#### Docker Environment
Execute the following command line, see [Install Docker Engine on Linux](https://docs.docker.com/engine/installation/linux/) for details.

		curl -fsSL https://get.docker.com/ | sh

#### Golang Environment
Execute the following command lines, see [Go Getting Started](https://golang.org/doc/install)

		wget https://storage.googleapis.com/golang/go1.5.3.linux-amd64.tar.gz
		tar xzvf go1.5.3.linux-amd64.tar.gz -C /usr/local/
		export PATH=$PATH:/usr/local/go/bin

#### Git Environment
Execute the following command line

		yum install -y gcc git epel-release

#### Python Environment
Execute the following command line

		yum install -y python-devel python-pip

> Note:
>
> 1. Each export command in this guide can be configured into a Linux profile file to make it permanent
>
> 2. For each server node, add the following iptables rule to allow access by other nodes in the cluster:

>			iptables -I INPUT -s <SUBNET>/<PREFIX> -j ACCEPT


2. DaoliNet Installation
-----------

DaoliNet installation work involves that for manager nodes and that for agent nodes. Repeat manager node installation on each of the server nodes which are selected as manager nodes, and repeat agent node installation on each of the server nodes which are selected as agent nodes. A server node may be used for both a mamager node and an agent node.

### 2.1 Manager Node Installation

The installation of a manager node involves the following six steps:

* Install Etcd
* Install Swarm Manager
* Install Ryu (OpenFlow Framework)
* Install DaoliNet (DaoliNet api service)
* Install Daolictl (DaoliNet command line tool)
* Install Daolicontroller (OpenFlow controller)

Each step above is detailed below.

#### 2.1.1. Install Etcd

	docker pull microbox/etcd
	docker run -ti -d -p 4001:4001 -p 7001:7001 --restart=always --name discovery microbox/etcd -addr <SWARM-MANAGER-IP>:4001 -peer-addr <SWARM-MANAGER-IP>:7001

#### 2.1.2. Install Swarm Manager

	docker pull swarm
	docker run -ti -d -p 3376:3376 --restart=always --name swarm-manager --link discovery:discovery swarm m --addr <SWARM-MANAGER-IP>:3376 --host tcp://0.0.0.0:3376 etcd://discovery:4001

#### 2.1.3. Install Ryu

	# Install openflow framework
	pip install ryu

#### 2.1.4. Install DaoliNet API Service

	mkdir $HOME/daolinet
	cd $HOME/daolinet
	export GOPATH=$HOME/daolinet
	go get github.com/tools/godep
	export PATH=$PATH:$GOPATH/bin
	mkdir -p src/github.com/daolinet
	cd src/github.com/daolinet
	git clone https://github.com/daolinet/daolinet.git
	cd daolinet
	godep go build
	mv daolinet ../../../../bin/

	# Run api server
	daolinet server --swarm tcp://<SWARM-MANAGER-IP>:3376 etcd://<ETCD-IP>:4001

#### 2.1.5. Install Daolictl Command Line Tool

> ***Note:*** Sometimes, you may need to repeat all command lines in Step 2.1.4 before carry on this step

	cd $HOME/daolinet/src/github.com/daolinet
	git clone https://github.com/daolinet/daolictl.git
	cd daolictl
	godep go build
	mv daolictl ../../../../bin/

#### 2.1.6. Install Openflow Controller

	# Install depend packages
	yum install -y python-requests python-docker-py
	# Install openflow controller
	git clone https://github.com/daolinet/daolicontroller.git
	cd daolicontroller; python ./setup.py install
	# Run daolicontroller
	systemctl start daolicontroller.service

### 2.2. Agent Node Installation

The installation of an agent node involves the following six steps:

* Configure Docker Startup Parameters
* Install Swarm Agent
* Configure and Install OpenvSwitch
* Install OpenvSwitch plugin
* Install DaoliNet Agent
* Connect OpenFlow Controller

Each step above is detailed below:

#### 2.2.1. Configure Docker Startup Parameters

1. Modify docker daemon startup parameters, add swarm management and add etcd support, e.g., in CentOS7, modify ExecStart parameter in file /usr/lib/systemd/system/docker.service

		ExecStart=/usr/bin/docker daemon -H fd:// -H tcp://0.0.0.0:2375 --cluster-store=etcd://<ETCD-IP>:4001

2. Restart services:

		systemctl daemon-reload
		systemctl restart docker.service

#### 2.2.2. Install Swarm Agent

	docker pull swarm
	docker run -ti -d --restart=always --name swarm-agent swarm j --addr <SWARM-AGENT-IP>:2375 etcd://<ETCD-IP>:4001

#### 2.2.3. Configure and Install OpenvSwitch

To install OpenvSwitch, execute the following command lines, for detailed installation guide, see [How to Install Open vSwitch on Linux, FreeBSD and NetBSD](https://github.com/openvswitch/ovs/blob/master/INSTALL.md)

	# Compile openvswitch source code
	yum install -y openssl-devel rpm-build
	wget http://openvswitch.org/releases/openvswitch-2.5.0.tar.gz
	mkdir -p ~/rpmbuild/SOURCES
	cp openvswitch-2.5.0.tar.gz ~/rpmbuild/SOURCES/
	tar xzf openvswitch-2.5.0.tar.gz
	rpmbuild -bb --without check openvswitch-2.5.0/rhel/openvswitch.spec

	# Install the created software packages
	yum localinstall -y rpmbuild/RPMS/x86_64/openvswitch-2.5.0-1.x86_64.rpm
	/etc/init.d/openvswitch start


	# Run OpenvSwitch script
	git clone https://github.com/daolinet/daolinet.git
	cd daolinet/
	./ovsconf

#### 2.2.4. Install OpenvSwitch Plugin

	pip install gunicorn flask netaddr
	git clone https://github.com/daolinet/ovsplugin.git
	cd ovsplugin/
	./start.sh

#### 2.2.5. Install DaoliNet Agent Service

	# Install daolinet
	mkdir $HOME/daolinet
	cd $HOME/daolinet
	export GOPATH=$HOME/daolinet
	go get github.com/tools/godep
	export PATH=$PATH:$GOPATH/bin
	mkdir -p src/github.com/daolinet
	cd src/github.com/daolinet
	git clone https://github.com/daolinet/daolinet.git
	cd daolinet
	godep go build
	mv daolinet ../../../../bin/

	# Run agent service
	daolinet agent --iface <DEVNAME:DEVIP> etcd://<ETCD-IP>:4001

#### 2.2.6. Connect OpenFlow Controller

In agent node, complete the above steps, and finally configure ovs connect to daolicontroller controller:

	ovs-vsctl set-controller daolinet tcp:<CONTROLLER-IP>:6633

***Node:*** In order to increase the system availability, you may start more than one daolicontroller controllers in the cluster, and in the configuration of ovs specify the addresses for these controllers:

	ovs-vsctl set-controller daolinet tcp:<CONTROLLER-IP1>:6633,tcp:<CONTROLLER-IP2>:6633

#### We are done! Try DaoliNet now! (see [DaoliNet User Guide](DaoliNetUserGuide-en.md))
