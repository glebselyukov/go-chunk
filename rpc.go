package uploader

import (
	"os"
	"path/filepath"
	"time"
	"context"
	"fmt"
	"github.com/pborman/uuid"
)

// SesRequest request SessionID
type SesRequest struct {
	ID SessionID
}

// FileRequest request file
type FileRequest struct {
	Filename string
	Path     string
}

// PathRequest request create path
type PathRequest struct {
	ID string
}

// WriteRequest request write block to file on server
type WriteRequest struct {
	ID     SessionID
	Offset int64
	Size   int
	Data   []byte
	EOF    bool
}

// WriteResponse response write block to file on server
type WriteResponse struct {
	ID     SessionID
	Offset int64
	Size   int
}

// StatResponse response statistic from server
type StatResponse struct {
	Type         string
	Size         int64
	LastModified time.Time
	Name         string
}

// SesResponse response SessionID
type SesResponse struct {
	ID     SessionID
	Result bool
}

// PathResponse response task
type PathResponse struct {
	ID     string
	Result bool
}

// RPC rpc session struct
type RPC struct {
	server  *Server
	session *Session
}

// Open open file on server
func (r *RPC) Open(ctx context.Context, req FileRequest, res *SesResponse) error {
	path := filepath.Join(r.server.writeDir, req.Filename)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	res.ID = r.session.Add(file)
	res.Result = true
	fmt.Printf("open %s, session ID = %s\n", req.Filename, res.ID)

	return nil
}

// Close close file on server
func (r *RPC) Close(ctx context.Context, req SesRequest, res *SesResponse) error {
	r.session.Delete(req.ID)
	res.Result = true
	fmt.Printf("close session ID = %s\n", req.ID)
	return nil
}

// WriteAt write block to file on server
func (r *RPC) WriteAt(ctx context.Context, req WriteRequest,
	res *WriteResponse) error {
	f := r.session.Get(req.ID)
	if f == nil {
		return fmt.Errorf("you must call open first\n")
	}
	_, err := f.WriteAt(req.Data, int64(req.Offset))
	if err != nil {
		return err
	}
	if req.EOF {
		f.Close()
	}
	return nil
}

// Create create file on server
func (r *RPC) Create(ctx context.Context, req FileRequest, res *SesResponse) error {
	if req.Path == "" {
		res.Result = false
	}

	file, err := os.OpenFile(
		filepath.Join(r.server.writeDir, req.Path, req.Filename),
		os.O_CREATE|os.O_WRONLY, FilePerm)
	if err != nil {
		return err
	}
	res.ID = r.session.Add(file)
	res.Result = true
	fmt.Printf("open %s, session ID = %s\n", req.Filename, res.ID)
	return nil
}

// CreatePath create task path on server
func (r *RPC) CreatePath(ctx context.Context, req PathRequest,
	res *PathResponse) error {
	if req.ID == CREATE {
		id := uuid.New()
		err := os.MkdirAll(filepath.Join(r.server.writeDir, id), os.ModePerm)
		if err != nil {
			res.Result = false
		} else {
			res.Result = true
			res.ID = id
		}
	}
	return nil
}

// IsDir test - is directory?
func (r *StatResponse) IsDir() bool {
	return r.Type == TypeDirectory
}
