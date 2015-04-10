package uninet

import (
	"errors"
	"net"

	"github.com/chzyer/adrs/utils"
	"gopkg.in/logex.v1"
)

var (
	ErrShortWritten = errors.New("short written")
)

type UDP struct {
	conn *net.UDPConn
}

func NewDialUDP(addr *URL) (*UDP, error) {
	c, err := net.Dial(addr.Network(), addr.Host())
	if err != nil {
		return nil, logex.Trace(err)
	}
	conn := c.(*net.UDPConn)
	conn.SetWriteBuffer(512)
	conn.SetReadBuffer(512)
	u := &UDP{
		conn: conn,
	}
	return u, nil
}

func NewUDP(host *URL) (*UDP, error) {
	addr, err := net.ResolveUDPAddr(host.Network(), host.Host())
	if err != nil {
		return nil, logex.Trace(err)
	}
	conn, err := net.ListenUDP(host.Network(), addr)
	if err != nil {
		return nil, logex.Trace(err)
	}
	conn.SetWriteBuffer(512)
	conn.SetReadBuffer(512)
	u := &UDP{
		conn: conn,
	}
	return u, nil
}

func (c *UDP) ReadBlock(b *utils.Block) error {
	n, err := c.conn.Read(b.All)
	if err != nil {
		return logex.Trace(err)
	}

	b.Length = n
	return nil
}

func (c *UDP) ReadBlockFrom(b *utils.Block) (*net.UDPAddr, error) {
	n, addr, err := c.conn.ReadFromUDP(b.All)
	if err != nil {
		return nil, logex.Trace(err)
	}

	b.Length = n
	return addr, nil
}

func (c *UDP) WriteBlock(b *utils.Block) error {
	n, err := c.conn.Write(b.Bytes())
	if err != nil {
		return logex.Trace(err)
	}
	if n != b.Length {
		return logex.Trace(ErrShortWritten)
	}
	return nil
}

func (c *UDP) WriteBlockTo(b *utils.Block, addr *net.UDPAddr) error {
	n, err := c.conn.WriteToUDP(b.Bytes(), addr)
	if err != nil {
		return logex.Trace(err)
	}
	if n != b.Length {
		return logex.Trace(ErrShortWritten)
	}
	return nil
}

func (c *UDP) Close() error {
	return c.conn.Close()
}
