package op_e2e

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	prng "math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-node/p2p/store"
	"github.com/ethereum-optimism/optimism/op-service/clock"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/sync"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoremem"

	ic "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	geth_eth "github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/require"

	bss "github.com/ethereum-optimism/optimism/op-batcher/batcher"
	"github.com/ethereum-optimism/optimism/op-batcher/compressor"
	batchermetrics "github.com/ethereum-optimism/optimism/op-batcher/metrics"
	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-e2e/config"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	"github.com/ethereum-optimism/optimism/op-node/chaincfg"
	"github.com/ethereum-optimism/optimism/op-node/metrics"
	rollupNode "github.com/ethereum-optimism/optimism/op-node/node"
	"github.com/ethereum-optimism/optimism/op-node/p2p"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/driver"
	"github.com/ethereum-optimism/optimism/op-node/sources"
	"github.com/ethereum-optimism/optimism/op-node/testlog"
	proposermetrics "github.com/ethereum-optimism/optimism/op-proposer/metrics"
	l2os "github.com/ethereum-optimism/optimism/op-proposer/proposer"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

var (
	testingJWTSecret = [32]byte{123}
)

func newTxMgrConfig(l1Addr string, privKey *ecdsa.PrivateKey) txmgr.CLIConfig {
	return txmgr.CLIConfig{
		L1RPCURL:                  l1Addr,
		PrivateKey:                hexPriv(privKey),
		NumConfirmations:          1,
		SafeAbortNonceTooLowCount: 3,
		ResubmissionTimeout:       3 * time.Second,
		ReceiptQueryInterval:      50 * time.Millisecond,
		NetworkTimeout:            2 * time.Second,
		TxNotInMempoolTimeout:     2 * time.Minute,
	}
}

func DefaultSystemConfig(t *testing.T) SystemConfig {
	secrets, err := e2eutils.DefaultMnemonicConfig.Secrets()
	require.NoError(t, err)
	deployConfig := config.DeployConfig.Copy()
	deployConfig.L1GenesisBlockTimestamp = hexutil.Uint64(time.Now().Unix())
	require.NoError(t, deployConfig.Check())
	l1Deployments := config.L1Deployments.Copy()
	require.NoError(t, l1Deployments.Check())

	require.Equal(t, secrets.Addresses().Batcher, deployConfig.BatchSenderAddress)
	require.Equal(t, secrets.Addresses().SequencerP2P, deployConfig.P2PSequencerAddress)
	require.Equal(t, secrets.Addresses().Proposer, deployConfig.L2OutputOracleProposer)

	// Tests depend on premine being filled with secrets addresses
	premine := make(map[common.Address]*big.Int)
	for _, addr := range secrets.Addresses().All() {
		premine[addr] = new(big.Int).Mul(big.NewInt(1000), big.NewInt(params.Ether))
	}

	return SystemConfig{
		Secrets:                secrets,
		Premine:                premine,
		DeployConfig:           deployConfig,
		L1Deployments:          config.L1Deployments,
		L1InfoPredeployAddress: predeploys.L1BlockAddr,
		JWTFilePath:            writeDefaultJWT(t),
		JWTSecret:              testingJWTSecret,
		Nodes: map[string]*rollupNode.Config{
			"sequencer": {
				Driver: driver.Config{
					VerifierConfDepth:  0,
					SequencerConfDepth: 0,
					SequencerEnabled:   true,
				},
				// Submitter PrivKey is set in system start for rollup nodes where sequencer = true
				RPC: rollupNode.RPCConfig{
					ListenAddr:  "0.0.0.0",
					ListenPort:  0,
					EnableAdmin: true,
				},
				L1EpochPollInterval: time.Second * 2,
				ConfigPersistence:   &rollupNode.DisabledConfigPersistence{},
			},
			"verifier": {
				Driver: driver.Config{
					VerifierConfDepth:  0,
					SequencerConfDepth: 0,
					SequencerEnabled:   false,
				},
				L1EpochPollInterval: time.Second * 4,
				ConfigPersistence:   &rollupNode.DisabledConfigPersistence{},
			},
		},
		Loggers: map[string]log.Logger{
			"verifier":  testlog.Logger(t, log.LvlInfo).New("role", "verifier"),
			"sequencer": testlog.Logger(t, log.LvlInfo).New("role", "sequencer"),
			"batcher":   testlog.Logger(t, log.LvlInfo).New("role", "batcher"),
			"proposer":  testlog.Logger(t, log.LvlCrit).New("role", "proposer"),
		},
		GethOptions:                map[string][]GethOption{},
		P2PTopology:                nil, // no P2P connectivity by default
		NonFinalizedProposals:      false,
		BatcherTargetL1TxSizeBytes: 100_000,
	}
}

