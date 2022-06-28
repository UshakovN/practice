package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

// hardcode
const zyteProxyURL = "http://7577e1ecdebc423492288d1e857ffcde:@proxy.crawlera.com:8011/"

type ZyteProxy struct {
	caCertPool *x509.CertPool
}

func NewZyteProxy(certificatePath string) *ZyteProxy {
	caCert, err := ioutil.ReadFile(certificatePath)
	if err != nil {
		log.Fatalf("cannot load the certificate from %s: %v", certificatePath, err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return &ZyteProxy{
		caCertPool: caCertPool,
	}
}

func (p *ZyteProxy) GetHttpTransport() *http.Transport {
	proxyURL, err := url.Parse(zyteProxyURL)
	if err != nil {
		log.Fatalf("cannot parse proxy url %s: %v", zyteProxyURL, err)
	}
	return &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{
			RootCAs: p.caCertPool,
		},
	}
}
