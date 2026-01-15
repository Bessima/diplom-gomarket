package config

import (
	"flag"
)

const defaultDBDNS = ""

type Flags struct {
	address string

	dbDNS          string
	accrualAddress string
}

func (flags *Flags) Init() {
	flag.StringVar(&flags.address, "a", ":8080", "Address and port to run server")

	flag.StringVar(&flags.dbDNS, "d", defaultDBDNS, "db dns")
	flag.StringVar(&flags.accrualAddress, "r", "", "accrual address")

	flag.Parse()
}
