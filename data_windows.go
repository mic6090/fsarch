package fsarch

import (
	"os"
)

func GetEtcPath() string {
	if userProfile, ok := os.LookupEnv("LOCALAPPDATA"); ok {
		return userProfile
	}
	return "."
}
