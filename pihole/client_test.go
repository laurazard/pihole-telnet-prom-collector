package pihole

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeConnection struct {
	clientWritten []string
	serverConn    bytes.Buffer
	closeCounter  int
}

func (f *fakeConnection) Write(b []byte) (int, error) {
	f.clientWritten = append(f.clientWritten, string(b))
	return len(b), nil
}

func (f *fakeConnection) Read(b []byte) (int, error) {
	return f.serverConn.Read(b)
}

func (f *fakeConnection) Close() error {
	f.closeCounter++
	return nil
}

func TestGetQueries(t *testing.T) {
	// TODO
	t.Run("connects", func(t *testing.T) {
		fakeConn := fakeConnection{}
		c := TelnetClient{
			URL:  "",
			conn: &fakeConn,
		}

		_ = c.sendCommand("meow")
		assert.Equal(t, "meow\n", fakeConn.clientWritten[0])
	})
}
