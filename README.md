# HashedRoute

Shipments on **Hyperledger Fabric** (**Go chaincode**), a **Go** REST API ([fabric-gateway](https://github.com/hyperledger/fabric-gateway)), and a **React** (Vite) UI.

## Prerequisites

- **Docker** + **Docker Compose v2.20+** (`include` in [`docker-compose.yml`](docker-compose.yml); upgrade Compose if `make up` fails on `include`).
- **Go** and **Node** if you use `make dev-api` / `make dev-web` instead of `make up`.

Pinned stacks: Fabric **2.5.x** + gateway **v1.10.x** + contract API **v2.2.x** (see `backend/go.mod` and `chaincode/delivery/go.mod`).

## Configuration

Copy [`.env.example`](.env.example) to **`.env`** and set at least **`FABRIC_SAMPLES_ROOT`** to the directory that contains **`test-network/`** (often `~/go/src/github.com/<user>/fabric-samples` after the install step below). If you used **`make fabric-install-hyperledger FABRIC_GITHUB_USER=X`**, use the same **`FABRIC_GITHUB_USER=X`** on later **`make`** commands or put it in **`.env`**.

Run **`make help`** anytime: it prints targets and the resolved **`FABRIC_SAMPLES_ROOT`**.

### Many UI/API stacks (same ledger)

**[`fabric-hosts.def`](fabric-hosts.def)** is the single place that controls how many Docker **api-*** / **web-*** stacks exist, their ports, and which MSP and peer each stack uses. See [`fabric-hosts.def.example`](fabric-hosts.def.example) for column meanings. The first column (**HOST_KEY**) must be **unique** on every row—Docker Compose service names are `api-<KEY>` / `web-<KEY>`.

- **Add or remove hosts:** edit **`fabric-hosts.def`** only, then **`make up`** (which regenerates [`docker-compose.hosts.yml`](docker-compose.hosts.yml) from that file).
- **`make up`** also runs [`scripts/ensure-fabric-peer-orgs.sh`](scripts/ensure-fabric-peer-orgs.sh): if **`fabric-hosts.def`** references **`org3.example.com`** but **`User1@org3…/msp`** is missing, it runs **`test-network/addOrg3/addOrg3.sh up`** (channel **mychannel** and Org1/Org2 must already exist). Other org directory names are not auto-generated—use a custom network or add material yourself.
- **`make fabric-deploy-chaincode`** (and [`scripts/deploy-chaincode.sh`](scripts/deploy-chaincode.sh)) builds the default endorsement policy as **`OR('<MSP>.peer',...)`** from the **unique MSP_ID** values in **`fabric-hosts.def`**, unless you set **`CC_END_POLICY`**. Keep that list aligned with orgs that are actually on your channel.
- Identities use **`User1@`** plus **`CRYPTO_PEER_DIR`** (same layout as fabric-samples test-network).

Optional: **`FABRIC_HOSTS_DEF=/path/to/hosts.def make up`** (and the same variable on **`make fabric-deploy-chaincode`** if you use a non-default path). Run **`make fabric-ensure-peer-orgs`** alone to only execute the Org3 provisioning step.

## Quick start (Make only)

From this repo root, in order:

| Step | Command |
|------|---------|
| 1. Install Fabric binaries, Docker images, and `fabric-samples` (once) | `make fabric-install-hyperledger` — optional: `FABRIC_GITHUB_USER=...` |
| 2. Start test network, create channel, deploy **delivery** chaincode | `make fabric-setup` — same optional `FABRIC_GITHUB_USER=...` if install used it |
| 3. Run Docker UI/API stacks | `make up` |

Open each UI/API using the **API_PORT** and **WEB_PORT** columns from your **`fabric-hosts.def`**. Stop: **`make down`**.

**Local dev instead of Docker:** `make dev-api` in one terminal, **`make dev-web`** in another; use the URL Vite prints (CORS allows **`http://localhost:5173`**).

## Other Make targets

- **`make compose-gen`** — regenerate **`docker-compose.hosts.yml`** from **`fabric-hosts.def`** without starting containers.
- **`make fabric-network`** — `./network.sh up createChannel` (first channel create + join); see troubleshooting if **`mychannel` already exists**.
- **`make fabric-network-up`** — `./network.sh up` only (bring up nodes; use when the channel is already created).
- **`make fabric-ensure-peer-orgs`** — run [`scripts/ensure-fabric-peer-orgs.sh`](scripts/ensure-fabric-peer-orgs.sh) only (Org3 via **`addOrg3.sh`** when **`fabric-hosts.def`** requires **`org3.example.com`** and MSP is missing).
- **`make fabric-deploy-chaincode`** — deploy chaincode only (network must be up). Default endorsement policy follows **`fabric-hosts.def`** MSPs unless **`CC_END_POLICY`** is set.

Chaincode redeploy: bump **`SEQ`** (and usually **`CC_VER`**) when you change chaincode or which orgs endorse—see [`scripts/deploy-chaincode.sh`](scripts/deploy-chaincode.sh) env vars or pass them on the **`make fabric-deploy-chaincode`** line.

## Project layout

| Path | Role |
|------|------|
| [`chaincode/delivery/`](chaincode/delivery/) | `DeliveryContract` on the ledger |
| [`fabric-hosts.def`](fabric-hosts.def) | Declarative list of Docker **api-**/ **web-** stacks (ports, MSP, peer); [`fabric-hosts.def.example`](fabric-hosts.def.example) |
| [`scripts/gen-docker-compose-hosts.sh`](scripts/gen-docker-compose-hosts.sh) | Generates **`docker-compose.hosts.yml`** from **`fabric-hosts.def`** |
| [`scripts/ensure-fabric-peer-orgs.sh`](scripts/ensure-fabric-peer-orgs.sh) | **`make up`**: create missing **`org3.example.com`** peer org material via fabric-samples **`addOrg3`** |
| [`backend/`](backend/) | REST → Fabric Gateway |
| [`frontend/`](frontend/) | Operator UI |
| [`docker-compose.yml`](docker-compose.yml) | **`include`s** generated [`docker-compose.hosts.yml`](docker-compose.hosts.yml) (from **`fabric-hosts.def`**) |

## Troubleshooting

- **`make fabric-setup` / `make up` can’t find `fabric-samples`:** fix **`.env`** (`FABRIC_SAMPLES_ROOT`, **`FABRIC_GITHUB_USER`**). Run **`make help`** to see what path is used.
- **API errors reading MSP / certs:** every row in **`fabric-hosts.def`** must have matching **`User1@…`** under **`test-network/organizations/peerOrganizations/`** (`make up` runs a check per line). **`make up`** will try to create Org3 material with **`addOrg3.sh`** when the third column is **`org3.example.com`**; Org1/Org2 require **`make fabric-network`** / **`fabric-setup`** first.
- **`Unavailable` / `connection refused` on a peer port:** match **`HOST_PEER_PORT`** in **`fabric-hosts.def`** to the gRPC port your org’s peer publishes on the host (see **`docker ps`**).
- **Endorsement or commit failures after changing `fabric-hosts.def`:** redeploy chaincode with a higher **`SEQ`** (and a policy that matches channel orgs), or trim the file to orgs that exist on the channel.
- **`ledger [mychannel] already exists` / join channel errors from `make fabric-network`:** peers are already on the channel. Bring up nodes with **`make fabric-network-up`** if needed, then **`make fabric-deploy-chaincode`**; or **`make fabric-clean-ledger`** and **`make fabric-setup`** for a clean slate.

Full Fabric teardown is done with **`fabric-samples/test-network`** scripts (not wrapped here).

## License

SPDX-License-Identifier: Apache-2.0
