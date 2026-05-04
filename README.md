# HashedRoute

Shipments on **Hyperledger Fabric** (**Go chaincode**), a **Go** REST API ([fabric-gateway](https://github.com/hyperledger/fabric-gateway)), and a **React** (Vite) UI.

## Prerequisites

- **Docker** + **Docker Compose v2** (CLI plugin).

Pinned stacks: Fabric **2.5.x** + gateway **v1.10.x** + contract API **v2.2.x** (see `backend/go.mod` and `chaincode/delivery/go.mod`).

## Configuration

Copy [`.env.example`](.env.example) to **`.env`** and set at least **`FABRIC_SAMPLES_ROOT`** to the directory that contains **`test-network/`** (often `~/go/src/github.com/<user>/fabric-samples` after the install step below). If you used **`make fabric-install-hyperledger FABRIC_GITHUB_USER=X`**, use the same **`FABRIC_GITHUB_USER=X`** on later **`make`** commands or put it in **`.env`**.

Run **`make help`** anytime: it prints targets and the resolved **`FABRIC_SAMPLES_ROOT`**.

### Three hosts (fixed topology)

HashedRoute always runs **three** test-network orgs (**Org1**, **Org2**, **Org3**): three peers on the channel, chaincode endorsement **`OR('Org1MSP.peer','Org2MSP.peer','Org3MSP.peer')`** (override with **`CC_END_POLICY`** on deploy if you must).

**[`fabric-hosts.def`](fabric-hosts.def)** must contain **exactly three** data rows, in order **org1 → org2 → org3**, with fixed MSPs, crypto dirs, gateway peers, and host peer gRPC ports (**7051** / **9051** / **11051**). You may change only the last two columns per row (**API_PORT**, **WEB_PORT**) for the Docker **api-*** / **web-*** services.

Default identities use **`User1@org1.example.com`** (and org2/org3), matching fabric-samples.

Optional: **`FABRIC_HOSTS_DEF=/path/to/file make compose-gen`** if the def file lives outside the repo (same three-row format).

## Quick start (Make only)

From this repo root, in order:

| Step | Command |
|------|---------|
| 1. Install Fabric binaries, Docker images, and `fabric-samples` (once) | `make fabric-install-hyperledger` — optional: `FABRIC_GITHUB_USER=...` |
| 2. Start test network, create **mychannel**, add Org3, deploy **delivery** chaincode | `make fabric-setup` — same optional `FABRIC_GITHUB_USER=...` if install used it |
| 3. Run Docker UI/API stacks | `make up` |

Open each UI/API using the **API_PORT** and **WEB_PORT** from **`fabric-hosts.def`**. Stop: **`make down`**.

**`make fabric-setup`** runs **`fabric-network`** (Org1 + Org2 channel create), **`fabric-add-org3`** (`addOrg3.sh up`), then **`fabric-deploy-chaincode`**. From a **clean** ledger you do not need to run **`addOrg3`** by hand.

If you already have **mychannel** without Org3, run **`make fabric-add-org3`**, then **`make fabric-deploy-chaincode SEQ=2`** (or the next sequence), then **`make up`**.

## Other Make targets

- **`make compose-gen`** — regenerate **`docker-compose.yml`** from **`fabric-hosts.def`** (port edits only).
- **`make fabric-network`** — `./network.sh up createChannel` (Org1 + Org2 on **mychannel**).
- **`make fabric-add-org3`** — **`addOrg3.sh up`** (Org3 crypto, peer, join channel).
- **`make fabric-network-up`** — `./network.sh up` only (nodes only; channel must exist).
- **`make fabric-deploy-chaincode`** — deploy chaincode only (Org3 material must exist under **`test-network/organizations/.../org3.example.com`**).

Chaincode redeploy: bump **`SEQ`** (and usually **`CC_VER`**) when you change chaincode — see [`scripts/deploy-chaincode.sh`](scripts/deploy-chaincode.sh) env vars or pass them on the **`make fabric-deploy-chaincode`** line.

## Project layout

| Path | Role |
|------|------|
| [`chaincode/delivery/`](chaincode/delivery/) | `DeliveryContract` on the ledger |
| [`fabric-hosts.def`](fabric-hosts.def) | Three fixed org rows; optional API/web port changes |
| [`scripts/lib-fabric-hosts.sh`](scripts/lib-fabric-hosts.sh) | Validates **`fabric-hosts.def`**; default endorsement policy constant |
| [`scripts/gen-docker-compose-hosts.sh`](scripts/gen-docker-compose-hosts.sh) | Writes **`docker-compose.yml`** |
| [`backend/`](backend/) | REST → Fabric Gateway |
| [`frontend/`](frontend/) | Operator UI |
| [`docker-compose.yml`](docker-compose.yml) | Generated app stack (three API + three web services) |

## Troubleshooting

- **`make fabric-setup` / `make up` can’t find `fabric-samples`:** fix **`.env`** (`FABRIC_SAMPLES_ROOT`, **`FABRIC_GITHUB_USER`**). Run **`make help`** to see what path is used.
- **API errors reading MSP / certs:** each org’s **`User1@…`** tree must exist under **`test-network/organizations/peerOrganizations/`**. After **`make fabric-setup`**, Org3 is added automatically; for manual recovery run **`make fabric-add-org3`** (from **`test-network/addOrg3`**).
- **`fabric-hosts.def` validation errors:** the file must stay three rows with the exact org1/org2/org3 fields documented in the file header; only the last two port columns are free to change.
- **`Unavailable` / `connection refused` on a peer port:** host gRPC ports are fixed at **7051** / **9051** / **11051** for peers 1–3; confirm **`docker ps`** matches the test-network.
- **`ledger [mychannel] already exists` / join errors from `make fabric-network`:** bring up nodes with **`make fabric-network-up`**, then **`make fabric-add-org3`** / **`make fabric-deploy-chaincode`** as needed; or **`make fabric-clean-ledger`** and **`make fabric-setup`**.

Full Fabric teardown is done with **`fabric-samples/test-network`** scripts (not wrapped here).

## License

SPDX-License-Identifier: Apache-2.0
