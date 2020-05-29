.. _nuts-network-technical:

Nuts Network engine
###################

This section describes the workings of the Nuts Network engine.

Status
******

This engine intents to provide a decentralized, persistent ledger for Nuts nodes to share information. While it is currently
Work-in-Progress, we intent to migrate the Registry Github-sync mechanism and Corda consents to it. These are known as
*applications* on the network. The Registry application will be realized first and will serve as Proof-of-Concept (POC) for the
network as a viable distributed ledger for Nuts. If successful we can decide whether to migrate Corda consents to the network too.

The Registry application will first operate alongside the Registry's Github-sync as to assess learn how the networks behaves
in real-world scenarios and to assess the POC. If successful we can decide whether switch from Github-sync to the network application.

Design
******

The business end of the engine consists of 5 layers:

.. raw:: html
    :file: ../../_static/images/network_engine_layers.svg

Applications
^^^^^^^^^^^^

When services want to exchange information with other Nuts nodes they can build their application on top of the *DocumentLog* layer.
Internal (to the network engine) applications can feed back information into one of the lower layers: the *NodeList* application
is used to discover new potential peers for the P2P layer.

Interfaces
^^^^^^^^^^

The Network engine has 2 interfaces:

1. REST interface for diagnostics (like all engines) and for applications to query/add documents.
2. Message queue for applications to subscribe on document events.

Applications subscribe to specific document types (based on the **type** property). Documents that are needed for the
basic functionality of the network use the **nuts.** prefix:

* **nuts.node-info**: documents produced/consumed by the NodeList

DocumentLog
^^^^^^^^^^^

The DocumentLog provides a persistent document log which contains all documents on the network (known by the local node),
fed by the network itself (remote peers) and the local node. Through this layer applications can add and query documents.
It makes sure the documents are in a consistent order and tries to build a complete view of documents (and their contents)
on the network.

While every participant receives every document hash, access to some documents might be restricted to a limited set of parties.
In this case, the document contents will only be present in the DocumentLog of those parties.

Protocol
^^^^^^^^

The protocol layer is responsible for knowing how to query the network and how to respond to queries from peers.

P2P
^^^

The P2P layer is responsible building and maintaining network connections to our peers, authenticating them and to provide
a channel for the protocol to communicate with them (our peers).

Components
**********

The following diagram shows which components make up the engine and how they relate.

.. raw:: html
    :file: ../../_static/images/network_component_diagram.svg

Configuration
*************

The following diagram shows which configuration artifacts are used by the engine and where they come from.

.. raw:: html
    :file: ../../_static/images/network_configuration_artifact_diagram.svg