func writeDefaultJWT(t *testing.T) string {
	// Sadly the geth node config cannot load JWT secret from memory, it has to be a file
	jwtPath := path.Join(t.TempDir(), "jwt_secret")
	if err := os.WriteFile(jwtPath, []byte(hexutil.Encode(testingJWTSecret[:])), 0600); err != nil {
		t.Fatalf("failed to prepare jwt file for geth: %v", err)
	}
	return jwtPath
}

type DepositContractConfig struct {
	L2Oracle           common.Address
	FinalizationPeriod *big.Int
}

type SystemConfig struct {
	Secrets                *e2eutils.Secrets
	L1InfoPredeployAddress common.Address

	DeployConfig  *genesis.DeployConfig
	L1Deployments *genesis.L1Deployments

	JWTFilePath string
	JWTSecret   [32]byte

	Premine        map[common.Address]*big.Int
	Nodes          map[string]*rollupNode.Config // Per node config. Don't use populate rollup.Config
	Loggers        map[string]log.Logger
	GethOptions    map[string][]GethOption
	ProposerLogger log.Logger
	BatcherLogger  log.Logger

	// map of outbound connections to other nodes. Node names prefixed with "~" are unconnected but linked.
	// A nil map disables P2P completely.
	// Any node name not in the topology will not have p2p enabled.
	P2PTopology map[string][]string

	// Enables req-resp sync in the P2P nodes
	P2PReqRespSync bool

	// If the proposer can make proposals for L2 blocks derived from L1 blocks which are not finalized on L1 yet.
	NonFinalizedProposals bool

	// Explicitly disable batcher, for tests that rely on unsafe L2 payloads
	DisableBatcher bool

	// Target L1 tx size for the batcher transactions
	BatcherTargetL1TxSizeBytes uint64

	// SupportL1TimeTravel determines if the L1 node supports quickly skipping forward in time
	SupportL1TimeTravel bool
}

type System struct {
	cfg SystemConfig

	RollupConfig *rollup.Config

	L2GenesisCfg *core.Genesis

	Espresso *EspressoSystem

	// Connections to running nodes
	Nodes             map[string]*node.Node
	Backends          map[string]*geth_eth.Ethereum
	Clients           map[string]*ethclient.Client
	RollupNodes       map[string]*rollupNode.OpNode
	L2OutputSubmitter *l2os.L2OutputSubmitter
	BatchSubmitter    *bss.BatchSubmitter
	Mocknet           mocknet.Mocknet

	// TimeTravelClock is nil unless SystemConfig.SupportL1TimeTravel was set to true
	// It provides access to the clock instance used by the L1 node. Calling TimeTravelClock.AdvanceBy
	// allows tests to quickly time travel L1 into the future.
	// Note that this time travel may occur in a single block, creating a very large difference in the Time
	// on sequential blocks.
	TimeTravelClock *clock.AdvancingClock
}

func (sys *System) NodeEndpoint(name string) string {
	return selectEndpoint(sys.Nodes[name])
}

func (sys *System) Close() {
	if sys.L2OutputSubmitter != nil {
		sys.L2OutputSubmitter.Stop()
	}
	if sys.BatchSubmitter != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		sys.BatchSubmitter.StopIfRunning(ctx)
	}

	for _, node := range sys.RollupNodes {
		node.Close()
	}
	if sys.Espresso != nil {
		sys.Espresso.Close()
	}
	for _, node := range sys.Nodes {
		node.Close()
	}
	sys.Mocknet.Close()
}

type EspressoSystem struct {
	composeFile   string
	projectName   string
	sequencerPort uint16
	proxyPort     uint16
	logsProcess   *exec.Cmd
}

func (e *EspressoSystem) SequencerUrl() string {
	return fmt.Sprintf("http://localhost:%d", e.sequencerPort)
}

func (e *EspressoSystem) ProxyUrl() string {
	return fmt.Sprintf("http://localhost:%d", e.proxyPort)
}

