package modbusclient

import (
	"bytes"
	"encoding/binary"

	"github.com/goburrow/modbus"
)

type Client interface {
	ReadInputRegister(address uint16) (int, error)
	ReadHoldingRegister(address uint16) (int, error)
	ReadDiscreteInput(address uint16) ([]byte, error)
	WriteSingleRegister(address, value uint16) (results []byte, err error)
	WriteSingleCoil(address, value uint16) (int, error)
}

type client struct {
	client modbus.Client
}

func New(c modbus.Client) *client {
	return &client{
		client: c,
	}
}

func (c *client) ReadInputRegister(address uint16) (int, error) {
	b, err := c.client.ReadInputRegisters(address, 1)
	return Decode(b), err
}

func (c *client) ReadHoldingRegister(address uint16) (int, error) {
	b, err := c.client.ReadHoldingRegisters(address, 1)
	return Decode(b), err
}

func (c *client) ReadDiscreteInput(address uint16) ([]byte, error) {
	return c.client.ReadDiscreteInputs(address, 1)
}

func (c *client) WriteSingleRegister(address, value uint16) ([]byte, error) {
	return c.client.WriteSingleRegister(address, value)
}
func (c *client) WriteSingleCoil(address, value uint16) (int, error) {
	b, err := c.client.WriteSingleCoil(address, value)
	return Decode(b), err
}

func Decode(data []byte) int {
	var i int16
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.BigEndian, &i)
	return int(i)
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
