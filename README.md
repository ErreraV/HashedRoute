# HashedRoute

Delivery logistics demo on **Hyperledger Fabric**: **Go chaincode** for shipment state, a **Go REST API** using [fabric-gateway-go](https://github.com/hyperledger/fabric-gateway), and a **React** (Vite + TypeScript) dashboard.

## Version pins (keep these aligned)

| Component | Version / notes |
|-----------|-----------------|
| Fabric network | **2.5.x LTS** test network from [fabric-samples](https://github.com/hyperledger/fabric-samples) (use a samples tag/branch that matches your installed `peer` + Docker images). |
| Chaincode API | `github.com/hyperledger/fabric-contract-api-go/v2` **v2.2.0**, `fabric-chaincode-go/v2` **v2.0.0** |
| Application SDK | `github.com/hyperledger/fabric-gateway` **v1.10.x** |
| Go toolchain | **1.23+** recommended (matches current fabric-samples REST sample). |

If binaries and Fabric container images disagree, installs and deploys will fail in confusing ways—always install the Fabric **peer**, **orderer**, **config**, and **docker images** from the same release.

## Repo layout

- [`chaincode/delivery/`](chaincode/delivery/) – `DeliveryContract` (create shipment, list, read, update status with a strict state machine).
- [`backend/`](backend/) – HTTP API on `:8080` (configurable) that submits/evaluates transactions through the Fabric Gateway.
- [`frontend/`](frontend/) – Operator UI (create, list, track, advance status).
- [`docker-compose.yml`](docker-compose.yml) + [`Makefile`](Makefile) – build/run **API + UI in Docker** (Fabric test-network still runs on the host).
- [`scripts/deploy-chaincode.sh`](scripts/deploy-chaincode.sh) – Wrapper around `network.sh deployCC` for this chaincode.

## Prerequisites

- **Docker** + **Docker Compose** (required by the Fabric test network).
- **Fabric binaries, Docker images, and samples**: follow the [Install docs](https://hyperledger-fabric.readthedocs.io/en/latest/install.html), or use the same flow as the doc in one step:
  ```bash
  make fabric-install-hyperledger
  ```
  It downloads `install-fabric.sh` from [fabric `main`](https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh) and runs `./install-fabric.sh -f 2.5.15 -c 1.5.17` (override with `FABRIC_VERSION`, `FABRIC_CA_VERSION`). With no `FABRIC_INSTALL_COMPONENTS`, the script installs **all** components; otherwise pass e.g. `FABRIC_INSTALL_COMPONENTS=docker binary samples`.
  This creates `$HOME/go/src/github.com/<FABRIC_GITHUB_USER>/` (default `FABRIC_GITHUB_USER=$(USER)`). If you pass a **different** `FABRIC_GITHUB_USER` than your login, put that same value in HashedRoute **`.env`** (or pass it to later `make fabric-setup` / `make up`) so the Makefile can find `fabric-samples`.
- **Go** 1.23+ and **Node.js** LTS (for the UI).

If you installed under the Go path above but **HashedRoute** lives elsewhere, the **Makefile auto-picks** `fabric-samples` from `../fabric-samples` when present; otherwise from `$HOME/go/src/github.com/$(FABRIC_GITHUB_USER)/fabric-samples` (same default user as `make fabric-install-hyperledger`). If it still fails, set **`FABRIC_SAMPLES_ROOT`** in `.env` to the folder that contains **`test-network/`**.

Recommended directory layout (default crypto path in the API assumes this):

```text
parent/
  fabric-samples/     # test-network lives here
  HashedRoute/        # this repository
```

## 1. Start the test network and channel

From **fabric-samples**, or from this repo (defaults to `../fabric-samples`):

```bash
cd fabric-samples/test-network
./network.sh up createChannel
```

Or equivalently:

```bash
make fabric-network
```

Leave the network running.

## 2. Deploy the delivery chaincode

From your machine (any cwd), point `FABRIC_TEST_NETWORK` at `fabric-samples/test-network` and run:

```bash
export FABRIC_TEST_NETWORK="$HOME/path/to/fabric-samples/test-network"
export HASHEDRO_HOME="$HOME/path/to/HashedRoute"
chmod +x "$HASHEDRO_HOME/scripts/deploy-chaincode.sh"
"$HASHEDRO_HOME/scripts/deploy-chaincode.sh"
```

Or run **both** §1 and §2 in order (after cloning `fabric-samples` next to this repo, or set `FABRIC_SAMPLES_ROOT`):

```bash
make fabric-setup
```

Defaults: chaincode name **`delivery`**, version **`1.0`**, sequence **`1`**. Override with `CC_NAME`, `CC_VER`, `SEQ` if needed. After a code change, redeploy with a higher **`SEQ`** (and matching policy/version per Fabric lifecycle rules).

### Verify with the peer CLI (optional)

With Fabric `peer` and `CORE_PEER_*` variables set for Org1 (see fabric-samples `scripts/envVar.sh` and `setGlobals 1`), you can evaluate a read:

```bash
peer chaincode query -C mychannel -n delivery \
  -c '{"function":"DeliveryContract:GetShipment","Args":["SHP-1"]}'
```

If your Fabric CLI version expects a capital `Function` key instead of `function`, adjust accordingly. When in doubt, use the REST API responses in section 6—they exercise the same chaincode paths as the Gateway.

## 3. Run API + UI in Docker (`make`)

The **Fabric peer/orderer network stays on the host** (your usual `fabric-samples/test-network` Docker stack). This project’s **Go API and React UI** run in separate app containers: crypto is bind-mounted read-only, and the API reaches the peer at `host.docker.internal:7051` (`extra_hosts: host-gateway` works on Linux too).

From the **HashedRoute** repository root:

```bash
make help          # targets and path hints
make up            # build + start api (8080) and web (8081)
```

- **UI** (nginx + static build): [http://localhost:8081](http://localhost:8081) — the browser calls **`/api/...`** on the same origin; nginx proxies to the API container.
- **API** directly: [http://localhost:8080](http://localhost:8080)

Paths default to `../fabric-samples/.../org1.example.com`. Override if needed:

```bash
export FABRIC_SAMPLES_ROOT=/path/to/fabric-samples
make up
```

Or copy [`.env.example`](.env.example) to `.env` and edit. Optional: `API_PORT`, `WEB_PORT`, `FABRIC_HOST_PEER_PORT` if the peer gRPC port on the host is not `7051`.

Other targets: `make down`, `make logs`, `make ps`, `make restart`, `make build-docker`, `make clean-docker`, `make dev-api`, `make dev-web`.

## 4. Run the Go API (native, no Docker)

```bash
cd HashedRoute/backend
go run .
```

By default the process resolves Org1 crypto material under:

`../../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com`

Run **from** the `backend/` directory so that relative path works, or override with env vars.

### Environment variables

| Variable | Purpose |
|----------|---------|
| `HTTP_ADDR` | Listen address (default `:8080`). |
| `FABRIC_CRYPTO_PATH` | Root MSP folder for Org1 (see default above). |
| `FABRIC_CERT_PATH` | Signing cert PEM (defaults under `.../users/User1@org1.example.com/msp/signcerts/`). |
| `FABRIC_KEY_DIR` | Private keystore directory for User1. |
| `FABRIC_TLS_CA` | Peer TLS CA cert (`.../peers/peer0.org1.example.com/tls/ca.crt`). |
| `FABRIC_PEER_ENDPOINT` | gRPC endpoint (default `dns:///localhost:7051`). |
| `FABRIC_GATEWAY_PEER` | TLS server name (default `peer0.org1.example.com`). |
| `FABRIC_MSP_ID` | MSP ID (default `Org1MSP`). |
| `FABRIC_CHANNEL` | Channel (default `mychannel`). |
| `FABRIC_CHAINCODE` | Deployed name (default `delivery`). |
| `FABRIC_CONTRACT_NAME` | Contract class name (default `DeliveryContract`). |

## 5. Run the React UI (native, no Docker)

```bash
cd HashedRoute/frontend
npm install
npm run dev
```

Open the printed local URL (default `http://localhost:5173`). CORS on the API allows that origin.

To point at a non-default API URL:

```bash
VITE_API_URL=http://localhost:9090 npm run dev
```

## 6. API smoke test (curl)

```bash
curl -s http://localhost:8080/health

curl -s http://localhost:8080/api/shipments

curl -s -X POST http://localhost:8080/api/shipments \
  -H 'Content-Type: application/json' \
  -d '{"id":"SHP-1","origin":"NYC","destination":"AUS","customer":"Alice","carrier":"DemoFreight"}'

curl -s http://localhost:8080/api/shipments/SHP-1

curl -s -X PATCH http://localhost:8080/api/shipments/SHP-1/status \
  -H 'Content-Type: application/json' \
  -d '{"status":"PICKED_UP","notes":"handoff at dock 4"}'
```

Valid status progression enforced on-chain: **CREATED → PICKED_UP → IN_TRANSIT → DELIVERED**.

When using Docker, the same routes are available via nginx, for example:

```bash
curl -s http://localhost:8081/api/shipments
```

## Troubleshooting

- **`failed to read ... cert.pem` / keystore**: confirm the test network is up and paths match your machine; set `FABRIC_CRYPTO_PATH` explicitly.
- **`chaincode not found` / endorsement errors**: redeploy with correct `-ccn` / sequence; ensure channel name matches `FABRIC_CHANNEL`.
- **`commit failed` / policy**: confirm both org peers endorse if your endorsement policy requires Org1 and Org2 (default deploy script uses both).
- **UI `Failed to fetch`**: start the API first; check browser DevTools network tab and CORS (native Vite UI origin must be `http://localhost:5173` unless you change [`backend/handlers.go`](backend/handlers.go)). The Docker UI uses same-origin `/api` and does not rely on CORS.
- **Docker API cannot reach peer**: confirm Fabric publishes `7051` on the host, `FABRIC_HOST_PEER_PORT` matches, and nothing blocks `host.docker.internal` (corporate Docker setups sometimes need extra configuration).
- **`hashedroute-api` / `hashedroute-web` stuck in `Restarting`**: rebuild after pulling latest images (`make down` then `make up`). Inspect logs with `docker compose logs api` and `docker compose logs web`. Typical causes: API cannot read bind-mounted MSP files (image runs as root for local dev), API cannot reach the peer (Fabric down or wrong port), or nginx could not resolve `api` at startup (Compose should use lazy DNS + `resolver 127.0.0.11` in [`frontend/nginx.conf`](frontend/nginx.conf)).

SPDX-License-Identifier: Apache-2.0 (align with Hyperledger components used in your course if needed).
