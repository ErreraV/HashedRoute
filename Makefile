.PHONY: help up down dev-api dev-web \
	fabric-samples-precheck fabric-network fabric-deploy-chaincode fabric-setup \
	fabric-install-hyperledger

.DEFAULT_GOAL := help

# Repo root
ROOT := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
-include $(ROOT)/.env

# Where install-fabric.sh typically clones fabric-samples (see fabric-install-hyperledger).
FABRIC_GITHUB_USER ?= $(USER)
FABRIC_INSTALL_WORKDIR ?= $(HOME)/go/src/github.com/$(FABRIC_GITHUB_USER)

# Prefer ../fabric-samples when it has real crypto (User1 MSP); otherwise install-fabric.sh clone.
# Stale/partial clones beside the repo often have network.sh but no org material — don't pick those.
FABRIC_SAMPLES_ROOT ?= $(shell \
	_sibling='$(abspath $(ROOT)/../fabric-samples)'; \
	_install='$(FABRIC_INSTALL_WORKDIR)/fabric-samples'; \
	_msp='test-network/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp'; \
	if test -f "$$_sibling/test-network/network.sh" && test -d "$$_sibling/$$_msp"; then echo "$$_sibling"; \
	elif test -f "$$_install/test-network/network.sh" && test -d "$$_install/$$_msp"; then echo "$$_install"; \
	elif test -f "$$_install/test-network/network.sh"; then echo "$$_install"; \
	elif test -f "$$_sibling/test-network/network.sh"; then echo "$$_sibling"; \
	else echo "$$_sibling"; fi)

FABRIC_CRYPTO_MOUNT ?= $(FABRIC_SAMPLES_ROOT)/test-network/organizations/peerOrganizations/org1.example.com

API_PORT ?= 8080
WEB_PORT ?= 8081
FABRIC_HOST_PEER_PORT ?= 7051

export FABRIC_CRYPTO_MOUNT
export API_PORT
export WEB_PORT
export FABRIC_HOST_PEER_PORT

# --- Download Fabric samples, Docker images, and binaries (install-fabric.sh) ---
# Docs: https://hyperledger-fabric.readthedocs.io/en/latest/install.html
INSTALL_FABRIC_SCRIPT_URL ?= https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh
FABRIC_VERSION ?= 2.5.15
FABRIC_CA_VERSION ?= 1.5.17
# Extra args for install-fabric.sh: docker binary samples (short: d b s). Leave empty to install all (script default).
FABRIC_INSTALL_COMPONENTS ?=

help:
	@echo "HashedRoute — Docker + Fabric test-network on the host"
	@echo ""
	@echo "Fabric prerequisites (official install-fabric.sh, as in the Fabric docs):"
	@echo "  make fabric-install-hyperledger   mkdir workdir, curl script, install -f $(FABRIC_VERSION) -c $(FABRIC_CA_VERSION)"
	@echo "    override: FABRIC_GITHUB_USER, FABRIC_INSTALL_WORKDIR, FABRIC_INSTALL_COMPONENTS=(docker binary samples)"
	@echo ""
	@echo "Fabric network + chaincode (README §1–2):"
	@echo "  make fabric-setup   ./network.sh up createChannel + deploy delivery chaincode"
	@echo "  make fabric-network       only bring up test-network + channel"
	@echo "  make fabric-deploy-chaincode   only deploy (network must be up)"
	@echo ""
	@echo "App (after fabric-setup):"
	@echo "  make up          Build images and start api + web (Docker Compose)"
	@echo "  make down        Stop and remove containers"
	@echo ""
	@echo "Paths (override if needed):"
	@echo "  FABRIC_SAMPLES_ROOT=$(FABRIC_SAMPLES_ROOT)"
	@echo "  FABRIC_TEST_NETWORK=$(FABRIC_TEST_NETWORK) (implied from FABRIC_SAMPLES_ROOT)"
	@echo ""
	@echo "URLs after make up:"
	@echo "  UI   → http://localhost:$(WEB_PORT)"
	@echo "  API  → http://localhost:$(API_PORT)  (also proxied as /api on the UI port)"
	@echo ""
	@echo "Local (no Docker) still works: see README.md"

