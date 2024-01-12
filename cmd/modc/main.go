package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/goburrow/modbus"
	"github.com/nergy-se/controller/pkg/modbusclient"
)

var decimals = flag.Int("decimals", 2, "")
var readCount = flag.Uint("read-count", 1, "how many addreses to read")

func main() {
	address := flag.String("addr", "", "tcp modbus address")

	inputreg := flag.Int("inputreg", 0, "input reg")
	discreteInput := flag.Int("discreteinputreg", 0, "descrete input reg")
	holdingreg := flag.Int("holdingreg", 0, "")
	coil := flag.Int("coil", 0, "")

	slaveID := flag.Int("slave", 0, "modbus slave id")
	value := flag.Int("value", 0, "value to write. will write any value")
	flag.Parse()

	handler := modbus.NewTCPClientHandler(*address)
	handler.SlaveId = byte(*slaveID)
	mcli := modbus.NewClient(handler)
	client := &Client{client: mcli}

	var f interface{}
	var err error
	if isFlagPassed("inputreg") {
		f, err = scale100itof(client.readInputRegister(uint16(*inputreg)))
	}
	if isFlagPassed("holdingreg") {
		if isFlagPassed("value") {
			f, err = client.client.WriteSingleRegister(uint16(*holdingreg), uint16(*value))
		} else {
			f, err = scale100itof(client.readHoldingRegister(uint16(*holdingreg)))
		}
	}

	if isFlagPassed("coil") {
		if isFlagPassed("value") {
			f, err = client.client.WriteSingleCoil(uint16(*coil), uint16(*value))
			// value måste vara 0xff00 dvs INT 65280
		} else {
			f, err = client.client.ReadCoils(uint16(*coil), 1)
		}
	}
	if isFlagPassed("discreteinputreg") {
		f, err = client.client.ReadDiscreteInputs(uint16(*discreteInput), 1)
	}

	if err != nil {
		log.Println("error was: ", err)
	}
	if v, ok := f.([]byte); ok {
		fmt.Printf("raw response: %# x (length: %d)\n", v, len(v))
	}
	log.Println("value is: ", f)
	handler.Close()
}
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
func IntPow(base, exp int) float64 {
	result := 1
	for {
		if exp&1 == 1 {
			result *= base
		}
		exp >>= 1
		if exp == 0 {
			break
		}
		base *= base
	}

	return float64(result)
}
func scale100itof(i int, err error) (float64, error) {
	f := float64(i) / IntPow(10, *decimals)
	return f, err
}

type Client struct {
	client modbus.Client
}

func (ts *Client) readInputRegister(address uint16) (int, error) {
	b, err := ts.client.ReadInputRegisters(address, uint16(*readCount))
	fmt.Printf("raw response: %# x (length: %d)\n", b, len(b))
	return modbusclient.Decode(b), err
}

func (ts *Client) readHoldingRegister(address uint16) (int, error) {
	b, err := ts.client.ReadHoldingRegisters(address, uint16(*readCount))
	fmt.Printf("raw response: %# x (length: %d)\n", b, len(b))
	return modbusclient.Decode(b), err
}
