package cmd

import (
	"fmt"
	"main/proxyctl"
)

func ListCores() {
	fmt.Printf("Xray [%v]", proxyctl.CoreVersion())
}
