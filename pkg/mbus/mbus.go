package mbus

import (
	"strconv"
	"sync"
	"time"

	"github.com/jonaz/gombus"
	"github.com/nergy-se/controller/pkg/api/v1/meter"
)

type Mbus struct {
	conn  gombus.Conn
	mutex *sync.Mutex
}

func New() *Mbus {
	return &Mbus{
		mutex: &sync.Mutex{},
	}
}

func (m *Mbus) Init() error {
	c, err := gombus.DialSerial("/dev/ttyAMA0")
	m.mutex.Lock()
	m.conn = c
	m.mutex.Unlock()
	return err
}

func (m *Mbus) Close() error {
	return m.conn.Close()
}

func (m *Mbus) ReadValues(model, idStr string) (*meter.Data, error) {

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, err
	}

	frame, err := m.readPrimaryAddress(id)
	if err != nil {
		return nil, err
	}

	data := &meter.Data{
		Time: time.Now(),
	}
	switch model {
	case "garo-GNM3D-MBUS":
		//TODO which frames a which?
		data.Total_WH = frame.DataRecords[0].Value
	}

	return data, nil
}

func (m *Mbus) readPrimaryAddress(primaryAddr int) (*gombus.DecodedFrame, error) {
	_, err := m.conn.Write(gombus.SndNKE(uint8(primaryAddr)))
	if err != nil {
		return nil, err
	}

	err = m.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		return nil, err
	}

	_, err = gombus.ReadSingleCharFrame(m.conn)
	if err != nil {
		return nil, err
	}

	frame, err := gombus.ReadSingleFrame(m.conn, primaryAddr)
	if err != nil {
		return nil, err
	}

	return frame, nil
}
