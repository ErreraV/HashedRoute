package main

import (
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/hash"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// GatewayConfig holds filesystem paths and Fabric connection targets.
type GatewayConfig struct {
	CertPath     string
	KeyDir       string
	TLSCertPath  string
	PeerEndpoint string
	GatewayPeer  string
	MSPID        string
	Channel      string
	Chaincode    string
	Contract     string
}

func gatewayConfigFromEnv() (GatewayConfig, error) {
	wd, err := os.Getwd()
	if err != nil {
		return GatewayConfig{}, err
	}
	// From repo: HashedRoute/backend as cwd → ../../fabric-samples/...
	defaultCrypto := filepath.Join(wd, "..", "..", "fabric-samples", "test-network", "organizations", "peerOrganizations", "org1.example.com")

	cryptoRoot := strings.TrimSpace(os.Getenv("FABRIC_CRYPTO_PATH"))
	if cryptoRoot == "" {
		cryptoRoot = defaultCrypto
	}
	cryptoRoot = filepath.Clean(cryptoRoot)

	keyDir := strings.TrimSpace(os.Getenv("FABRIC_KEY_DIR"))
	if keyDir == "" {
		keyDir = filepath.Join(cryptoRoot, "users", "User1@org1.example.com", "msp", "keystore")
	} else {
		keyDir = filepath.Clean(keyDir)
	}

	certPath := strings.TrimSpace(os.Getenv("FABRIC_CERT_PATH"))
	if certPath == "" {
		signcerts := filepath.Join(filepath.Dir(keyDir), "signcerts")
		var err error
		certPath, err = resolveSignCert(signcerts)
		if err != nil {
			return GatewayConfig{}, err
		}
	}
	tlsCA := os.Getenv("FABRIC_TLS_CA")
	if tlsCA == "" {
		tlsCA = filepath.Join(cryptoRoot, "peers", "peer0.org1.example.com", "tls", "ca.crt")
	}

	return GatewayConfig{
		CertPath:     certPath,
		KeyDir:       keyDir,
		TLSCertPath:  tlsCA,
		PeerEndpoint: getEnvDefault("FABRIC_PEER_ENDPOINT", "dns:///localhost:7051"),
		GatewayPeer:  getEnvDefault("FABRIC_GATEWAY_PEER", "peer0.org1.example.com"),
		MSPID:        getEnvDefault("FABRIC_MSP_ID", "Org1MSP"),
		Channel:      getEnvDefault("FABRIC_CHANNEL", "mychannel"),
		Chaincode:    getEnvDefault("FABRIC_CHAINCODE", "delivery"),
		Contract:     getEnvDefault("FABRIC_CONTRACT_NAME", "DeliveryContract"),
	}, nil
}

func getEnvDefault(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

// resolveSignCert picks an identity cert under MSP signcerts (cryptogen uses *-cert.pem; CA enroll may use cert.pem).
func resolveSignCert(signcertsDir string) (string, error) {
	def := filepath.Join(signcertsDir, "cert.pem")
	if _, err := os.Stat(def); err == nil {
		return def, nil
	}
	entries, err := os.ReadDir(signcertsDir)
	if err != nil {
		return "", fmt.Errorf("msp signcerts %s: %w", signcertsDir, err)
	}
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if strings.EqualFold(filepath.Ext(e.Name()), ".pem") {
			return filepath.Join(signcertsDir, e.Name()), nil
		}
	}
	return "", fmt.Errorf("no .pem certificate in %s (expected cert.pem or *-cert.pem)", signcertsDir)
}

func connectGateway(cfg GatewayConfig) (*client.Gateway, *grpc.ClientConn, error) {
	conn, err := newGrpcConnection(cfg)
	if err != nil {
		return nil, nil, err
	}
	id, err := newIdentity(cfg)
	if err != nil {
		conn.Close()
		return nil, nil, err
	}
	sign, err := newSign(cfg)
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithHash(hash.SHA256),
		client.WithClientConnection(conn),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(time.Minute),
	)
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("gateway connect: %w", err)
	}
	return gw, conn, nil
}

func newGrpcConnection(cfg GatewayConfig) (*grpc.ClientConn, error) {
	cert, err := loadCertificate(cfg.TLSCertPath)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AddCert(cert)
	creds := credentials.NewClientTLSFromCert(pool, cfg.GatewayPeer)
	return grpc.NewClient(cfg.PeerEndpoint, grpc.WithTransportCredentials(creds))
}

func newIdentity(cfg GatewayConfig) (*identity.X509Identity, error) {
	cert, err := loadCertificate(cfg.CertPath)
	if err != nil {
		return nil, err
	}
	return identity.NewX509Identity(cfg.MSPID, cert)
}

func newSign(cfg GatewayConfig) (identity.Sign, error) {
	keyFile := filepath.Join(cfg.KeyDir, "priv_sk")
	if _, err := os.Stat(keyFile); err == nil {
		return signFromPEMFile(keyFile)
	}
	files, err := os.ReadDir(cfg.KeyDir)
	if err != nil {
		return nil, fmt.Errorf("read keystore: %w", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no private key file in %s", cfg.KeyDir)
	}
	return signFromPEMFile(filepath.Join(cfg.KeyDir, files[0].Name()))
}

func signFromPEMFile(file string) (identity.Sign, error) {
	pemBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	pk, err := identity.PrivateKeyFromPEM(pemBytes)
	if err != nil {
		return nil, err
	}
	return identity.NewPrivateKeySign(pk)
}

func loadCertificate(file string) (*x509.Certificate, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read cert %s: %w", file, err)
	}
	return identity.CertificateFromPEM(b)
}

func fabricContract(cfg GatewayConfig, gw *client.Gateway) *client.Contract {
	net := gw.GetNetwork(cfg.Channel)
	return net.GetContractWithName(cfg.Chaincode, cfg.Contract)
}
