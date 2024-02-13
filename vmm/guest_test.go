package vmm_test

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/c35s/hype/virtio"
	"github.com/c35s/hype/vmm"
	"github.com/mdlayher/vsock"
	"golang.org/x/sync/errgroup"
)

func TestConsole(t *testing.T) {
	GuestTest{
		Host: func(t *testing.T, runGuest func(vmm.Config)) {
			out := new(bytes.Buffer)
			runGuest(vmm.Config{
				Devices: []virtio.DeviceConfig{
					&virtio.ConsoleDevice{
						Out: out,
					},
				},
			})

			if !strings.Contains(out.String(), "hello from the guest") {
				t.Error("the guest didn't say hello")
			}
		},

		Guest: func(t *testing.T) {
			if _, err := fmt.Fprintln(os.Stdout, "hello from the guest"); err != nil {
				t.Fatal(err)
			}
		},
	}.Run(t)
}

func TestSocket(t *testing.T) {
	GuestTest{
		Host: func(t *testing.T, runGuest func(vmm.Config)) {
			sockdev := new(virtio.SocketDevice)
			lis, err := sockdev.Listen(vsock.Host, 1)
			if err != nil {
				t.Fatal(err)
			}

			defer lis.Close()

			go func() {
				conn, err := lis.Accept()
				if err != nil {
					panic(err)
				}

				defer conn.Close()

				b := make([]byte, os.Getpagesize())

				for {
					if err := conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
						panic(err)
					}

					n, err := conn.Read(b)
					if err != nil {
						return
					}

					if err := conn.SetWriteDeadline(time.Now().Add(1 * time.Second)); err != nil {
						panic(err)
					}

					if _, err := conn.Write(b[:n]); err != nil {
						panic(err)
					}
				}
			}()

			runGuest(vmm.Config{
				Devices: []virtio.DeviceConfig{
					sockdev,
				},
			})
		},

		Guest: func(t *testing.T) {
			conn, err := vsock.Dial(vsock.Host, 1, nil)
			if err != nil {
				t.Fatal(err)
			}

			data := make([]byte, 8*1024*1024)
			if _, err := io.ReadFull(rand.Reader, data); err != nil {
				t.Fatal(err)
			}

			eg := new(errgroup.Group)

			eg.Go(func() (err error) {
				defer func() {
					if err != nil {
						err = fmt.Errorf("write: %w", err)
					}
				}()

				if err := conn.SetWriteDeadline(time.Now().Add(1 * time.Second)); err != nil {
					return err
				}

				if n, err := io.Copy(conn, bytes.NewReader(data)); err != nil {
					return fmt.Errorf("after copying %d bytes to conn: %w", n, err)
				}

				return nil
			})

			buf := new(bytes.Buffer)

			eg.Go(func() (err error) {
				defer func() {
					if err != nil {
						err = fmt.Errorf("read: %w", err)
					}
				}()

				if err := conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
					return err
				}

				if n, err := io.CopyN(buf, conn, int64(len(data))); err != nil {
					return fmt.Errorf("after copying %d bytes from conn: %w", n, err)
				}

				return nil
			})

			if err := eg.Wait(); err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(data, buf.Bytes()) {
				t.Fatal("roundtrip failed")
			}

			if err := conn.Close(); err != nil {
				t.Fatal(err)
			}
		},
	}.Run(t)
	// sockdev := new(virtio.SocketDevice)
	// lis, err := sockdev.Listen(vsock.Host, 1)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// defer lis.Close()

	// go func() {
	// 	conn, err := lis.Accept()
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	defer conn.Close()

	// 	b := make([]byte, os.Getpagesize())

	// 	for {
	// 		if err := conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
	// 			panic(err)
	// 		}

	// 		n, err := conn.Read(b)
	// 		if err != nil {
	// 			return
	// 		}

	// 		if err := conn.SetWriteDeadline(time.Now().Add(1 * time.Second)); err != nil {
	// 			panic(err)
	// 		}

	// 		if _, err := conn.Write(b[:n]); err != nil {
	// 			panic(err)
	// 		}
	// 	}
	// }()

	// runGuest(t, sockdev)
}

func TestSocketGuest(t *testing.T) {

}