func (e *EspressoSystem) WaitForBlockHeight(ctx context.Context, height uint64) error {
	url := e.SequencerUrl() + "/status/latest_block_height"
	for {
		res, err := http.Get(url)
		if err == nil {
			defer res.Body.Close()
		}

		if err == nil && res.StatusCode == 200 {
			var currentHeight uint64
			if err := json.NewDecoder(res.Body).Decode(&currentHeight); err != nil {
				return err
			}
			if currentHeight >= height {
				return nil
			}
			log.Info("waiting for Espresso block height", "current", currentHeight, "desired", height)
		} else {
			log.Warn("failed to get latest Espresso block height", "res", res, "err", err)
		}

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (e *EspressoSystem) StartGethProxy(sequencer *node.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx,
		"docker", "compose", "--project-name", e.projectName, "-f", e.composeFile,
		"up", "op-geth-proxy", "-V", "--force-recreate", "--wait")
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	cmd.Stdout = &stderr
	// Point the L2 RPC proxy at the OP sequencer Geth node.
	cmd.Env = append(cmd.Env, fmt.Sprintf("OP_GETH_PROXY_GETH_ADDR=%s", httpEndpointForDocker(sequencer)))
	// Enable INFO level logging for Rust services.
	cmd.Env = append(cmd.Env, "RUST_LOG=info")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker compose up (%v) error: %w output: %s", cmd, err, stderr.String())
	}

	// Find the ports which were randomly assigned to the services.
	proxyPort, err := dockerComposePort(e.projectName, e.composeFile, "op-geth-proxy", 9090)
	if err != nil {
		return err
	}
	e.proxyPort = proxyPort
	return nil
}

func (e *EspressoSystem) PrintLogs() {
	logs := exec.Command("docker", "compose", "--project-name", e.projectName, "-f", e.composeFile, "logs")
	logs.Stdout = os.Stdout
	logs.Stderr = os.Stderr
	if err := logs.Run(); err != nil {
		log.Error("failed to get docker-compose logs: %w", err)
	}
}

func (e *EspressoSystem) AttachLogs() error {
	// Forward service logs to our stdout.
	logs := exec.Command("docker", "compose", "--project-name", e.projectName, "-f", e.composeFile, "logs", "-f")
	logs.Stdout = os.Stdout
	logs.Stderr = os.Stderr
	if err := logs.Start(); err != nil {
		return fmt.Errorf("failed to attach to docker compose logs (%v): %w", logs, err)
	}
	e.logsProcess = logs
	return nil
}

func (e *EspressoSystem) Close() {
	// Kill the docker-compose environment.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "compose", "--project-name", e.projectName,
		"-f", e.composeFile, "down", "-v")
	if err := cmd.Run(); err != nil {
		log.Error("failed to kill docker-compose", "err", err)
	}

	// Kill the logs process.
	if e.logsProcess != nil {
		if err := e.logsProcess.Process.Kill(); err != nil {
			log.Error("failed to kill docker-compose logs", "err", err)
		}
		if err := e.logsProcess.Wait(); err != nil {
			log.Error("failed to wait for docker-compose logs", "err", err)
		}
	}
}

func dockerComposePort(projectName string, composeFile string, service string, internalPort uint16) (uint16, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx,
		"docker", "compose", "--project-name", projectName, "-f", composeFile, "port",
		service, strconv.Itoa(int(internalPort)))
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("docker compose port failed: %w, %v", err, cmd)
	}
	port, err := strconv.Atoi(strings.TrimSpace(strings.Split(string(output), ":")[1]))
	if err != nil {
		return 0, err
	}
	return uint16(port), nil
}

type systemConfigHook func(sCfg *SystemConfig, s *System)

type SystemConfigOption struct {
	key    string
	role   string
	action systemConfigHook
}

type SystemConfigOptions struct {
	opts map[string]systemConfigHook
}

func NewSystemConfigOptions(_opts []SystemConfigOption) (SystemConfigOptions, error) {
	opts := make(map[string]systemConfigHook)
	for _, opt := range _opts {
		if _, ok := opts[opt.key+":"+opt.role]; ok {
			return SystemConfigOptions{}, fmt.Errorf("duplicate option for key %s and role %s", opt.key, opt.role)
		}
		opts[opt.key+":"+opt.role] = opt.action
	}

	return SystemConfigOptions{
		opts: opts,
	}, nil
}

