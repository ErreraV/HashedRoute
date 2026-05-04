# HashedRoute

Shipments on **Hyperledger Fabric** (**Go chaincode**), a **Go** REST API ([fabric-gateway](https://github.com/hyperledger/fabric-gateway)), and a **React** (Vite) UI.

## Prerequisites

- **Docker** + **Docker Compose v2** (CLI plugin).

Pinned stacks: Fabric **2.5.x** + gateway **v1.10.x** + contract API **v2.2.x** (see `backend/go.mod` and `chaincode/delivery/go.mod`).

## Configuration

Copy [`.env.example`](.env.example) to **`.env`** and set at least **`FABRIC_SAMPLES_ROOT`** to the directory that contains **`test-network/`** (often `~/go/src/github.com/<user>/fabric-samples` after the install step below). If you used **`make install FABRIC_GITHUB_USER=X`**, use the same **`FABRIC_GITHUB_USER=X`** on later **`make`** commands or put it in **`.env`**.

Compose uses **`FABRIC_SAMPLES_ROOT`** inside [`docker-compose.yml`](docker-compose.yml) for bind mounts to each org’s peer crypto. **`make up`** exports it from the Makefile (which reads **`.env`**).

Run **`make help`** anytime: it prints targets and the resolved **`FABRIC_SAMPLES_ROOT`**.

### Three orgs (fixed topology)

HashedRoute always runs **three** test-network orgs (**Org1**, **Org2**, **Org3**). Chaincode endorsement defaults to **`OR('Org1MSP.peer','Org2MSP.peer','Org3MSP.peer')`** (override with **`CC_END_POLICY`** when deploying).

Identities are **`User1@org1.example.com`** (and org2/org3), matching fabric-samples.

## Quick start (Make only)

From this repo root, in order:

| Step | Command |
|------|---------|
| 1. Install Fabric binaries, Docker images, and `fabric-samples` (once) | `make install` — optional: `FABRIC_GITHUB_USER=...` |
| 2. Start test network, create **mychannel**, add Org3, deploy **delivery** chaincode | `make setup` — same optional `FABRIC_GITHUB_USER=...` if install used it |
| 3. Run Docker UI/API stacks | `make up` |

Default **API** ports **8081–8083** and **web** ports **8091–8093** are defined in **`docker-compose.yml`** — change them there if needed. Stop: **`make down`**.

**`make setup`** runs **`fabric-network`**, **`fabric-add-org3`**, then **`fabric-deploy-chaincode`**.

If you already have **mychannel** without Org3, run **`make fabric-add-org3`**, then **`make fabric-deploy-chaincode SEQ=2`** (or the next sequence), then **`make up`**.

## Other Make targets

- **`make install`** — download and run Hyperledger **`install-fabric.sh`** (binaries, images, **`fabric-samples`**).
- **`make fabric-network`** — `./network.sh up createChannel` (Org1 + Org2 on **mychannel**).
- **`make fabric-add-org3`** — **`addOrg3.sh up`** (Org3 crypto, peer, join channel).
- **`make fabric-network-up`** — `./network.sh up` only (nodes only; channel must exist).
- **`make fabric-deploy-chaincode`** — deploy chaincode only (Org3 material must exist under **`test-network/organizations/.../org3.example.com`**).
- **`make seed`** — submit mock **CreateShipment** / **UpdateStatus** txs via the Org1 peer CLI.
- **`make clean`** — tear down test-network and delete org/channel artifacts on disk (then **`make setup`**).

Chaincode redeploy: bump **`SEQ`** (and usually **`CC_VER`**) when you change chaincode — see [`scripts/deploy-chaincode.sh`](scripts/deploy-chaincode.sh) env vars or pass them on the **`make fabric-deploy-chaincode`** line.

## Project layout

| Path | Role |
|------|------|
| [`chaincode/delivery/`](chaincode/delivery/) | `DeliveryContract` on the ledger |
| [`docker-compose.yml`](docker-compose.yml) | Three API + three web services; binds **`FABRIC_SAMPLES_ROOT`** into each container |
| [`scripts/check-fabric-msp.sh`](scripts/check-fabric-msp.sh) | **`make up`** precheck: **`User1@…`** MSP dirs exist for all three orgs |
| [`backend/`](backend/) | REST → Fabric Gateway |
| [`frontend/`](frontend/) | Operator UI |

## Troubleshooting

- **`make setup` / `make up` can’t find `fabric-samples`:** fix **`.env`** (`FABRIC_SAMPLES_ROOT`, **`FABRIC_GITHUB_USER`**). Run **`make help`** to see what path is used.
- **API errors reading MSP / certs:** each org’s **`User1@…`** tree must exist under **`test-network/organizations/peerOrganizations/`**. After **`make setup`**, Org3 is added automatically; for manual recovery run **`make fabric-add-org3`**.
- **`Unavailable` / `connection refused` on a peer port:** default host gRPC ports are **7051** / **9051** / **11051** for peers 1–3; they must match **`FABRIC_PEER_ENDPOINT`** in **`docker-compose.yml`**.
- **`ledger [mychannel] already exists` / join errors from `make fabric-network`:** bring up nodes with **`make fabric-network-up`**, then **`make fabric-add-org3`** / **`make fabric-deploy-chaincode`** as needed; or **`make clean`** and **`make setup`**.

Full Fabric teardown is done with **`fabric-samples/test-network`** scripts (not wrapped here).

## License

SPDX-License-Identifier: Apache-2.0
