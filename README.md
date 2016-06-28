DaoliNet for Dynamic, Efficient, Scalable and Virtualized Docker Networking
=================

DaoliNet is a Software Defined Networking (SDN) system to provide dynamic, efficient, scalable and virtualized connection for Docker containers.

Top-Level Features
------------------
* Dynamic connection: Containers are hot-plug connected according to their PaaS business required relationship rather than their IP address definition, and the control of a connection is dynamic following the dynamic change of the PaaS business requirement.

* Efficient connection: Connection of containers consumes little host resource when the containers are not in active communication, but can instantly switch to providing full connection capacity. This is in the same fashion of containers efficient utilization of host CPU resource. You get more out of your server resource.

* Scalability: Container cloud can grow to very large scale in pyramid structure of L3 routers connecting L2 switches to keep the connection quick. This pyramid structured networking is in contrast to container-MAC-in-UDP encapsulation flattened L2 switches which have to hold a large number of container MAC addresses to slow down the connection.

* Network virtualization: A container keeps IP address unchanged when moving physical locations. A multi-tenancy cloud allows tenants to freely choose IP addresses for their containers.

* Pure software implementation using Open-V-Switch (OVS): Providing network functions as distributed switches, routers, gateways and firewalls. System deployment in a plug-n-play simplicity.

**To learn more about DaoliNet**:  http://www.daolicloud.com

Docker in Need of Dynamic, Scalable, Efficient and Virtualized Networking
=================
Containers can be highly ephemeral in lifecycle when providing efficient micro-services: a large number of containers in a Docker cloud need quick establishing connections, dynamic changing connection status, or real-time providing scale-out PaaS capabilities. Known container networking technologies include: (1) Container-IP route learning network which requires Docker servers to frequently learn and update routing information for a large number of containers; (2) Container-MAC overlay network by encapsulating container-MAC-in-UDP which requires virtual switches to cache large numbers of container MAC addresses. A major limitation of (1) is that an IP route learning network is physical and can hardly be virtualized. A major limitation of (2) is that a flatten container-MAC-in-UDP network limits the scalability of the container cloud, and also such a MAC network lacks a dynamic structure which a PaaS platform needs. To date, networking is a core feature of Docker that is relatively immature and still under heavy development.

DaoliNet
========

Architecture
------------
The networking architecture of DaoliNet is based on the OpenFlow standard. It uses an OpenFlow Controller as an intelligent control plane, and Open-V-Switches (OVSes) to implement the datapath. The OpenFlow Controller in DaoliNet is a logically centralized entity but physically a set of HA distributed web-service-like agents. OVSes are ubiquitously available in Linux kernels and hence in all Docker servers.

In a DaoliNet network, all Docker servers are either in an Ethernet which is physically connected, or in a large L2 which is distributed over a scalable IP fabric (e.g., server MAC-in-UDP). Each Docker server acts as a virtual router for all of the container workloads that are hosted on that server. However, these virtual routers work in OpenFlow technology and they do not run any routing algorithms. Upon a container initiating a connection, the involved virtual routers will be real-time configured by the OpenFlow Controller to establish a route.

![DaoliNet Architecture](http://www.daolicloud.com/topology/topologynew.png)

How it Works
------------
When a container initiates a connection, the OVS in the hosting Docker server as the source router will issue a PacketIn request to the OpenFlow Controller. The PacketIn request is just the first packet from the initiating container. The OpenFlow Controller, knowing all Docker servers as OpenFlow routers in the system and seeing PacketIn, can identify another Docker server which hosts the container as the destination workload. This second Docker server is the destination router for the connection. The OpenFlow Controller will respond with a pair of PacketOut flows, one for the source server, and the other for the destination server. These PacketOut flows establish a hot-plug route between the two containers, see figure below.

![Hot-Plug Route Establishment](http://www.daolicloud.com/topology/topology2.png)

Scalability
-----------
A hot-plug route consists of three IP hops in general: (1) src-container-src-server hop, where the containers' logical IPs are changed into their physical IPs , (2) src-server-dst-server hop, and (3) dst-server-dst-container hop, where the containers' physical IPs are changed back to their logical IPs. The scalability of the container cloud network is achieved by connecting the servers over an IP fabric, e.g., VXLAN encapsulation of the servers MAC-in-UDP, see figure below.

![IP Hops of Hot-Plug Route](http://www.daolicloud.com/topology/topology4.png)

In case of the two containers being hosted in the same Docker server, the PacketOut flow route consists of one hop only: src-container-dst-container hop.

Network Virtualization
----------------------
The logical IP of a container is visible to applications or other micro-service containers, and is fixed to the identity of the container, and will not change in the lifecycle of the container. When a container moves physical location, its logical IP will not change, its physical IP changes. The physical IP of a container is invisible to applications or other micro-service containers. When a connection becomes idle and upon a time threshold, the hot-plug route flows will be time-out and deleted to release servers resource. Since hot-plug route establishment is fast, deleted inactive connection can be re-hot-plug upon reconnection. Therefore Docker servers as routers in DaoliNet work in a no-connection, no-resource-consumption style. This style of networking resource utilization matches exactly the fashion of container utilizing server CPU in that, an idle container consumes little server resource. DaoliNet is an efficient and dynamic networking technology for connecting Docker containers.

Simple Networking for Containers
--------------------------------
In DaoliNet, Docker servers in the system are in a simple state of not-knowing-one-another, completely independent from one another. This architecture not only conserves resource utilization, but more importantly the independent relationship among the Docker servers greatly simplifies the management of resource. Extending the resource pool is as simple as plug-n-play style of adding a server to the pool and notifying the OpenFlow Controller. No complex routing table discovery and update among the routers is needed. There is also no need for Docker servers to pairwise run some packet encapsulation protocol which is not only inefficient in resource utilization but will also nullify network diagnosing and troubleshooting tools such as traceroute.

**More in our website:** http://www.daolicloud.com/html/technology.html
