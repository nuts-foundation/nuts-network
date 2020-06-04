.. _nuts-network-configuration:

Nuts Network configuration
###########################

.. marker-for-readme

====================================     ============================    =============================================================================================================
Key                                      Default                         Description
====================================     ============================    =============================================================================================================
network.grpcAddr                         :5555                           Local address for gRPC to listen on.
network.bootstrapNodes                                                   Space-separated list of bootstrap nodes (`<host>:<port>`) which the node initially connect to.
network.publicAddr                                                       Public address (of this node) other nodes can use to connect to it. If set, it is registered on the nodelist.
network.nodeID                                                           Instance ID of this node under which the public address is registered on the nodelist. If not set, the Nuts node's identity will be used.
network.certFile                                                         PEM file containing the certificate this node will identify itself with to other nodes.
network.certKeyFile                                                      PEM file containing the key belonging to this node's certificate.
====================================     ============================    =============================================================================================================
