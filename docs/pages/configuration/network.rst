.. _nuts-network-configuration:

Nuts Network configuration
###########################

.. marker-for-readme

====================================     ============================    =============================================================================================================
Key                                      Default                         Description
====================================     ============================    =============================================================================================================
network.grpcAddr                         :5555                           Local address for gRPC to listen on.
network.publicAddr                                                       Public address (of this node) other nodes can use to connect to it. If set, it is registered on the nodelist.
network.bootstrapNodes                                                   Space-separated list of bootstrap nodes (`<host>:<port>`) which the node initially connect to.
====================================     ============================    =============================================================================================================
