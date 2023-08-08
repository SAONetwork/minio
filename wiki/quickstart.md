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

Minio uses the existing configuration of SAOClient. Depending on the network you want to connect to (beta network for example), you can initialize the SAOClient with the following commands:

### For Beta Network:

```bash
./saoclient --chain-address https://rpc-beta.sao.network:443 --gateway https://gateway-beta.sao.network/rpc/v0 init --chain-id sao-20230629 --key-name nancy
```

### For a Custom Node:

If you have your own node, you can initialize the SAOClient with the following command, replacing the chain address and gateway with your node's specific addresses. Also, specify the chain ID according to the network you want to connect to (`sao-20230629` for beta and `sao-testnet1` for testnet):

```bash
./saoclient --chain-address YOUR_CHAIN_ADDRESS --gateway YOUR_GATEWAY_ADDRESS init --chain-id YOUR_CHAIN_ID --key-name nancy
```

Replace `YOUR_CHAIN_ADDRESS` with the address of your chain (e.g., `http://127.0.0.1:26657`), `YOUR_GATEWAY_ADDRESS` with the address of your gateway (e.g., `http://127.0.0.1:5151/rpc/v0`), and `YOUR_CHAIN_ID` with the chain ID of the network you want to connect to (`sao-20230629` for beta).

After successful initialization, by default the configuration should be located at `~/.sao-cli/config.toml` and look like this:

```toml
GroupId = "YOUR_GROUP_ID"
KeyName = "nancy"
ChainAddress = "https://rpc-beta.sao.network:443"
Gateway = "https://gateway-beta.sao.network/rpc/v0"
Token = "DEFAULT_TOKEN"
```

### Configuring MultiAddr:

In addition to the basic configuration, you can also set the `MultiAddr` option in the `config.toml` file of your SAOClient. This option specifies the multiaddress of the libp2p node that the SAOClient will connect to. If `MultiAddr` is set, file uploading to SAO will go through libp2p, which is recommended for large file uploads.

Here's an example of how to set the `MultiAddr` option:

```toml
MultiAddr = "/ip4/8.222.225.178/tcp/5153/p2p/12D3KooWJA2R7RTd6aD2pUdvjN29FdiC8f5edSifXA2tXBcbA2UX"
```

You can find the MultiAddr by running the following command, replacing `YOUR_GATEWAY_NODE_ADDRESS` with your specific node address:

```bash
saod query node show-node YOUR_GATEWAY_NODE_ADDRESS
```

Alternatively, you can use a curl command to find the MultiAddr:

```bash
curl -X GET "https://api-beta.sao.network/SaoNetwork/sao/node/node/YOUR_GATEWAY_NODE_ADDRESS" -H "accept: application/json"
```


Replace `YOUR_GATEWAY_NODE_ADDRESS` with the specific address of your node in the curl command as well.

Look for a TCP peer in the `peer` field, such as `/ip4/8.222.225.178/tcp/5153/p2p/12D3KooWJA2R7RTd6aD2pUdvjN29FdiC8f5edSifXA2tXBcbA2UX`

## Start the Server

You can start the Minio server by running the following command:

```bash
./minio server /data_folder
```

Replace `/data_folder` with the path to the folder where you want Minio to store its data.

## Concept

In the context of SAONetwork, a Minio bucket maps to a group ID in SAO, and an object name in Minio maps to an alias in SAO. This mapping allows you to interact with SAONetwork using familiar S3 operations.

You should now have a working Minio server that's configured to work with SAONetwork. You can use the Minio client or any S3-compatible client to interact with your Minio server.
