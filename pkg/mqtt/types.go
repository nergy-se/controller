package mqtt

import (
	"time"

	"github.com/nergy-se/controller/pkg/api/v1/meter"
)

type P1ib struct {
	P1IbHourlyActiveImportQ1Q4   float64 `json:"p1ib_hourly_active_import_q1_q4"`
	P1IbHourlyActiveExportQ2Q3   float64 `json:"p1ib_hourly_active_export_q2_q3"`
	P1IbHourlyReactiveImportQ1Q2 float64 `json:"p1ib_hourly_reactive_import_q1_q2"`
	P1IbHourlyReactiveExportQ3Q4 float64 `json:"p1ib_hourly_reactive_export_q3_q4"`
	P1IbActivePowerPlusQ1Q4      float64 `json:"p1ib_active_power_plus_q1_q4"`
	P1IbActivePowerMinusQ2Q3     int     `json:"p1ib_active_power_minus_q2_q3"`
	P1IbReactivePowerPlusQ1Q2    int     `json:"p1ib_reactive_power_plus_q1_q2"`
	P1IbReactivePowerMinusQ3Q4   float64 `json:"p1ib_reactive_power_minus_q3_q4"`
	P1IbActivePowerPlusL1        float64 `json:"p1ib_active_power_plus_l1"`
	P1IbActivePowerMinusL1       int     `json:"p1ib_active_power_minus_l1"`
	P1IbActivePowerPlusL2        float64 `json:"p1ib_active_power_plus_l2"`
	P1IbActivePowerMinusL2       int     `json:"p1ib_active_power_minus_l2"`
	P1IbActivePowerPlusL3        float64 `json:"p1ib_active_power_plus_l3"`
	P1IbActivePowerMinusL3       int     `json:"p1ib_active_power_minus_l3"`
	P1IbReactivePowerPlusL1      int     `json:"p1ib_reactive_power_plus_l1"`
	P1IbReactivePowerMinusL1     float64 `json:"p1ib_reactive_power_minus_l1"`
	P1IbReactivePowerPlusL2      int     `json:"p1ib_reactive_power_plus_l2"`
	P1IbReactivePowerMinusL2     float64 `json:"p1ib_reactive_power_minus_l2"`
	P1IbReactivePowerPlusL3      int     `json:"p1ib_reactive_power_plus_l3"`
	P1IbReactivePowerMinusL3     float64 `json:"p1ib_reactive_power_minus_l3"`
	P1IbVoltageL1                float64 `json:"p1ib_voltage_l1"`
	P1IbVoltageL2                int     `json:"p1ib_voltage_l2"`
	P1IbVoltageL3                float64 `json:"p1ib_voltage_l3"`
	P1IbCurrentL1                float64 `json:"p1ib_current_l1"`
	P1IbCurrentL2                float64 `json:"p1ib_current_l2"`
	P1IbCurrentL3                float64 `json:"p1ib_current_l3"`
	P1IbFirmware                 string  `json:"p1ib_firmware"`
	P1IbUpdateAvailable          string  `json:"p1ib_update_available"`
	P1IbImportExportL1           float64 `json:"p1ib_import_export_l1"`
	P1IbImportExportL2           float64 `json:"p1ib_import_export_l2"`
	P1IbImportExportL3           float64 `json:"p1ib_import_export_l3"`
	P1IbImportExport             float64 `json:"p1ib_import_export"`
	P1IbRssi                     string  `json:"p1ib_rssi"`
	P1IbMeter                    string  `json:"p1ib_meter"`
	P1IbTelegramsCrcOk           int     `json:"p1ib_telegrams_crc_ok"`
	P1IbIP                       string  `json:"p1ib_ip"`
	P1IbWifiMac                  string  `json:"p1ib_wifi_mac"`
}

func (p P1ib) AsMeterData(id string) meter.Data {
	return meter.Data{
		Id:          id,
		Model:       "p1ib",
		Time:        time.Time{},
		Current_W:   0.0,
		Current_VLL: 0.0,
		Current_VLN: 0.0,
		Total_WH:    0.0,
		L1_A:        0.0,
		L2_A:        0.0,
		L3_A:        0.0,
		L1_V:        0.0,
		L2_V:        0.0,
		L3_V:        0.0,
	}
}
