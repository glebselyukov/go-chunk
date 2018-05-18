package uploader

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"time"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
)

// Client struct
type Client struct {
	Addr       string
	rpcxClient *client.Client
	crypt      string
}

// NewClient new client
func NewClient(addr string) *Client {
	return &Client{Addr: addr}
}

// Dial to rpcx srv
func (c *Client) Dial(timeout time.Duration) error {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	options := client.Option{
		TLSConfig:      conf,
		ConnectTimeout: timeout,
		SerializeType:  protocol.JSON,
	}

	cli := client.NewClient(options)

	cli.Connect(NetTCP, c.Addr)

	c.rpcxClient = cli

	return nil
}

// Close connection to rpcx srv
func (c *Client) Close() error {
	return c.rpcxClient.Close()
}

// Create file on server
func (c *Client) Create(filename, pathID string) (SessionID, error) {
	var res SesResponse

	if err := c.rpcxClient.Call(context.Background(),
		ServicePath, MethodCreateFile,
		FileRequest{Filename: filename, Path: pathID}, &res); err != nil {
		return "", err
	}

	return res.ID, nil
}

// CreatePath path on server
func (c *Client) CreatePath() (string, error) {
	var res PathResponse

	if err := c.rpcxClient.Call(context.Background(),
		ServicePath, MethodCreatePath,
		PathRequest{ID: CREATE}, &res); err != nil {
		return "", err
	}

	return res.ID, nil
}

// Stat get stat file on server
func (c *Client) Stat(filename string) (*StatResponse, error) {
	var res StatResponse
	if err := c.rpcxClient.Call(context.Background(),
		ServicePath, MethodStat,
		FileRequest{Filename: filename}, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetBlock set block to file
func (c *Client) SetBlock(filename string, blockID int) ([]byte, error) {
	return c.ReadLocalBlock(filename,
		int64(blockID)*BlockSize, BlockSize)
}

// ReadLocalBlock read custom block to local file
func (c *Client) ReadLocalBlock(filename string,
	offset int64, size int) ([]byte, error) {
	data := make([]byte, size)

	file, err := os.Open(filename)

	n, err := file.ReadAt(data, offset)
	if err != nil && err != io.EOF {
		return nil, err
	}

	if size > n {
		size = n
	}

	return data[:size], err
}

// WriteAt write block to file on server
func (c *Client) WriteAt(sessionID SessionID, offset int64, size int,
	data []byte, eof bool) error {
	res := new(WriteResponse)
	divCall := c.rpcxClient.Go(context.Background(), ServicePath, WriteAt,
		WriteRequest{ID: sessionID, Size: size, Offset: offset,
			Data: data, EOF: eof}, &res, nil)
	replyCall := <-divCall.Done
	return replyCall.Error
}

// CloseWriteSession close file on server
func (c *Client) CloseWriteSession(sessionID SessionID) error {
	res := &SesResponse{}
	if err := c.rpcxClient.Call(context.Background(),
		ServicePath, Close,
		SesRequest{ID: sessionID}, &res); err != nil {
		return err
	}
	return nil
}

// Upload upload file
func (c *Client) Upload(filename, pathID string) error {
	return c.UploadAt(filename, pathID, 0)
}

// UploadAt upload custom block of file
func (c *Client) UploadAt(filename,
pathID string, blockID int) error {
	stat, err := Stat(filename)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return fmt.Errorf("%s is directory", filename)
	}
	blocks := int(stat.Size / BlockSize)
	if stat.Size%BlockSize != 0 {
		blocks++
	}
	fmt.Printf("upload %s in %d blocks\n", filename, blocks-blockID)

	sessionID, err := c.Create(stat.Name, pathID)
	if err != nil {
		return err
	}

	for i := blockID; i < blocks; i++ {
		eof := false
		buf, rErr := c.SetBlock(filename, i)
		if rErr != nil && rErr != io.EOF {
			return rErr
		}
		if rErr == io.EOF {
			eof = true
		}
		err = c.WriteAt(sessionID, int64(i)*BlockSize, BlockSize, buf, eof)
		if err != nil {
			return err
		}
		if (int64(i)*BlockSize)%(int64(blocks-blockID)/100+1) == 0 {
			fmt.Printf("uploading %s [%d/%d] blocks\n", filename, i-blockID+1,
				blocks-blockID)
		}
	}
	fmt.Printf("upload %s completed\n", filename)

	c.CloseWriteSession(sessionID)

	return nil
}

// Stat get stat of local file
func Stat(filename string) (*StatResponse, error) {
	stat := new(StatResponse)

	fi, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		stat.Type = TypeDirectory
	} else {
		stat.Type = TypeFile
		stat.Size = fi.Size()
	}

	stat.LastModified = fi.ModTime()
	stat.Name = fi.Name()

	return stat, nil
}
