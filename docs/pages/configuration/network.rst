.. _nuts-network-configuration:

Nuts Network configuration
##########################

SSL/TLS Deployment Layouts
**************************

This section describes which deployment layouts are supported regarding SSL/TLS. In all layouts there should be a valid
X.509 server certificate and private key issued by a publicly trusted Certificate Authority if the node operator wants
other Nuts nodes to be able to connect to the node.

Direct WAN Connection
---------------------

This is the simplest layout where the Nuts node is directly accessible from the internet:

.. raw:: html
    :file: ../../_static/images/network_layouts_directwan.svg

This layout has the following requirements:

* X.509 server certificate and private key must be present on the Nuts node and configured in the Nuts Network engine.

SSL/TLS Offloading
------------------

In this layout incoming TLS traffic is decrypted on a SSL/TLS terminator and then being forwarded to the Nuts node.
This is layout is typically used to provide layer 7 load balancing and/or securing traffic "at the gates":

.. raw:: html
    :file: ../../_static/images/network_layouts_tlsoffloading.svg

This layout has the following requirements:

* X.509 server certificate and private key must be present on the SSL/TLS terminator.
* X.509 client certificates presented to the SSL/TLS terminator must be forwarded as a header to the Nuts node.
* SSL/TLS terminator must use the truststore managed by the Nuts node (which is sourced from the Nuts Registry) as root CA trust bundle.

SSL/TLS Pass-through
--------------------

In this layout incoming TLS traffic is forwarded to the Nuts node without being decrypted:

.. raw:: html
    :file: ../../_static/images/network_layouts_tlspassthrough.svg

Requirements are the same as for the Direct WAN Connection layout.