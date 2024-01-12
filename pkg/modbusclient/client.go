package modbusclient

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/goburrow/modbus"
	"github.com/sirupsen/logrus"
)

type Client interface {
	ReadInputRegister(address uint16) (int, error)
	ReadHoldingRegister32(address uint16) (int, error)
	ReadHoldingRegister16(address uint16) (int, error)
	ReadDiscreteInput(address uint16) ([]byte, error)
	WriteSingleRegister(address, value uint16) (results []byte, err error)
	WriteSingleCoil(address, value uint16) (int, error)
}

type client struct {
	client modbus.Client
	close  func() error
}

func New(c modbus.Client, close func() error) *client {
	return &client{
		client: c,
		close:  close,
	}
}
func (c *client) closeIfNeeded(e error) {
	if e == nil {
		return
	}

	if errors.Is(e, syscall.EPIPE) {
		logrus.Warn("reconnect due to broken pipe")
		err := c.close()
		if err != nil {
			logrus.Error("error closing client: %w", err)
		}
	}

	if errors.Is(e, os.ErrDeadlineExceeded) {
		logrus.Warn("reconnect due to i/o timeout")
		err := c.close()
		if err != nil {
			logrus.Error("error closing client: %w", err)
		}
	}
}

func (c *client) ReadInputRegister(address uint16) (int, error) {
	b, err := c.client.ReadInputRegisters(address, 1)
	if err != nil {
		c.closeIfNeeded(err)
		err = fmt.Errorf("error reading address %d: %w", address, err)
	}
	return Decode(b), err
}

func (c *client) ReadHoldingRegister16(address uint16) (int, error) {
	return c.readHoldingRegister(address, 1)
}

func (c *client) ReadHoldingRegister32(address uint16) (int, error) {
	return c.readHoldingRegister(address, 2)
}

func (c *client) readHoldingRegister(address, count uint16) (int, error) {
	b, err := c.client.ReadHoldingRegisters(address, count)
	if err != nil {
		c.closeIfNeeded(err)
		err = fmt.Errorf("error reading address %d: %w", address, err)
	}
	return Decode(b), err
}

func (c *client) ReadDiscreteInput(address uint16) ([]byte, error) {
	b, err := c.client.ReadDiscreteInputs(address, 1)

	if err != nil {
		c.closeIfNeeded(err)
		err = fmt.Errorf("error reading address %d: %w", address, err)
	}
	return b, err
}

func (c *client) WriteSingleRegister(address, value uint16) ([]byte, error) {
	b, err := c.client.WriteSingleRegister(address, value)
	if err != nil {
		c.closeIfNeeded(err)
		err = fmt.Errorf("error writing address %d value %d error: %w", address, value, err)
	}
	return b, err
}
func (c *client) WriteSingleCoil(address, value uint16) (int, error) {
	b, err := c.client.WriteSingleCoil(address, value)
	if err != nil {
		c.closeIfNeeded(err)
		err = fmt.Errorf("error writing address %d value %d error: %w", address, value, err)
	}
	return Decode(b), err
}

// Decode High byte first high word first (big endian)
func Decode(data []byte) int {

	switch len(data) {
	case 1:
		var i int8
		binary.Read(bytes.NewBuffer(data), binary.BigEndian, &i)
		return int(i)
	case 2:
		var i int16
		binary.Read(bytes.NewBuffer(data), binary.BigEndian, &i)
		return int(i)
	case 4:
		var i int32
		binary.Read(bytes.NewBuffer(data), binary.BigEndian, &i)
		return int(i)
	case 8:
		var i int64
		binary.Read(bytes.NewBuffer(data), binary.BigEndian, &i)
		return int(i)
	}

	return 0
}

func CoilValue(b bool) uint16 {
	if b {
		return WriteCoilValueOn
	}
	return WriteCoilValueOff
}

const (
	WriteCoilValueOn  uint16 = 0xff00
	WriteCoilValueOff uint16 = 0
)
