# Quickstart Guide: Using Minio with SAONetwork

This guide will walk you through the process of using Minio with SAONetwork.

## Prerequisites

- Git
- Go (version 1.19 or later)

## Clone the Repository

First, clone the Minio repository from GitHub. You can do this by running the following command:

```bash
git clone git@github.com:SAONetwork/minio.git
```

## Checkout the SAONetwork Branch

Navigate into the cloned repository and checkout the `saonetwork` branch:

```bash
cd minio
git checkout saonetwork
```

## Build the Project

You can build the project by running the following command:

```bash
make build
```

## Configure Credentials

The default credentials for Minio are `minioadmin:minioadmin`. You can change these by setting the `MINIO_ROOT_USER` and `MINIO_ROOT_PASSWORD` environment variables.

## Configure SAOClient

Minio uses the existing configuration of SAOClient. For more information on how to configure SAOClient, refer to the [SAO Network CLI Tutorial](https://docs.sao.network/build-apps-on-sao-network/cli-tutorial#1.-initialize-a-cli-sao-client).

In addition to the basic configuration, you can also set the `MultiAddr` option in the `config.toml` file of your SAOClient. This option specifies the multiaddress of the libp2p node that the SAOClient will connect to. If `MultiAddr` is set, file uploading to SAO will go through libp2p, which is recommended for large file uploads.

Here's an example of how to set the `MultiAddr` option:

```toml
MultiAddr = "/ip4/127.0.0.1/udp/5154/quic/webtransport/certhash/uEiButuMTGfgB1bq7LbxNokyJERpX6YgWrinliofB1ZH4iw/certhash/uEiBBWmKBExCiVPuLG5ao0LUd-MxXT9Akat6gQdM3JVmomg/p2p/12D3KooWGNY3LBSyVYJ1ELS4B5WeHovZf16MiiDqp8vbtzC8tBR6"
```

Replace it with the multiaddress of your libp2p node.

## Start the Server

You can start the Minio server by running the following command:

```bash
./minio server /data_folder
```

Replace `/data_folder` with the path to the folder where you want Minio to store its data.

## Concept

In the context of SAONetwork, a Minio bucket maps to a group ID in SAO, and an object name in Minio maps to an alias in SAO. This mapping allows you to interact with SAONetwork using familiar S3 operations.

You should now have a working Minio server that's configured to work with SAONetwork. You can use the Minio client or any S3-compatible client to interact with your Minio server.
