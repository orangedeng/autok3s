package common

import (
	"sync"

	"github.com/orangedeng/autok3s/pkg/settings"
	"github.com/orangedeng/autok3s/pkg/utils"
)

var (
	installUUIDOnce  sync.Once
	installUUIDValue string
)

// SetupInstallUUID assigns the persistent identity of this autok3s
// installation. The value is generated once, stored in the settings table
// and reused by every later run, so nodes stamped with it can always be
// tied back to the installation that created them.
func SetupInstallUUID() error {
	var setupErr error
	installUUIDOnce.Do(func() {
		installUUID := settings.InstallUUID.Get()
		if installUUID == "" {
			installUUID, setupErr = utils.RandomToken(16)
			if setupErr != nil {
				return
			}
			if setupErr = settings.InstallUUID.Set(installUUID); setupErr != nil {
				return
			}
		}
		installUUIDValue = installUUID
	})
	return setupErr
}

// GetInstallUUID returns the install uuid of this autok3s installation,
// falling back to the settings value when setup has not run yet.
func GetInstallUUID() string {
	if installUUIDValue != "" {
		return installUUIDValue
	}
	return settings.InstallUUID.Get()
}
