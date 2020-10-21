package controllers

import (
	"bytes"
	"github.com/versent/saml2aws/v2/pkg/awsconfig"
	ini "gopkg.in/ini.v1"
)

func generateIni(profile string, creds *awsconfig.AWSCredentials) []byte {
	iniFile := ini.Empty()
	sec, _ := iniFile.NewSection(profile)
	sec.ReflectFrom(creds)

	buf := bytes.Buffer{}
	iniFile.WriteTo(&buf)

	return buf.Bytes()
}
