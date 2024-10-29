package main

import (
	"fmt"
	"os"
	"path"
)

// Settings - provider settings file.
type MavenGlobalSettings struct {
	MavenRepoLocation string
	SharedDir         string
}

const GLOBAL_MAVEN_SETTINGS_KEY = "mavenGlobalSettings"
const GLOBAL_MAVEN_FILE_NAME = "globalSettings.xml"

func (m *MavenGlobalSettings) Build() (err error) {
	fileContentTemplate := `
<settings xmlns="http://maven.apache.org/SETTINGS/1.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
  xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.0.0 https://maven.apache.org/xsd/settings-1.0.0.xsd">
  <localRepository>%s</localRepository>
</settings>
	`
	path := path.Join(m.SharedDir, GLOBAL_MAVEN_FILE_NAME)
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	_, err = f.Write([]byte(fmt.Sprintf(fileContentTemplate, m.MavenRepoLocation)))
	return
}

// Function to update the settings with the maven global config
// to use the correct directory for the maven cache.
func (m *MavenGlobalSettings) UpdateSettings(settings *Settings) (err error) {
	for _, config := range settings.content {
		for _, initConfig := range config.InitConfig {
			if initConfig.ProviderSpecificConfig == nil {
				initConfig.ProviderSpecificConfig = map[string]interface{}{}
			}
			initConfig.ProviderSpecificConfig[GLOBAL_MAVEN_SETTINGS_KEY] = path.Join(m.SharedDir, GLOBAL_MAVEN_FILE_NAME)
		}
	}
	return nil
}
