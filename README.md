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

Edit **[`fabric-hosts.def`](fabric-hosts.def)** (see [`fabric-hosts.def.example`](fabric-hosts.def.example)): one data row per stack (**MSP**, crypto directory under `peerOrganizations/`, **gateway peer**, host gRPC port, **API port**, **UI port**). You do not change backend, frontend, or chaincode—only this file—then run **`make up`** (which regenerates **[`docker-compose.hosts.yml`](docker-compose.hosts.yml)**). Identities use **`User1@`** the same name as `CRYPTO_PEER_DIR` (for example **`User1@org1.example.com`**) as in fabric-samples test-network.

Optional: **`FABRIC_HOSTS_DEF=/path/to/hosts.def make up`**.

## Quick start (Make only)

From this repo root, in order:

| Step | Command |
|------|---------|
| 1. Install Fabric binaries, Docker images, and `fabric-samples` (once) | `make fabric-install-hyperledger` — optional: `FABRIC_GITHUB_USER=...` |
| 2. Start test network, create channel, deploy **delivery** chaincode | `make fabric-setup` — same optional `FABRIC_GITHUB_USER=...` if install used it |
| 3. Run Docker UI/API stacks | `make up` — reads **`fabric-hosts.def`**, writes **`docker-compose.hosts.yml`**, one **api-** / **web-** pair per row (**HOST_KEY** column) |

Then open each UI/API URL from your **`fabric-hosts.def`** (defaults: Org1 **<http://localhost:8081>** and API **8080**; Org2 **<http://localhost:8091>** and **8090**). Stop: **`make down`**.

**Local dev instead of Docker:** `make dev-api` in one terminal, **`make dev-web`** in another; use the URL Vite prints (CORS allows **`http://localhost:5173`**).

## Other Make targets

- **`make compose-gen`** — regenerate **`docker-compose.hosts.yml`** from **`fabric-hosts.def`** without starting containers.
- **`make fabric-network`** — only bring up network + channel (no chaincode deploy).
- **`make fabric-deploy-chaincode`** — only deploy chaincode (network must already be up).

Chaincode redeploy: bump **`SEQ`** (and usually **`CC_VER`**) when you change chaincode—see [`scripts/deploy-chaincode.sh`](scripts/deploy-chaincode.sh) env vars or pass them on the **`make fabric-deploy-chaincode`** line if you extend the Makefile.

## Project layout

| Path | Role |
|------|------|
| [`chaincode/delivery/`](chaincode/delivery/) | `DeliveryContract` on the ledger |
| [`fabric-hosts.def`](fabric-hosts.def) | Declarative list of Docker **api-**/ **web-** stacks (ports, MSP, peer); [`fabric-hosts.def.example`](fabric-hosts.def.example) |
| [`scripts/gen-docker-compose-hosts.sh`](scripts/gen-docker-compose-hosts.sh) | Generates **`docker-compose.hosts.yml`** from **`fabric-hosts.def`** |
| [`backend/`](backend/) | REST → Fabric Gateway |
| [`frontend/`](frontend/) | Operator UI |
| [`docker-compose.yml`](docker-compose.yml) | **`include`s** generated [`docker-compose.hosts.yml`](docker-compose.hosts.yml) (from **`fabric-hosts.def`**) |

## Troubleshooting

- **`make fabric-setup` / `make up` can’t find `fabric-samples`:** fix **`.env`** (`FABRIC_SAMPLES_ROOT`, **`FABRIC_GITHUB_USER`**). Run **`make help`** to see what path is used.
- **API errors reading MSP / certs:** run **`make fabric-setup`** (or **`fabric-network`**) so every org in **`fabric-hosts.def`** has **`User1@…`** under **`peerOrganizations/`** (`make up` checks each line).
- **`Unavailable` / `connection refused` … `:7051` (or `:9051`) from the API container:** the test-network peer on the host is not accepting gRPC on that port—start it with **`make fabric-network`** or **`make fabric-setup`**, then confirm **`docker ps`** shows `peer0.org1.example.com` / `peer0.org2.example.com`.
- **Native API:** use **`make dev-api`** (sets **`FABRIC_CRYPTO_PATH`**) rather than raw **`go run`** unless you export paths yourself.

Full Fabric teardown is done with **`fabric-samples/test-network`** scripts (not wrapped here).

## License

SPDX-License-Identifier: Apache-2.0
