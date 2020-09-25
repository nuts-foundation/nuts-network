nuts-network
#############

P2P network serving as a base layer for sharing data between Nuts nodes.

.. image:: https://circleci.com/gh/nuts-foundation/nuts-network.svg?style=svg
    :target: https://circleci.com/gh/nuts-foundation/nuts-network
    :alt: Build Status

.. image:: https://readthedocs.org/projects/nuts-network/badge/?version=latest
    :target: https://nuts-documentation.readthedocs.io/projects/nuts-network/en/latest/?badge=latest
    :alt: Documentation Status

.. image:: https://codecov.io/gh/nuts-foundation/nuts-network/branch/master/graph/badge.svg
    :target: https://codecov.io/gh/nuts-foundation/nuts-network
    :alt: Code coverage

.. image:: https://api.codeclimate.com/v1/badges/5475f4d4e696f43285b5/maintainability
   :target: https://codeclimate.com/github/nuts-foundation/nuts-network/maintainability
   :alt: Maintainability

The network is written in Go and should be part of nuts-go as an engine.

Dependencies
************

This projects is using go modules, so version > 1.12 is recommended. 1.10 would be a minimum.

Running tests
*************

Tests can be run by executing

.. code-block:: shell

    go test ./...

Building
********

This project is part of https://github.com/nuts-foundation/nuts-go. If you do however would like a binary, just use ``go build``.

The (internal) server and client API is generated from the open-api spec:

.. code-block:: shell

    oapi-codegen -generate types,server,client -package api docs/_static/nuts-network.yaml > api/generated.go

The peer-to-peer API uses gRPC. To generate Go code from the protobuf specs you need the `protoc-gen-go` package:

.. code-block:: shell

    go get -u github.com/golang/protobuf/protoc-gen-go

To generate the Go server and client code, run the following command:

.. code-block:: shell

    protoc -I network network/network.proto --go_out=plugins=grpc,paths=source_relative:network

To generate the mocks, run the following commands:

.. code-block:: shell

    ~/go/bin/mockgen -destination=pkg/mock.go -package=pkg -source=pkg/interface.go
    ~/go/bin/mockgen -destination=pkg/proto/mock.go -package=proto -source=pkg/proto/interface.go Protocol
    ~/go/bin/mockgen -destination=pkg/documentlog/mock.go -package=documentlog -source=pkg/documentlog/interface.go DocumentLog
    ~/go/bin/mockgen -destination=pkg/documentlog/store/mock.go -package=store -source=pkg/documentlog/store/interface.go DocumentStore
    ~/go/bin/mockgen -destination=pkg/nodelist/mock.go -package=nodelist -source=pkg/nodelist/interface.go NodeList
    ~/go/bin/mockgen -destination=pkg/p2p/mock.go -package=p2p -source=pkg/p2p/interface.go P2PNetwork

Binary format migrations
------------------------

The database migrations are packaged with the binary by using the ``go-bindata`` package.

.. code-block:: shell

    NOT_IN_PROJECT $ go get -u github.com/go-bindata/go-bindata/...
    nuts-network $ cd migrations && go-bindata -pkg migrations .

Running in Docker
*****************

Since nuts-network forms a p2p network it's useful to be able to quickly spawn a lot of nodes. The easiest way to do so it using the provided docker-compose file:

`docker-compose up`

It now should start a bootstrap node and a single (non-bootstrap) node.

To expand the network, add a few nodes:

`docker-compose up -d --scale node=5`

README
******

The readme is auto-generated from a template and uses the documentation to fill in the blanks.

.. code-block:: shell

    ./generate_readme.sh

This script uses ``rst_include`` which is installed as part of the dependencies for generating the documentation.

Documentation
*************

To generate the documentation, you'll need python3, sphinx and a bunch of other stuff. See :ref:`nuts-documentation-development-documentation`
The documentation can be build by running

.. code-block:: shell

    /docs $ make html

The resulting html will be available from ``docs/_build/html/index.html``

Configuration
*************

Parameters
==========

The following configuration parameters are available:

=======================  ===============  =================================================================================================================================================================
Key                      Default          Description
=======================  ===============  =================================================================================================================================================================
address                                   Interface and port for http server to bind to, defaults to global Nuts address.
bootstrapNodes                            Space-separated list of bootstrap nodes (`<host>:<port>`) which the node initially connect to.
certFile                                  PEM file containing the server certificate for the gRPC server. If not set the Nuts node won't start the gRPC server and other nodes will not be able to connect.
certKeyFile                               PEM file containing the private key of the server certificate. If not set the Nuts node won't start the gRPC server and other nodes will not be able to connect.
grpcAddr                 \:5555            Local address for gRPC to listen on.
mode                                      server or client, when client it uses the HttpClient
nodeID                                    Instance ID of this node under which the public address is registered on the nodelist. If not set, the Nuts node's identity will be used.
publicAddr                                Public address (of this node) other nodes can use to connect to it. If set, it is registered on the nodelist.
storageConnectionString  file:network.db  SQLite3 connection string to the database where the network should persist its documents.
=======================  ===============  =================================================================================================================================================================

