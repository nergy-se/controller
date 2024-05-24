package version

import (
	"encoding/json"
	"log"
	"runtime/debug"
)

var Version = func() string {
	type versionInfo struct {
		Commit string `json:"commit"`
		Time   string `json:"time"`
	}
	v := versionInfo{}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				v.Commit = setting.Value
			}
			if setting.Key == "vcs.time" {
				v.Time = setting.Value
			}
		}
	}
	b, err := json.Marshal(&v)
	if err != nil {
		log.Fatal(err)
	}

	return string(b)
}()
