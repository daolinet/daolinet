DaoliNet for Efficient and Dynamic Docker Networking
=================

DaoliNet is a Software Defined Networking (SDN) system that is designed to provide efficient and dynamic connection for Docker containers, which is suitable for the lightweight and ephemeral nature of micro-servicing workloads of Docker containers.

Top-Level Features
------------------
* Resource efficiency: Connection of containers does not consume host resource when the containers are not in active communication, but can instantly switch to providing full connection capacity. This is in the same fashion of containers efficient utilization of host CPU resource. You get more out of your server resource.

* Distribution anywhere: Docker servers can be laptops or PCs inside the firewalls of your office or home, servers in your own datacenter, or virtual machines in public clouds such as AWS. Trans-datacenter traffic is always encrypted.

* Network virtualization: You can choose any CIDR IP addresses for your containers, and a container can keep IP address unchanged after moving physical locations.

* Pure software implementation using Open-V-Switch (OVS): Providing network functions as distributed switches, routers, gateways and firewalls. System deployment in a plug-n-play simplicity.

**To learn more about DaoliNet**:  http://www.daolinet.org

Docker in Need of Efficient and Dynamic Networking
=================

Containers can be highly ephemeral in lifecycle when providing efficient micro-services: a large number of dynamic containers in a Docker cloud need quick establishing connections as well as frequent changing connection status. Traditional data plane route learning technologies require Docker servers to frequently learn and update routing information for a large number of dynamic containers, which translate to low efficient use of server resource. To date, networking is a core feature of Docker that is relatively immature and still under heavy development.

DaoliNet for Efficient and Dynamic Docker Networking
==========================================

Architecture
------------
The networking architecture of DaoliNet is based on the OpenFlow standard. It uses an OpenFlow Controller as an intelligent control plane, and Open-V-Switches (OVSes) to implement the datapath. The OpenFlow Controller in DaoliNet is a logically centralized entity but physically a set of HA distributed web-service-like agents. OVSes are ubiquitously available in Linux kernels and hence in all Docker servers.

In a DaoliNet network, all Docker servers are in an Ethernet which is either physically or VPN connected. Each Docker server acts as a virtual router for all of the container workloads that are hosted on that server. However, these virtual routers work in OpenFlow technology and they do not run any routing algorithms. Upon a container initiating a connection, the involved virtual routers will be real-time configured by the OpenFlow Controller to establish a route.

How it Works
------------
When a container initiates a connection, the OVS in the hosting Docker server as the source router will issue a PacketIn request to the OpenFlow Controller. The PacketIn request is just the first packet from the initiating container. The OpenFlow Controller, knowing all Docker servers as OpenFlow routers in the system and seeing PacketIN, can identify another Docker server which hosts the container as the destination workload. This second Docker server is the destination router for the connection. The OpenFlow Controller will respond with a pair of PacketOut flows, one for the source router, and the other for the destination router. These PacketOut flows construct a route consisting of three IP hops in general: (1) src-container-src-server hop, (2) src-server-dst-server hop, and (3) dst-server-dst-container hop. In case of the two containers being hosted in the same Docker server, the PacketOut flow route consists of one hop only: src-container-dst-container hop.

The OpenFlow established connection is in a hot-plug fashion in that, the involved routers have not run any routing algorithm to learn and propagate routing tables. It is the OpenFlow Controller that adds the route information to the respective routers. Also, if a connection becomes idle and upon a time threshold, the hot-plug route flows will be time-out and deleted. Therefore Docker servers as routers in DaoliNet work in a no-connection, no-resource-consumption style. This style of networking resource utilization is very similar to the Linux Container technology utilizing server CPU in that, an idling container consumes little server resource. DaoliNet is an efficient and dynamic networking technology for connecting Docker containers.

Simple Networking for Containers
--------------------------------
In DaoliNet, Docker servers in the system are in a simple state of not-knowing-one-another, completely independent from one another. This architecture not only conserves resource utilization, but more importantly the independent relationship among the Docker servers greatly simplifies the management of resource. Extending the resource pool is as simple as plug-n-play style of adding a server to the pool and notifying the OpenFlow Controller. No complex routing table discovery and update among the routers is needed. There is also no need for Docker servers to pairwise run some packet encapsulation protocol which is not only inefficient in resource utilization but will also nullify network diagnosing and troubleshooting tools such as traceroute.

**More in our website:** http://www.daolinet.org/html/technology.html
