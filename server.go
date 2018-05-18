package uploader

import (
	"crypto/tls"
	"os"
	"sync"
	"time"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
)

// Server rpc server struct
type Server struct {
	addr     string
	writeDir string
	certFile string
	keyFile  string
}

// NewServer init new server
func NewServer(addr, writeDirectory, certFile, keyFile string) *Server {
	return &Server{addr: addr, writeDir: writeDirectory,
		certFile: certFile, keyFile: keyFile}
}

// ListenAndServe start rpc server
func (srv *Server) ListenAndServe() error {
	cert, err := tls.LoadX509KeyPair(srv.certFile, srv.keyFile)
	if err != nil {
		return err
	}

	config := &tls.Config{Certificates: []tls.Certificate{cert}}

	tlsOpt := server.WithTLSConfig(config)

	session := &Session{mu: &sync.Mutex{}, files: make(map[SessionID]*os.File)}

	s := server.NewServer(tlsOpt)

	s.RegisterName(ServicePath, &RPC{server: srv, session: session}, "")
	s.Plugins.Add(serverplugin.NewRateLimitingPlugin(time.Second, 1000))

	return s.Serve(NetTCP, srv.addr)
}
