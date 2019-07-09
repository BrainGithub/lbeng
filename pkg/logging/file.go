package logging

import (
	"fmt"

	"lbeng/pkg/setting"
)

// getLogFilePath get the log file save path
func getLogFilePath() string {
	return fmt.Sprintf("%s", setting.AppSetting.LogSavePath)
}

// getLogFileName get the save name of the log file
func getLogFileName() string {
	return fmt.Sprintf("%s.%s",
		setting.AppSetting.LogSaveName,
		setting.AppSetting.LogFileExt,
	)
}
