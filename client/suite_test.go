package client

import (
	pb "MintDb/proto"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const rpcTimeout = 5 * time.Second

// repoRoot is the MintDb module root; go test runs from client/.
var repoRoot, _ = filepath.Abs("..")

// MintDbSuite runs every test against a real server process.
//
//	SetupSuite       build the server binary once
//	SetupTest        fresh data dir, start the server, connect a client
//	SetupSubTest     same, so every s.Run case gets its own server
//	TearDownSubTest  dump server.log and mint.aof if the case failed, stop the server
//	TearDownTest     same
//	TearDownSuite    delete the binary
//
// Data dirs live under <repo>/testrun/<test name>. They are kept after the run
// for inspection and wiped when that test runs again.
type MintDbSuite struct {
	suite.Suite
	binDir string

	server *testServer
	conn   *grpc.ClientConn
	client pb.MintDbClient
}

func TestMintDb(t *testing.T) {
	suite.Run(t, new(MintDbSuite))
}

func (s *MintDbSuite) SetupSuite() {
	binDir, err := os.MkdirTemp("", "mintdb-bin-")
	s.Require().NoError(err)
	s.binDir = binDir

	build := exec.Command("go", "build", "-o", filepath.Join(binDir, "mintdb"), ".")
	build.Dir = repoRoot
	out, err := build.CombinedOutput()
	s.Require().NoError(err, "building server:\n%s", out)
}

func (s *MintDbSuite) TearDownSuite() {
	os.RemoveAll(s.binDir)
}

func (s *MintDbSuite) SetupTest()    { s.startFreshServer() }
func (s *MintDbSuite) TearDownTest() { s.stopServer() }

// SetupSubTest replaces the parent test's server, so cases can't leak state
// (or a corrupted log) into each other.
func (s *MintDbSuite) SetupSubTest() {
	s.stopServer()
	s.startFreshServer()
}

func (s *MintDbSuite) TearDownSubTest() { s.stopServer() }

func (s *MintDbSuite) startFreshServer() {
	dataDir := filepath.Join(repoRoot, "testrun", s.T().Name())
	s.Require().NoError(os.RemoveAll(dataDir))
	s.Require().NoError(os.MkdirAll(dataDir, 0o755))
	addr, err := freeAddr()
	s.Require().NoError(err)

	s.server = &testServer{bin: filepath.Join(s.binDir, "mintdb"), dataDir: dataDir, addr: addr}
	s.Require().NoError(s.server.Start())
	s.connect()
}

func (s *MintDbSuite) stopServer() {
	if s.server == nil {
		return
	}
	s.disconnect()
	if s.T().Failed() {
		s.logServerFiles()
	}
	s.NoError(s.server.Stop())
	s.server = nil
}

// crash kills the server without warning and starts it on the same data dir,
// so it comes back with only what was persisted to disk.
func (s *MintDbSuite) crash() {
	s.disconnect()
	s.Require().NoError(s.server.Kill())
	s.Require().NoError(s.server.Start())
	s.connect()
}

// restart stops the server gracefully and starts it on the same data dir.
func (s *MintDbSuite) restart() {
	s.disconnect()
	s.Require().NoError(s.server.Stop())
	s.Require().NoError(s.server.Start())
	s.connect()
}

func (s *MintDbSuite) connect() {
	conn, err := grpc.NewClient(s.server.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn
	s.client = pb.NewMintDbClient(conn)
}

func (s *MintDbSuite) disconnect() {
	if s.conn != nil {
		s.conn.Close()
		s.conn, s.client = nil, nil
	}
}

func (s *MintDbSuite) logServerFiles() {
	if log, err := os.ReadFile(filepath.Join(s.server.dataDir, "server.log")); err == nil {
		s.T().Logf("server.log:\n%s", log)
	}
	if aof, err := os.ReadFile(filepath.Join(s.server.dataDir, "mint.aof")); err == nil {
		s.T().Logf("mint.aof:\n%s", abbrev(string(aof)))
	}
}

// Client helpers. They fail the test on transport errors, so tests only deal
// with database behaviour.

func (s *MintDbSuite) set(key, value string) {
	ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)
	defer cancel()
	_, err := s.client.Set(ctx, &pb.SetRequest{Key: key, Value: value})
	s.Require().NoError(err, "Set(%q)", key)
}

func (s *MintDbSuite) del(key string) {
	ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)
	defer cancel()
	_, err := s.client.Delete(ctx, &pb.DeleteRequest{Key: key})
	s.Require().NoError(err, "Delete(%q)", key)
}

func (s *MintDbSuite) get(key string) (value string, found bool) {
	ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)
	defer cancel()
	res, err := s.client.Get(ctx, &pb.GetRequest{Key: key})
	if err != nil {
		if status.Convert(err).Message() == "key not found" {
			return "", false
		}
		s.Require().NoError(err, "Get(%q)", key)
	}
	return res.Value, true
}

// abbrev quotes s, shortening long values so failures stay readable.
func abbrev(s string) string {
	if len(s) <= 80 {
		return fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf("%q...(%d bytes)", s[:40], len(s))
}
