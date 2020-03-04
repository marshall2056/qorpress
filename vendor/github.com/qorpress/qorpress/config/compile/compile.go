// based on https://github.com/qorpress/bindatafs#usage
// extended with https://github.com/qorpress/bindatafs#using-namespace
package main

import (
	"log"

	"github.com/qorpress/qorpress/config/bindatafs"
)

func main() {
	assetFS := bindatafs.AssetFS

	// Register view paths into AssetFS under "admin" namespace
	err := assetFS.NameSpace("admin").RegisterPath("qor/admin/views")
	if err != nil {
		log.Fatalln("RegisterPath:", "qor/admin/views", err)
	}

	// Compile templates under registered view paths into binary
	err = assetFS.Compile()
	if err != nil {
		log.Fatalln("Compile:", err)
	}
}
