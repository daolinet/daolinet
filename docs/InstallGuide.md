daolinet安装文档
==========

此文档描述daolinet在CentOS7系统中的安装过程，其它发行版本只需要将yum安装方式使用apt-get替换即可。包括安装前准备工作和manager及agent节点的安装过程。

安装要求
----------

使用daolinet可以在docker单机环境和集群环境进行部署和安装。进行docker集群管理最少使用一个manager节点和一个agent节点。

安装daolinet前所有节点必须进行如下环境准备：

* docker Version 1.9 or later
* Git
* Python 2.7
* *Golang Version 1.5 or later* (可选)

> ***注意：***如果所有节点使用系统版本一致，可以只在manager节点安装Golang编译环境，其它agent节点直接拷贝编译生成的二进制文件即可)

#### 安装依赖

* Linux安装docker使用如下命令安装，详细安装请参考[On Linux distributions](https://docs.docker.com/engine/installation/linux/)

		curl -fsSL https://get.docker.com/ | sh

* 安装Golang环境可以通过如下步骤进行，详细安装请参考[Go Getting Started](https://golang.org/doc/install)，此步为可选操作，请查看上面注意事项

		wget https://storage.googleapis.com/golang/go1.5.3.linux-amd64.tar.gz
		tar xzvf go1.5.3.linux-amd64.tar.gz -C /usr/local/
		export PATH=$PATH:/usr/local/go/bin

* 安装编译环境 (Linux发行版都带有python环境)

		yum install -y git epel-release
		yum install -y python-devel git python-pip

> ***注意：***所有export命令都可以配置到profile文件中永久生效

构建和安装
-------------

### 安装manager节点安装

manager节点需要安装与daolinet相关组件和其它软件：

* etcd (键值存储/服务发现)
* swarm manager (docker集群manager)
* ryu (openflow框架)
* daolinet (daolinet api server)
* daolictl (daolinet命令行工具)
* daolicontroller (openflow控制器)

以下详细说明manager节点安装步骤

> ***注意：***请先添加iptables规则允许其它节点访问
>
>  		iptables -I INPUT -s <SUBNET>/<PREFIX> -j ACCEPT

#### 1. 安装etcd

	docker pull microbox/etcd
	docker run -ti -d -p 4001:4001 -p 7001:7001 --restart=always --name discovery microbox/etcd -addr <SWARM-IP>:4001 -peer-addr <SWARM-IP>:7001

#### 2. 安装swarm manager

	docker pull swarm
	docker run -ti -d -p 3376:3376 --restart=always --name swarm-manager --link discovery:discovery swarm m --addr <SWARM-MANAGER-IP>:3376 --host tcp://0.0.0.0:3376 etcd://discovery:4001

#### 3. 编译安装daolinet api server

	mkdir $HOME/daolinet
	cd $HOME/daolinet
	export GOPATH=$HOME/daolinet
	go get github.com/tools/godep
	export PATH=$PATH:$GOPATH/bin
	mkdir -p src/github.com/daolicloud
	cd src/github.com/daolicloud
	git clone https://github.com/daolicloud/daolinet.git
	cd daolinet
	godep go build
	mv daolinet ../../../../bin/

	# 启动api server
	daolinet server --swarm tcp://<SWARM-MANAGER-IP>:3376 etcd://<ETCD-IP>:4001

#### 4. 编译安装daolictl命令行工具

接着第三步环境继续进行daolictl的编译和安装

	cd $HOME/daolinet/src/github.com/daolicloud
	git clone https://github.com/daolicloud/daolictl.git
	cd daolictl
	godep go build
	mv daolictl ../../../../bin/

#### 5. 安装openflow控制器

	# 安装openflow框架
	pip install ryu
	# 安装依赖
	yum install -y python-requests python-docker-py
	# 安装openflow控制器程序
	git clone https://github.com/daolicloud/daolicontroller.git
	cd daolicontroller; python ./setup.py install
	# 启动控制器
	daolicontroller

### 安装agent节点安装

agent节点需要安装与daolinet相关组件和其它软件：

* openvswitch
* swarm agent (docker集群agent)
* daolinet (daolinet agent)
* ovsplugin (daolinet ovs插件)

以下详细说明agent节点安装步骤

> ***注意：***请先添加iptables规则允许其它节点访问
>
>  		iptables -I INPUT -s <SUBNET>/<PREFIX> -j ACCEPT

#### 1. 配置docker启动参数

修改docker daemon启动参数，添加swarm管理和etcd支持。例如在CentOS7下修改/usr/lib/systemd/system/docker.service文件中如下ExecStart参数：

	ExecStart=/usr/bin/docker daemon -H fd:// -H tcp://0.0.0.0:2375 --cluster-store=etcd://<ETCD-IP>:4001

然后重启服务：

	systemctl daemon-reload
	systemctl restart docker.service

#### 2. 安装swarm agent

	docker pull swarm
	docker run -ti -d --restart=always --name swarm-agent swarm j --addr <SWARM-AGENT-IP>:2375 etcd://<ETCD-IP>:4001

#### 3. 安装配置openvswitch

OpenVswitch安装请执行以下命令，详细安装请参考[How to Install Open vSwitch on Linux, FreeBSD and NetBSD](https://github.com/openvswitch/ovs/blob/master/INSTALL.md)

	# 编译openvswitch源码
	yum install -y openssl-devel rpm-build
	wget http://openvswitch.org/releases/openvswitch-2.5.0.tar.gz
	mkdir -p ~/rpmbuild/SOURCES
	cp openvswitch-2.5.0.tar.gz ~/rpmbuild/SOURCES/
	tar xzf openvswitch-2.5.0.tar.gz
	rpmbuild -bb --without check openvswitch-2.5.0/rhel/openvswitch.spec

	# 安装
	yum localinstall -y rpmbuild/RPMS/x86_64/openvswitch-2.5.0-1.x86_64.rpm
	/etc/init.d/openvswitch start

	# 配置 (注意：如果通过控制台登录Linux系统，以下操作可能会导致控制台退出或服务器不能连接情况，请将以下<DEVNAME>，<DEVIP>和<DEVMAC>变量指定正确后写入脚本执行)
	systemctl stop NetworkManager
	systemctl disable NetworkManager
	ovs-vsctl add-br daolinet
	ovs-vsctl add-port daolinet <DEVNAME>eno16777728
	ovs-vsctl add-port daolinet eno16777728
	ip addr del <DEVIP> dev <DEVNAME>
	ip addr change <DEVIP> dev <DEVNAME>
	ovs-vsctl set Bridge daolinet other_config:hwaddr="<DEVMAC>"
	# 如果<DEVNAME>作为缺省网关，则需要执行命令
	ip route add default via <GATEWAYIP>

> ***注意：***如果所有agent节点系统版本一致，可以将当前编译好的openvswitch-2.5.0-1.x86_64.rpm安装包直接拷贝并安装到其它agent节点上

#### 4. 安装ovs plugin

	pip install gunicorn flask netaddr
	git clone https://github.com/daolicloud/ovsplugin.git
	cd ovsplugin/
	./start.sh

#### 5. 安装daolinet agent

如果agent节点与manager节点操作系统环境一样，此步可以直接拷贝**安装manager节点**时编译完成的daolinet二进制文件；如果系统环境不一样，此步直接按照**安装manager节点 \* 编译安装daolinet api server**完成编译，再执行以下命令启动agent服务：

	daolinet agent --iface <DEVNAME:DEVIP> etcd://<ETCD-IP>:4001

#### 6. 连接控制器

完成agent节点以上步骤，最后配置ovs连接到daolicontroller控制器：

	ovs-vsctl set-controller daolinet tcp:<CONTROLLER-IP>:6633

## 总结

以上为daolinet安装步骤，下一步，请参考[用户手册文档](UserGuide.md)。

