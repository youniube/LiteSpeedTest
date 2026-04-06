package core

import (
	"github.com/xxf098/lite-proxy/dns"
	"github.com/xxf098/lite-proxy/transport/resolver"
)

func setDefaultResolver() {
	resolver.DefaultResolver = dns.DefaultResolver()
}
