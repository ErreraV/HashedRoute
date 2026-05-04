.PHONY: help compose-gen up down \
	fabric-samples-precheck fabric-network fabric-network-up fabric-add-org3 fabric-clean-ledger fabric-deploy-chaincode fabric-seed-mock fabric-setup \
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

FABRIC_HOSTS_DEF ?= $(ROOT)/fabric-hosts.def

FABRIC_TEST_NETWORK := $(FABRIC_SAMPLES_ROOT)/test-network

export FABRIC_SAMPLES_ROOT
export FABRIC_HOSTS_DEF

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
	@echo "  make fabric-setup   network + createChannel + addOrg3 + deploy delivery chaincode (3-org)"
	@echo "  make fabric-network       ./network.sh up createChannel (first-time channel; see README if this errors)"
	@echo "  make fabric-network-up    ./network.sh up only (nodes on host; channel must already exist)"
	@echo "  make fabric-add-org3     add Org3 to mychannel (after fabric-network; idempotent if already added)"
	@echo "  make fabric-deploy-chaincode   only deploy (network + Org3 must be up; policy OR Org1–Org3 peers)"
	@echo "  make fabric-seed-mock     CLI: add mock shipments to the ledger (network + CC up)"
	@echo "  make fabric-clean-ledger  ./network.sh down + wipe org/channel dirs on disk (then make fabric-setup)"
	@echo ""
	@echo "App (after fabric-setup):"
	@echo "  make compose-gen   Regenerate docker-compose.yml (three api/web stacks; optional port edits in fabric-hosts.def)"
	@echo "  make up          Build and start all API/UI stacks (after fabric-setup)"
	@echo "  make down        Stop and remove containers"
	@echo ""
	@echo "Paths (override if needed):"
	@echo "  FABRIC_SAMPLES_ROOT=$(FABRIC_SAMPLES_ROOT)"
	@echo "  FABRIC_HOSTS_DEF=$(FABRIC_HOSTS_DEF)  (optional path; file must stay three-row org1–org3)"
	@echo "  FABRIC_TEST_NETWORK=$(FABRIC_TEST_NETWORK) (implied from FABRIC_SAMPLES_ROOT)"
	@echo ""
	@echo "UI/API ports: last two columns per row in $(FABRIC_HOSTS_DEF) (three rows only)"
	@echo ""

compose-gen:
	@HASHEDRO_HOME="$(ROOT)" FABRIC_HOSTS_DEF="$(FABRIC_HOSTS_DEF)" bash "$(ROOT)/scripts/gen-docker-compose-hosts.sh"

_up-precheck: compose-gen fabric-samples-precheck
	@HASHEDRO_HOME="$(ROOT)" FABRIC_HOSTS_DEF="$(FABRIC_HOSTS_DEF)" FABRIC_SAMPLES_ROOT="$(FABRIC_SAMPLES_ROOT)" bash "$(ROOT)/scripts/check-fabric-hosts.sh"

up: _up-precheck
	docker compose -f "$(ROOT)/docker-compose.yml" up --build -d

down: compose-gen
	docker compose -f "$(ROOT)/docker-compose.yml" down --remove-orphans

# --- Fabric test-network + chaincode (matches README prerequisites layout) ---

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

# Start peers/orderer only (no createChannel). Use when mychannel already exists and fabric-network would fail on re-join.
fabric-network-up: fabric-samples-precheck
	cd "$(FABRIC_TEST_NETWORK)" && ./network.sh up

# Start peers/orderer/CA and create channel mychannel (Docker on host). Fails if channel already exists and peers already joined — use fabric-network-up + fabric-deploy-chaincode, or fabric-clean-ledger for a full reset.
fabric-network: fabric-samples-precheck
	cd "$(FABRIC_TEST_NETWORK)" && ./network.sh up createChannel

# Join Org3 to mychannel (run after fabric-network). Safe to re-run if Org3 is already on the channel (script may error — use fabric-clean-ledger for a clean slate).
fabric-add-org3: fabric-samples-precheck
	cd "$(FABRIC_TEST_NETWORK)/addOrg3" && ./addOrg3.sh up

# Tear down test-network: removes peer/orderer containers, volumes, and generated org/channel artifacts.
# Extra host-side rm avoids a fabric-samples edge case: stale organizations/ordererOrganizations without
# peerOrganizations, which skips cryptogen and breaks createChannel (missing orderer TLS for configtxgen).
fabric-clean-ledger: fabric-samples-precheck
	cd "$(FABRIC_TEST_NETWORK)" && ./network.sh down
	rm -rf "$(FABRIC_TEST_NETWORK)/organizations/peerOrganizations" \
		"$(FABRIC_TEST_NETWORK)/organizations/ordererOrganizations" \
		"$(FABRIC_TEST_NETWORK)/system-genesis-block" \
		"$(FABRIC_TEST_NETWORK)/channel-artifacts"
	@echo ""
	@echo "Ledger cleared. Recreate: make fabric-setup"
	@echo "  (or: make fabric-network && make fabric-add-org3 && make fabric-deploy-chaincode)"
	@echo "Ignore harmless 'no such volume' lines for docker_peer0.* / docker_orderer.* if they appear (compose uses compose_* volumes)."

# Deploy HashedRoute chaincode to mychannel (Docker + test-network + Org3 material on disk).
# Optional on the make line: SEQ= CC_VER= CC_END_POLICY= CC_NAME=
fabric-deploy-chaincode: fabric-samples-precheck
	@test -d "$(ROOT)/chaincode/delivery" || ( echo "Missing $(ROOT)/chaincode/delivery"; exit 1 )
	FABRIC_TEST_NETWORK="$(FABRIC_TEST_NETWORK)" HASHEDRO_HOME="$(ROOT)" \
		CC_VER="$(CC_VER)" SEQ="$(SEQ)" CC_END_POLICY="$(CC_END_POLICY)" CC_NAME="$(CC_NAME)" \
		bash "$(ROOT)/scripts/deploy-chaincode.sh"

# Submit mock CreateShipment / UpdateStatus txs via peer CLI (Org1 Admin).
fabric-seed-mock: fabric-samples-precheck
	FABRIC_TEST_NETWORK="$(FABRIC_TEST_NETWORK)" HASHEDRO_HOME="$(ROOT)" bash "$(ROOT)/scripts/seed-mock-ledger.sh"

# Full setup: network + channel, Org3 join, delivery chaincode (3-org endorsement).
fabric-setup: fabric-network fabric-add-org3 fabric-deploy-chaincode
	@echo ""
	@echo "fabric-setup complete. Next: make up"
