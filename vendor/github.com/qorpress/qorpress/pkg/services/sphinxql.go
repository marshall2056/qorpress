package services

import (
	"github.com/manticoresoftware/go-sdk/manticore"
)

var (
	cl manticore.Client
)

func InitSphinxQL(host string, port uint16) (manticore.Client, bool, error) {
	cl = manticore.NewClient()
	cl.SetServer(host, port)
	ok, err := cl.Open()
	return cl, ok, err
}
