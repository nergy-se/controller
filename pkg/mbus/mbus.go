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

func (m *Mbus) init() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.conn != nil {
		return nil
	}
	c, err := gombus.DialSerial("/dev/ttyAMA0")
	if err != nil {
		return err
	}
	m.conn = c
	return nil
}

func (m *Mbus) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.conn != nil {
		err := m.conn.Close()
		m.conn = nil
		return err
	}
	return nil
}

func (m *Mbus) ReadValues(model, idStr string) (*meter.Data, error) {
	err := m.init()
	if err != nil {
		return nil, err
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, err
	}

	frame, err := m.read(id)
	if err != nil {
		return nil, err
	}

	data := &meter.Data{
		Id:    idStr,
		Model: model,
		Time:  time.Now(),
	}
	switch model {
	case "garo-GNM3D-MBUS":
		data.Total_WH = frame.DataRecords[0].Value
		data.Current_W = frame.DataRecords[2].Value
		data.Current_VLL = frame.DataRecords[6].Value
		data.Current_VLN = frame.DataRecords[7].Value
		data.L1_A = frame.DataRecords[8].Value
		data.L2_A = frame.DataRecords[9].Value
		data.L3_A = frame.DataRecords[10].Value
	}

	return data, nil
}

func (m *Mbus) read(primaryAddr int) (*gombus.DecodedFrame, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
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

	return gombus.ReadSingleFrame(m.conn, primaryAddr)
}