func (s *SystemConfigOptions) Get(key, role string) (systemConfigHook, bool) {
	v, ok := s.opts[key+":"+role]
	return v, ok
}

func (cfg SystemConfig) Start(_opts ...SystemConfigOption) (*System, error) {
	opts, err := NewSystemConfigOptions(_opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create system config: %w", err)
	}

	sys := &System{
		cfg:         cfg,
		Nodes:       make(map[string]*node.Node),
		Backends:    make(map[string]*geth_eth.Ethereum),
		Clients:     make(map[string]*ethclient.Client),
		RollupNodes: make(map[string]*rollupNode.OpNode),
	}
	didErrAfterStart := false
	defer func() {
		if didErrAfterStart {
			for _, node := range sys.RollupNodes {
				node.Close()
			}
			for _, node := range sys.Nodes {
				node.Close()
			}
		}
	}()

	c := clock.SystemClock
	if cfg.SupportL1TimeTravel {
		sys.TimeTravelClock = clock.NewAdvancingClock(100 * time.Millisecond)
		c = sys.TimeTravelClock
	}

	if err := cfg.DeployConfig.Check(); err != nil {
		return nil, fmt.Errorf("invalid DeployConfig: %w", err)
	}

	l1Genesis, err := genesis.BuildL1DeveloperGenesis(cfg.DeployConfig, config.L1Allocs, config.L1Deployments, true)
	if err != nil {
		return nil, fmt.Errorf("failed to build L1 genesis: %w", err)
	}

	for addr, amount := range cfg.Premine {
		if existing, ok := l1Genesis.Alloc[addr]; ok {
			l1Genesis.Alloc[addr] = core.GenesisAccount{
				Code:    existing.Code,
				Storage: existing.Storage,
				Balance: amount,
				Nonce:   existing.Nonce,
			}
		} else {
			l1Genesis.Alloc[addr] = core.GenesisAccount{
				Balance: amount,
				Nonce:   0,
			}
		}
	}

	// Initialize L1
	l1Node, l1Backend, err := initL1Geth(&cfg, l1Genesis, c, cfg.GethOptions["l1"]...)
	if err != nil {
		return nil, fmt.Errorf("failed to init L1 geth: %w", err)
	}
	sys.Nodes["l1"] = l1Node
	sys.Backends["l1"] = l1Backend
	err = l1Node.Start()
	if err != nil {
		didErrAfterStart = true
		return nil, fmt.Errorf("failed to starat L1 geth: %w", err)
	}

	// Connect to L1 Geth client
	l1Srv, err := l1Node.RPCHandler()
	if err != nil {
		didErrAfterStart = true
		return nil, fmt.Errorf("failed to connect to L1 geth: %w", err)
	}
	l1Client := ethclient.NewClient(rpc.DialInProc(l1Srv))
	sys.Clients["l1"] = l1Client

	// Start an Espresso sequencer network, if required.
	if cfg.DeployConfig.Espresso {
		// Find the docker-compose file.
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get cwd: %w", err)
		}
		root, err := config.FindMonorepoRoot(cwd)
		if err != nil {
			return nil, fmt.Errorf("failed to find monorepo root: %w", err)
		}
		composeFile := filepath.Join(root, "ops-bedrock", "docker-compose.yml")

		// Generate a random project name to distinguish this docker-compose network from that of
		// other tests running in parallel.
		projectName := fmt.Sprintf("e2e-tests-%d", prng.Int63())

		// Start the services.
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx,
			"docker", "compose", "--project-name", projectName, "-f", composeFile,
			"up", "orchestrator", "da-server", "consensus-server", "sequencer0", "sequencer1", "commitment-task",
			"-V", "--force-recreate", "--wait")
		stderr := bytes.Buffer{}
		cmd.Stderr = &stderr
		cmd.Stdout = &stderr
		// Point the sequencer at the L1 Geth node.
		cmd.Env = append(cmd.Env, fmt.Sprintf("ESPRESSO_SEQUENCER_L1_PROVIDER=%s", httpEndpointForDocker(l1Node)))
		// Make the Espresso block time faster than the OP block time, or else tests will time out.
		cmd.Env = append(cmd.Env, fmt.Sprintf("ESPRESSO_ORCHESTRATOR_MAX_PROPOSE_TIME=%dms", cfg.DeployConfig.L2BlockTime*1000/2))
		cmd.Env = append(cmd.Env, "RUST_LOG=info")
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("docker compose up (%v) error: %w output: %s", cmd, err, stderr.String())
		}

		// Find the ports which were randomly assigned to the services.
		sequencerPort, err := dockerComposePort(projectName, composeFile, "sequencer0", 8080)
		if err != nil {
			return nil, fmt.Errorf("failed to get sequencer0 port: %w", err)
		}

		sys.Espresso = &EspressoSystem{
			projectName:   projectName,
			composeFile:   composeFile,
			sequencerPort: sequencerPort,
		}

		// Wait for Espresso to start producing blocks. Because of pipelining, the first block can
		// take a few seconds, which we don't want to count against the test timeout.
		if err := sys.Espresso.WaitForBlockHeight(ctx, 1); err != nil {
			// If we never reached a height of a single block, something is probably wrong with the
			// Espresso configuration, and we will want to look at the logs.
			sys.Espresso.PrintLogs()
			return nil, fmt.Errorf("failed to reach Espresso block height: %w", err)
		}
	}

	// Initialize L2
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	l1Block, err := l1Client.BlockByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get L1 head: %w", err)
	}
	l2Genesis, err := genesis.BuildL2Genesis(cfg.DeployConfig, l1Block)
	if err != nil {
		return nil, fmt.Errorf("failed to build L2 genesis: %w", err)
	}
	sys.L2GenesisCfg = l2Genesis
	for addr, amount := range cfg.Premine {
		if existing, ok := l2Genesis.Alloc[addr]; ok {
			l2Genesis.Alloc[addr] = core.GenesisAccount{
				Code:    existing.Code,
				Storage: existing.Storage,
				Balance: amount,
				Nonce:   existing.Nonce,
			}
		} else {
			l2Genesis.Alloc[addr] = core.GenesisAccount{
				Balance: amount,
				Nonce:   0,
			}
		}
	}

	makeRollupConfig := func() rollup.Config {
		return rollup.Config{
			Genesis: rollup.Genesis{
				L1: eth.BlockID{
					Hash:   l1Block.Hash(),
					Number: l1Block.Number().Uint64(),
				},
				L2: eth.BlockID{
					Hash:   l2Genesis.ToBlock().Hash(),
					Number: 0,
				},
				L2Time:       l2Genesis.Timestamp,
				SystemConfig: e2eutils.SystemConfigFromDeployConfig(cfg.DeployConfig),
			},
			BlockTime:              cfg.DeployConfig.L2BlockTime,
			MaxSequencerDrift:      cfg.DeployConfig.MaxSequencerDrift,
			SeqWindowSize:          cfg.DeployConfig.SequencerWindowSize,
			ChannelTimeout:         cfg.DeployConfig.ChannelTimeout,
			L1ChainID:              cfg.L1ChainIDBig(),
			L2ChainID:              cfg.L2ChainIDBig(),
			BatchInboxAddress:      cfg.DeployConfig.BatchInboxAddress,
			HotShotContractAddress: cfg.DeployConfig.HotShotContractAddress,
			DepositContractAddress: cfg.DeployConfig.OptimismPortalProxy,
			L1SystemConfigAddress:  cfg.DeployConfig.SystemConfigProxy,
			RegolithTime:           cfg.DeployConfig.RegolithTime(uint64(cfg.DeployConfig.L1GenesisBlockTimestamp)),
		}
	}
	defaultConfig := makeRollupConfig()
	if err := defaultConfig.Check(); err != nil {
		return nil, fmt.Errorf("rollup config is invalid: %w", err)
	}
	sys.RollupConfig = &defaultConfig

	// Init nodes
	for name := range cfg.Nodes {
		node, backend, err := initL2Geth(name, big.NewInt(int64(cfg.DeployConfig.L2ChainID)), l2Genesis, cfg.JWTFilePath, cfg.GethOptions[name]...)
		if err != nil {
			return nil, fmt.Errorf("failed to init L2 Geth %s: %w", name, err)
		}
		sys.Nodes[name] = node
		sys.Backends[name] = backend
	}

	// Start
	for name, node := range sys.Nodes {
		if name == "l1" {
			continue
		}
		err = node.Start()
		if err != nil {
			didErrAfterStart = true
			return nil, fmt.Errorf("failed to start L2 Geth: %s: %w", name, err)
		}
	}

	// Now that the L2 nodes are running, start an Espresso proxy for the L2 sequencer Geth node.
	if sys.Espresso != nil {
		if err := sys.Espresso.StartGethProxy(sys.Nodes["sequencer"]); err != nil {
			return nil, fmt.Errorf("failed to start Geth proxy: %w", err)
		}
		// Now that all the Docker services are running, attach to logs.
		if err := sys.Espresso.AttachLogs(); err != nil {
			return nil, fmt.Errorf("failed to attach to Docker logs: %w", err)
		}
	}

	// Configure connections to L1 and L2 for rollup nodes.
	// TODO: refactor testing to use in-process rpc connections instead of websockets.

	for name, rollupCfg := range cfg.Nodes {
		configureL1(rollupCfg, l1Node)
		configureL2(rollupCfg, sys.Nodes[name], sys.Espresso, cfg.JWTSecret)

		rollupCfg.L2Sync = &rollupNode.PreparedL2SyncEndpoint{
			Client:   nil,
			TrustRPC: false,
		}
	}

	// Geth Clients
	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for name, node := range sys.Nodes {
		var endpoint string
		if sys.Espresso != nil && name == "sequencer" {
			// In Espresso mode, clients should talk to the sequencer through the proxy RPC, which
			// forwards submitted transactions to the Espresso sequencer.
			endpoint = sys.Espresso.ProxyUrl()
		} else {
			endpoint = node.WSEndpoint()
		}

		client, err := ethclient.DialContext(ctx, endpoint)
		if err != nil {
			didErrAfterStart = true
			return nil, fmt.Errorf("failed to dial eth client %s: %w", endpoint, err)
		}
		sys.Clients[name] = client
	}

	_, err = waitForBlock(big.NewInt(2), l1Client, 6*time.Second*time.Duration(cfg.DeployConfig.L1BlockTime))
	if err != nil {
		return nil, fmt.Errorf("waiting for blocks: %w", err)
	}

	sys.Mocknet = mocknet.New()

	p2pNodes := make(map[string]*p2p.Prepared)
	if cfg.P2PTopology != nil {
		// create the peer if it doesn't exist yet.
		initHostMaybe := func(name string) (*p2p.Prepared, error) {
			if p, ok := p2pNodes[name]; ok {
				return p, nil
			}
			h, err := sys.newMockNetPeer()
			if err != nil {
				return nil, fmt.Errorf("failed to init p2p host for node %s", name)
			}
			h.Network()
			_, ok := cfg.Nodes[name]
			if !ok {
				return nil, fmt.Errorf("node %s from p2p topology not found in actual nodes map", name)
			}
			// TODO we can enable discv5 in the testnodes to test discovery of new peers.
			// Would need to mock though, and the discv5 implementation does not provide nice mocks here.
			p := &p2p.Prepared{
				HostP2P:           h,
				LocalNode:         nil,
				UDPv5:             nil,
				EnableReqRespSync: cfg.P2PReqRespSync,
			}
			p2pNodes[name] = p
			return p, nil
		}
		for k, vs := range cfg.P2PTopology {
			peerA, err := initHostMaybe(k)
			if err != nil {
				return nil, fmt.Errorf("failed to setup mocknet peer %s", k)
			}
			for _, v := range vs {
				v = strings.TrimPrefix(v, "~")
				peerB, err := initHostMaybe(v)
				if err != nil {
					return nil, fmt.Errorf("failed to setup mocknet peer %s (peer of %s)", v, k)
				}
				if _, err := sys.Mocknet.LinkPeers(peerA.HostP2P.ID(), peerB.HostP2P.ID()); err != nil {
					return nil, fmt.Errorf("failed to setup mocknet link between %s and %s", k, v)
				}
				// connect the peers after starting the full rollup node
			}
		}
	}

	// Don't log state snapshots in test output
	snapLog := log.New()
	snapLog.SetHandler(log.DiscardHandler())

	// Rollup nodes

	// Ensure we are looping through the nodes in alphabetical order
	ks := make([]string, 0, len(cfg.Nodes))
	for k := range cfg.Nodes {
		ks = append(ks, k)
	}
	// Sort strings in ascending alphabetical order
	sort.Strings(ks)

	for _, name := range ks {
		nodeConfig := cfg.Nodes[name]
		c := *nodeConfig // copy
		c.Rollup = makeRollupConfig()
		if err := c.LoadPersisted(cfg.Loggers[name]); err != nil {
			return nil, fmt.Errorf("failed to load persisted logger: %w", err)
		}

		if p, ok := p2pNodes[name]; ok {
			c.P2P = p

			if c.Driver.SequencerEnabled && c.P2PSigner == nil {
				c.P2PSigner = &p2p.PreparedSigner{Signer: p2p.NewLocalSigner(cfg.Secrets.SequencerP2P)}
			}
		}

		c.Rollup.LogDescription(cfg.Loggers[name], chaincfg.L2ChainIDToNetworkName)

		node, err := rollupNode.New(context.Background(), &c, cfg.Loggers[name], snapLog, "", metrics.NewMetrics(""))
		if err != nil {
			didErrAfterStart = true
			return nil, fmt.Errorf("failed to create rollup node %s: %w", name, err)
		}
		err = node.Start(context.Background())
		if err != nil {
			didErrAfterStart = true
			return nil, fmt.Errorf("failed to start rollup node %s: %w", name, err)
		}
		sys.RollupNodes[name] = node

		if action, ok := opts.Get("afterRollupNodeStart", name); ok {
			action(&cfg, sys)
		}
	}

	if cfg.P2PTopology != nil {
		// We only set up the connections after starting the actual nodes,
		// so GossipSub and other p2p protocols can be started before the connections go live.
		// This way protocol negotiation happens correctly.
		for k, vs := range cfg.P2PTopology {
			peerA := p2pNodes[k]
			for _, v := range vs {
				unconnected := strings.HasPrefix(v, "~")
				if unconnected {
					v = v[1:]
				}
				if !unconnected {
					peerB := p2pNodes[v]
					if _, err := sys.Mocknet.ConnectPeers(peerA.HostP2P.ID(), peerB.HostP2P.ID()); err != nil {
						return nil, fmt.Errorf("failed to setup mocknet connection between %s and %s", k, v)
					}
				}
			}
		}
	}

	// Don't start batch submitter and proposer if there's no sequencer.
	if sys.RollupNodes["sequencer"] == nil {
		return sys, nil
	}

	// L2Output Submitter
	sys.L2OutputSubmitter, err = l2os.NewL2OutputSubmitterFromCLIConfig(l2os.CLIConfig{
		L1EthRpc:          sys.Nodes["l1"].WSEndpoint(),
		RollupRpc:         sys.RollupNodes["sequencer"].HTTPEndpoint(),
		L2OOAddress:       config.L1Deployments.L2OutputOracleProxy.Hex(),
		PollInterval:      50 * time.Millisecond,
		TxMgrConfig:       newTxMgrConfig(sys.Nodes["l1"].WSEndpoint(), cfg.Secrets.Proposer),
		AllowNonFinalized: cfg.NonFinalizedProposals,
		LogConfig: oplog.CLIConfig{
			Level:  "info",
			Format: "text",
		},
	}, sys.cfg.Loggers["proposer"], proposermetrics.NoopMetrics)
	if err != nil {
		return nil, fmt.Errorf("unable to setup l2 output submitter: %w", err)
	}

	if err := sys.L2OutputSubmitter.Start(); err != nil {
		return nil, fmt.Errorf("unable to start l2 output submitter: %w", err)
	}

	// Batch Submitter
	sys.BatchSubmitter, err = bss.NewBatchSubmitterFromCLIConfig(bss.CLIConfig{
		L1EthRpc:               sys.Nodes["l1"].WSEndpoint(),
		L2EthRpc:               sys.Nodes["sequencer"].WSEndpoint(),
		RollupRpc:              sys.RollupNodes["sequencer"].HTTPEndpoint(),
		MaxPendingTransactions: 0,
		MaxChannelDuration:     1,
		MaxL1TxSize:            240_000,
		CompressorConfig: compressor.CLIConfig{
			TargetL1TxSizeBytes: cfg.BatcherTargetL1TxSizeBytes,
			TargetNumFrames:     1,
			ApproxComprRatio:    0.4,
		},
		SubSafetyMargin: 4,
		PollInterval:    50 * time.Millisecond,
		TxMgrConfig:     newTxMgrConfig(sys.Nodes["l1"].WSEndpoint(), cfg.Secrets.Batcher),
		LogConfig: oplog.CLIConfig{
			Level:  "info",
			Format: "text",
		},
	}, sys.cfg.Loggers["batcher"], batchermetrics.NoopMetrics)
	if err != nil {
		return nil, fmt.Errorf("failed to setup batch submitter: %w", err)
	}

	// Batcher may be enabled later
	if !sys.cfg.DisableBatcher {
		if err := sys.BatchSubmitter.Start(); err != nil {
			return nil, fmt.Errorf("unable to start batch submitter: %w", err)
		}
	}

	return sys, nil
}

