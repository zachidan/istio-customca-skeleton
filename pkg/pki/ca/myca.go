package ca

import (
	"fmt"
	"time"

	"istio.io/istio/security/pkg/pki/util"
)

const certsFolder = "ca-samples"

type CustomCA struct {}

// NewCustomCA returns a new CusomCA instance.
func NewCustomCA() (*CustomCA, error) {
	ca := &CustomCA{}
	return ca, nil
}

func (ca *CustomCA) Run(stopChan chan struct{}) {
}

// Sign takes a PEM-encoded CSR, subject IDs and lifetime, and returns a signed certificate. If forCA is true,
// the signed certificate is a CA certificate, otherwise, it is a workload certificate.
func (ca *CustomCA) Sign(csrPEM []byte, subjectIDs []string, requestedLifetime time.Duration, forCA bool) ([]byte, error) {
	fmt.Println("TODO: add the Sign code here");
	return nil, nil
}

// SignWithCertChain is similar to Sign but returns the leaf cert and the entire cert chain.
func (ca *CustomCA) SignWithCertChain(csrPEM []byte, subjectIDs []string, ttl time.Duration, forCA bool) ([]byte, error) {
	fmt.Println("TODO: add the SignWithCertChain code here");
	return nil, nil
}

// GetCAKeyCertBundle returns the KeyCertBundle for the CA.
func (ca *CustomCA) GetCAKeyCertBundle() util.KeyCertBundle {
	fmt.Println("TODO: add the GetCAKeyCertBundle code here");
	bundle, _ := util.NewVerifiedKeyCertBundleFromFile(certsFolder + "/ca-cert.pem", certsFolder + "/ca-key.pem", certsFolder + "/cert-chain.pem", certsFolder + "/root-cert.pem")
	return bundle
}
