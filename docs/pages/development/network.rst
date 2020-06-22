.. _nuts-network-development:

Nuts Network development
##########################

.. marker-for-readme

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
    ~/go/bin/mockgen -destination=pkg/nodelist/mock.go -package=nodelist -source=pkg/nodelist/interface.go NodeList
    ~/go/bin/mockgen -destination=pkg/p2p/mock.go -package=p2p -source=pkg/p2p/interface.go P2PNetwork

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