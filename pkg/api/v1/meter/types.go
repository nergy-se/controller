package meter

import "time"

type Data struct {
	Id          string    `json:"id"`
	Model       string    `json:"model"`
	Time        time.Time `json:"time"`
	Current_W   float64   `json:"w,omitempty"`
	Current_VLL float64   `json:"vll,omitempty"`
	Current_VLN float64   `json:"vln,omitempty"`
	Total_WH    float64   `json:"wh,omitempty"`
	L1_A        float64   `json:"l1_a,omitempty"`
	L2_A        float64   `json:"l2_a,omitempty"`
	L3_A        float64   `json:"l3_a,omitempty"`
	L1_V        float64   `json:"l1_v,omitempty"`
	L2_V        float64   `json:"l2_v,omitempty"`
	L3_V        float64   `json:"l3_v,omitempty"`
}