// IP6 range that gets blackholed (in case our traffic ever makes it out onto
// the internet).
var blackholeIP6 = net.ParseIP("100::")

// mocknet doesn't allow us to add a peerstore without fully creating the peer ourselves
func (sys *System) newMockNetPeer() (host.Host, error) {
	sk, _, err := ic.GenerateECDSAKeyPair(rand.Reader)
	if err != nil {
		return nil, err
	}
	id, err := peer.IDFromPrivateKey(sk)
	if err != nil {
		return nil, err
	}
	suffix := id
	if len(id) > 8 {
		suffix = id[len(id)-8:]
	}
	ip := append(net.IP{}, blackholeIP6...)
	copy(ip[net.IPv6len-len(suffix):], suffix)
	a, err := ma.NewMultiaddr(fmt.Sprintf("/ip6/%s/tcp/4242", ip))
	if err != nil {
		return nil, fmt.Errorf("failed to create test multiaddr: %w", err)
	}
	p, err := peer.IDFromPublicKey(sk.GetPublic())
	if err != nil {
		return nil, err
	}

	ps, err := pstoremem.NewPeerstore()
	if err != nil {
		return nil, err
	}
	ps.AddAddr(p, a, peerstore.PermanentAddrTTL)
	_ = ps.AddPrivKey(p, sk)
	_ = ps.AddPubKey(p, sk.GetPublic())

	ds := sync.MutexWrap(ds.NewMapDatastore())
	eps, err := store.NewExtendedPeerstore(context.Background(), log.Root(), clock.SystemClock, ps, ds, 24*time.Hour)
	if err != nil {
		return nil, err
	}
	return sys.Mocknet.AddPeerWithPeerstore(p, eps)
}