_up-precheck:
	@test -d "$(FABRIC_CRYPTO_MOUNT)/users/User1@org1.example.com/msp" || ( \
		echo "Missing Org1 MSP at: $(FABRIC_CRYPTO_MOUNT)"; \
		echo "Start the test network and enroll users, or set FABRIC_SAMPLES_ROOT / FABRIC_CRYPTO_MOUNT."; \
		exit 1; )

up: _up-precheck
	docker compose -f "$(ROOT)/docker-compose.yml" up --build -d

down:
	docker compose -f "$(ROOT)/docker-compose.yml" down

dev-api:
	cd "$(ROOT)/backend" && FABRIC_CRYPTO_PATH="$(FABRIC_CRYPTO_MOUNT)" go run .

dev-web:
	cd "$(ROOT)/frontend" && npm install && npm run dev

# --- Fabric test-network + chaincode (matches README prerequisites layout) ---

FABRIC_TEST_NETWORK := $(FABRIC_SAMPLES_ROOT)/test-network

# Download https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh
# into FABRIC_INSTALL_WORKDIR and run: ./install-fabric.sh -f FABRIC_VERSION -c FABRIC_CA_VERSION [components...]
fabric-install-hyperledger:
	@mkdir -p "$(FABRIC_INSTALL_WORKDIR)"
	@echo "Working directory: $(FABRIC_INSTALL_WORKDIR)"
	cd "$(FABRIC_INSTALL_WORKDIR)" && curl -sSL -O "$(INSTALL_FABRIC_SCRIPT_URL)"
	cd "$(FABRIC_INSTALL_WORKDIR)" && chmod +x install-fabric.sh
	cd "$(FABRIC_INSTALL_WORKDIR)" && ./install-fabric.sh -f "$(FABRIC_VERSION)" -c "$(FABRIC_CA_VERSION)" $(FABRIC_INSTALL_COMPONENTS)
	@echo ""
	@echo "fabric-samples is at: $(FABRIC_INSTALL_WORKDIR)/fabric-samples"
	@echo "If you set FABRIC_GITHUB_USER on this install, set the same in HashedRoute/.env (or pass it on later make commands) so paths stay in sync:"
	@echo "  FABRIC_GITHUB_USER=$(FABRIC_GITHUB_USER)"
	@echo "  FABRIC_SAMPLES_ROOT=$(FABRIC_INSTALL_WORKDIR)/fabric-samples"

fabric-samples-precheck:
	@test -f "$(FABRIC_TEST_NETWORK)/network.sh" || ( \
		echo "Expected fabric-samples test-network at: $(FABRIC_TEST_NETWORK)"; \
		echo "Set FABRIC_SAMPLES_ROOT in .env to the directory that contains test-network/, e.g."; \
		echo "  FABRIC_SAMPLES_ROOT=$(FABRIC_INSTALL_WORKDIR)/fabric-samples"; \
		echo "If you installed with: make fabric-install-hyperledger FABRIC_GITHUB_USER=X, run:"; \
		echo "  make fabric-setup FABRIC_GITHUB_USER=X"; \
		echo "or add FABRIC_GITHUB_USER=X (and FABRIC_SAMPLES_ROOT=...) to .env"; \
		echo "Or clone: https://github.com/hyperledger/fabric-samples"; \
		exit 1; )

# Start peers/orderer/CA and create channel mychannel (Docker on host).
fabric-network: fabric-samples-precheck
	cd "$(FABRIC_TEST_NETWORK)" && ./network.sh up createChannel

# Deploy HashedRoute chaincode to mychannel (requires Docker + network up).
fabric-deploy-chaincode: fabric-samples-precheck
	@test -d "$(ROOT)/chaincode/delivery" || ( echo "Missing $(ROOT)/chaincode/delivery"; exit 1 )
	FABRIC_TEST_NETWORK="$(FABRIC_TEST_NETWORK)" HASHEDRO_HOME="$(ROOT)" bash "$(ROOT)/scripts/deploy-chaincode.sh"

# Full setup from README §1 and §2: network + channel + delivery chaincode.
fabric-setup: fabric-network fabric-deploy-chaincode
	@echo ""
	@echo "fabric-setup complete. Next: make up   (or dev-api / dev-web)"
