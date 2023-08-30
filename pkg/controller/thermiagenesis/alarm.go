package thermiagenesis

import "github.com/sirupsen/logrus"

var alarmsMap = map[int]string{
	0:   "Alarm active, Class: A",
	1:   "Alarm active, Class: B",
	2:   "Alarm active, Class: C",
	3:   "Alarm active, Class: D - Genesis secondary",
	4:   "Alarm active, Class: E - Legacy secondary",
	9:   "High pressure switch alarm",
	10:  "Low pressure level alarm",
	11:  "High discharge pipe temperature alarm",
	12:  "Operating pressure limit indication",
	13:  "Discharge pipe sensor alarm",
	14:  "Liquid line sensor alarm",
	15:  "Suction gas sensor alarm",
	16:  "Flow/pressure switch alarm",
	22:  "Power input phase detection alarm",
	23:  "Inverter unit alarm",
	24:  "System supply low temperature alarm",
	25:  "Compressor low speed alarm",
	26:  "Low super heat alarm",
	27:  "Pressure ratio out of range alarm",
	28:  "Compressor pressure outside envelope alarm",
	29:  "Brine temperature out of range alarm",
	30:  "Brine in sensor alarm",
	31:  "Brine out sensor alarm",
	32:  "Condenser in sensor alarm",
	33:  "Condenser out sensor alarm",
	34:  "Outdoor sensor alarm",
	35:  "System supply line sensor alarm",
	36:  "Mix valve 1 supply line sensor alarm",
	37:  "Mix valve 2 supply line sensor alarm (EM)",
	38:  "Mix valve 3 supply line sensor alarm (EM)",
	39:  "Mix valve 4 supply line sensor alarm (EM)",
	40:  "Mix valve 5 supply line sensor alarm (EM)",
	44:  "WCS return line sensor alarm (EM)",
	45:  "TWC supply line sensor alarm (EM)",
	46:  "Cooling tank sensor alarm (EM)",
	47:  "Cooling supply line sensor alarm (EM)",
	48:  "Cooling circuit return line sensor alarm (EM)",
	49:  "Brine delta out of range alarm",
	50:  "Tap water mid sensor alarm",
	51:  "TWC circulation return sensor alarm (EM)",
	55:  "Brine in high temperature alarm",
	56:  "Brine in low temperature alarm",
	57:  "Brine out low temperature alarm",
	58:  "TWC circulation return low temperature alarm (EM)",
	59:  "TWC supply low temperature alarm (EM)",
	60:  "Mix valve 1 supply temperature deviation alarm",
	61:  "Mix valve 2 supply temperature deviation alarm (EM)",
	62:  "Mix valve 3 supply temperature deviation alarm (EM)",
	63:  "Mix valve 4 supply temperature deviation alarm (EM)",
	64:  "Mix valve 5 supply temperature deviation alarm (EM)",
	65:  "WCS return line temperature deviation alarm (EM)",
	66:  "Sum alarm",
	67:  "Cooling circuit supply line temperature deviation alarm (EM)",
	68:  "Cooling tank temperature deviation alarm (EM)",
	69:  "Surplus heat temperature deviation alarm (EM)",
	70:  "Humidity room sensor alarm",
	71:  "Surplus heat supply line sensor alarm (EM)",
	72:  "Surplus heat return line sensor alarm (EM)",
	73:  "Cooling tank return line sensor alarm (EM)",
	74:  "Temperature room sensor alarm",
	75:  "Inverter unit communication alarm",
	76:  "Pool return line sensor alarm",
	77:  "External stop for pool, read only",
	78:  "External start brine pump, read only",
	79:  "External relay for brine/ground water pump.",
	81:  "Tap water end tank sensor alarm",
	83:  "Genesis secondary unit alarm - this specific secondary unit can't communicate with its primary unit",
	84:  "Primary unit alarm - the primary has detected other primary units on the same network with a network mask that is allowing conflict. Change network settings in order to avoid problem. For instance change port number on the primary and its secondary unit.",
	85:  "Primary unit alarm - the primary has not detected all secondary units. Make sure that the primary/secondary settings are correct and the network mask and port and number of Genesis secondaries settings are correct. on/off 10 86 10087 1 Oil boost in progress",
	87:  "Tap water top sensor alarm.",
	202: "External alarm input",
}

func (ts *Thermiagenesis) Alarms() ([]string, error) {
	b, err := ts.client.ReadDiscreteInputs(0, 203)
	if err != nil {
		return nil, err
	}

	errs := make([]string, 0)
	for i, desc := range alarmsMap {
		val := b[i]
		if val == 0 {
			continue
		}
		if val == 1 { // byte array with 1 means true [1]
			errs = append(errs, desc)
			continue
		}

		logrus.Warnf("found input for %d (%s) with unexpected value: %#v", i, desc, val)
	}
	return errs, nil
}
