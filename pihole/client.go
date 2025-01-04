package pihole

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	telnet "github.com/reiver/go-telnet"
)

type Client interface {
	GetQueries(time.Time) ([]Query, error)
}

type TelnetClient struct {
	URL  string
	conn io.ReadWriteCloser
}

// TelnetClient must implement Client
var _ Client = &TelnetClient{}

func (c *TelnetClient) GetQueries(since time.Time) ([]Query, error) {
	closeConn, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer closeConn()

	qCh, errCh := processQueries(c.conn)

	t := strconv.Itoa(int(since.Unix()))
	now := strconv.Itoa(int(time.Now().Unix()))
	command := ">getallqueries-time " + t + " " + now
	err = c.sendCommand(command)
	if err != nil {
		return nil, err
	}

	var queries []Query
LOOP:
	for {
		select {
		case q := <-qCh:
			queries = append(queries, q)
		case err := <-errCh:
			if err != nil {
				println(err.Error())
			}
			break LOOP
		case <-time.After(3 * time.Second):
			closeConn()
		}
	}

	return queries, nil
}

func processQueries(r io.Reader) (chan Query, chan error) {
	qCh := make(chan Query)
	errCh := make(chan error)

	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			q, err := parseQueryLine(line)
			if err != nil {
				println(err.Error())
				continue
			}
			qCh <- q

		}
		err := scanner.Err()
		errCh <- err
		if err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
		}
	}()

	return qCh, errCh
}

func (c *TelnetClient) connect() (func(), error) {
	conn, err := telnet.DialTo(c.URL)
	if err != nil {
		return nil, err
	}

	c.conn = conn
	return func() {
		_, _ = c.conn.Write([]byte(">quit\n"))
	}, nil
}

func (c *TelnetClient) sendCommand(command string) error {
	_, err := c.conn.Write([]byte(command + "\n"))
	if err != nil {
		return err
	}

	return nil
}
