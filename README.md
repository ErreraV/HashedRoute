# HashedRoute

Shipments on **Hyperledger Fabric** (**Go chaincode**), a **Go** REST API ([fabric-gateway](https://github.com/hyperledger/fabric-gateway)), and a **React** (Vite) UI.

## Prerequisites

- **Docker** + **Docker Compose** (Fabric test network).
- **Go** and **Node** if you use `make dev-api` / `make dev-web` instead of `make up`.

Pinned stacks: Fabric **2.5.x** + gateway **v1.10.x** + contract API **v2.2.x** (see `backend/go.mod` and `chaincode/delivery/go.mod`).

## Configuration

Copy [`.env.example`](.env.example) to **`.env`** and set at least **`FABRIC_SAMPLES_ROOT`** to the directory that contains **`test-network/`** (often `~/go/src/github.com/<user>/fabric-samples` after the install step below). If you used **`make fabric-install-hyperledger FABRIC_GITHUB_USER=X`**, use the same **`FABRIC_GITHUB_USER=X`** on later **`make`** commands or put it in **`.env`**.

Run **`make help`** anytime: it prints targets and the resolved **`FABRIC_SAMPLES_ROOT`**.

## Quick start (Make only)

From this repo root, in order:

| Step | Command |
|------|---------|
| 1. Install Fabric binaries, Docker images, and `fabric-samples` (once) | `make fabric-install-hyperledger` — optional: `FABRIC_GITHUB_USER=...` |
| 2. Start test network, create channel, deploy **delivery** chaincode | `make fabric-setup` — same optional `FABRIC_GITHUB_USER=...` if install used it |
| 3. Run API + UI in Docker | `make up` |

Then open **<http://localhost:8081>** (UI; nginx proxies **`/api`**). API directly: **<http://localhost:8080>**. Stop app containers: **`make down`**.

**Local dev instead of Docker:** `make dev-api` in one terminal, **`make dev-web`** in another; use the URL Vite prints (CORS allows **`http://localhost:5173`**).

## Other Make targets

- **`make fabric-network`** — only bring up network + channel (no chaincode deploy).
- **`make fabric-deploy-chaincode`** — only deploy chaincode (network must already be up).

Chaincode redeploy: bump **`SEQ`** (and usually **`CC_VER`**) when you change chaincode—see [`scripts/deploy-chaincode.sh`](scripts/deploy-chaincode.sh) env vars or pass them on the **`make fabric-deploy-chaincode`** line if you extend the Makefile.

## Project layout

| Path | Role |
|------|------|
| [`chaincode/delivery/`](chaincode/delivery/) | `DeliveryContract` on the ledger |
| [`backend/`](backend/) | REST → Fabric Gateway |
| [`frontend/`](frontend/) | Operator UI |
| [`docker-compose.yml`](docker-compose.yml) | **api** + **web** containers (Fabric stays on the host) |

## Troubleshooting

- **`make fabric-setup` / `make up` can’t find `fabric-samples`:** fix **`.env`** (`FABRIC_SAMPLES_ROOT`, **`FABRIC_GITHUB_USER`**). Run **`make help`** to see what path is used.
- **API errors reading MSP / certs:** bring up **`make fabric-network`** (or full **`fabric-setup`**) from the same **`fabric-samples`** tree so Org1 **User1** exists.
- **Containers in `Restarting`:** **`make down`** then **`make up`**; use Docker / Compose to inspect logs if it persists.
- **Native API:** use **`make dev-api`** (sets **`FABRIC_CRYPTO_PATH`**) rather than raw **`go run`** unless you export paths yourself.

Full Fabric teardown is done with **`fabric-samples/test-network`** scripts (not wrapped here).

## License

SPDX-License-Identifier: Apache-2.0