func selectEndpoint(node *node.Node) string {
	useHTTP := os.Getenv("OP_E2E_USE_HTTP") == "true"
	if useHTTP {
		log.Info("using HTTP client")
		return node.HTTPEndpoint()
	}
	return node.WSEndpoint()
}

func httpEndpointForDocker(node *node.Node) string {
	url, err := url.Parse(node.HTTPEndpoint())
	if err != nil {
		panic(fmt.Sprintf("geth HTTPEndpoint returned malformed URL (%v)", err))
	}
	port := url.Port()

	// This is how Docker containers address services running on the host.
	return fmt.Sprintf("http://host.docker.internal:%s", port)
}

func configureL1(rollupNodeCfg *rollupNode.Config, l1Node *node.Node) {
	l1EndpointConfig := selectEndpoint(l1Node)
	rollupNodeCfg.L1 = &rollupNode.L1EndpointConfig{
		L1NodeAddr:       l1EndpointConfig,
		L1TrustRPC:       false,
		L1RPCKind:        sources.RPCKindBasic,
		RateLimit:        0,
		BatchSize:        20,
		HttpPollInterval: time.Millisecond * 100,
	}
}
func configureL2(rollupNodeCfg *rollupNode.Config, l2Node *node.Node, espresso *EspressoSystem, jwtSecret [32]byte) {
	useHTTP := os.Getenv("OP_E2E_USE_HTTP") == "true"
	l2EndpointConfig := l2Node.WSAuthEndpoint()
	if useHTTP {
		l2EndpointConfig = l2Node.HTTPAuthEndpoint()
	}

	rollupNodeCfg.L2 = &rollupNode.L2EndpointConfig{
		L2EngineAddr:      l2EndpointConfig,
		L2EngineJWTSecret: jwtSecret,
	}
	if espresso != nil {
		rollupNodeCfg.EspressoUrl = espresso.SequencerUrl()
	}
}

func (cfg SystemConfig) L1ChainIDBig() *big.Int {
	return new(big.Int).SetUint64(cfg.DeployConfig.L1ChainID)
}

func (cfg SystemConfig) L2ChainIDBig() *big.Int {
	return new(big.Int).SetUint64(cfg.DeployConfig.L2ChainID)
}

func hexPriv(in *ecdsa.PrivateKey) string {
	b := e2eutils.EncodePrivKey(in)
	return hexutil.Encode(b)
}
