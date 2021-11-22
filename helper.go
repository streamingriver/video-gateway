package main

import "gitlab.com/avarf/getenvs"

func isDebug() bool {
	return getenvs.GetEnvString("DEBUG", "false") == "true"
}
