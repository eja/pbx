// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package main

import (
	"log"

	"github.com/eja/pbx/asterisk"
	pbxSys "github.com/eja/pbx/sys"
	pbxWeb "github.com/eja/pbx/web"

	tibulaSys "github.com/eja/tibula/sys"
	tibulaWeb "github.com/eja/tibula/web"
)

func main() {
	if err := pbxSys.Configure(); err != nil {
		log.Fatal(err)
	}

	if tibulaSys.Commands.DbSetup {
		if err := tibulaSys.Setup(); err != nil {
			log.Fatal(err)
		}
	} else if tibulaSys.Commands.Wizard {
		if err := tibulaSys.WizardSetup(); err != nil {
			log.Fatal(err)
		}
		if err := pbxSys.Wizard(); err != nil {
			log.Fatal(err)
		}

	} else if tibulaSys.Commands.Start {
		if err := pbxWeb.Router(); err != nil {
			log.Fatal(err)
		}
		if pbxSys.Options.Asterisk {
			go asterisk.Start()
		}
		if err := tibulaWeb.Start(); err != nil {
			log.Fatal("Cannot start the web service: ", err)
		}
	} else {
		pbxSys.Help()
	}
}
