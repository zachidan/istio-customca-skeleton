package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	customca "github.com/zachidan/customca-skeleton/pkg/pki/ca"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	pkgcmd "istio.io/istio/pkg/cmd"
	kubelib "istio.io/istio/pkg/kube"
	"istio.io/istio/pkg/spiffe"
	"istio.io/istio/security/pkg/cmd"
	probecontroller "istio.io/istio/security/pkg/probe"
	"istio.io/istio/security/pkg/registry"
	"istio.io/istio/security/pkg/registry/kube"
	caserver "istio.io/istio/security/pkg/server/ca"
	"istio.io/pkg/collateral"
	"istio.io/pkg/env"
	"istio.io/pkg/log"
	"istio.io/pkg/version"
)

const (
	MaxDefaultMaxWorkloadCertTTL = 90 * 24 * time.Hour
)

type cliOptions struct { // nolint: maligned
	// Comma separated string containing all listened namespaces
	listenedNamespaces        string
	kubeConfigFile            string

	// Comma separated string containing all possible host name that clients may use to connect to.
	grpcHosts  string
	grpcPort   int

}

var (
	opts = cliOptions{}

	rootCmd = &cobra.Command{
		Use:   "istio_ca",
		Short: "Istio Certificate Authority (CA).",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			runCA()
		},
	}
)

func fatalf(template string, args ...interface{}) {
	if len(args) > 0 {
		log.Errorf(template, args...)
	} else {
		log.Errorf(template)
	}
	os.Exit(-1)
}

func init() {
	initCLI()
}

func initCLI() {
	flags := rootCmd.Flags()

	// General configuration.
	flags.StringVar(&opts.listenedNamespaces, "listened-namespace", metav1.NamespaceAll, "deprecated")
	if err := flags.MarkDeprecated("listened-namespace", "please use --listened-namespaces instead"); err != nil {
		panic(err)
	}

	// Default to NamespaceAll, which equals to "". Kuberentes library will then watch all the namespace.
	flags.StringVar(&opts.listenedNamespaces, "listened-namespaces", metav1.NamespaceAll,
		"Select the namespaces for the Citadel to listen to, separated by comma. If unspecified, Citadel tries to use the ${"+
			cmd.ListenedNamespaceKey+"} environment variable. If neither is set, Citadel listens to all namespaces.")

	flags.StringVar(&opts.kubeConfigFile, "kube-config", "",
		"Specifies path to kubeconfig file. This must be specified when not running inside a Kubernetes pod.")

	// gRPC server for signing CSRs.
	flags.StringVar(&opts.grpcHosts, "grpc-host-identities", "istio-ca,istio-citadel",
		"The list of hostnames for istio ca server, separated by comma.")
	flags.IntVar(&opts.grpcPort, "grpc-port", 8060, "The port number for Citadel GRPC server. "+
		"If unspecified, Citadel will not serve GRPC requests.")

	rootCmd.AddCommand(version.CobraCommand())

	rootCmd.AddCommand(collateral.CobraCommand(rootCmd, &doc.GenManHeader{
		Title:   "CustomCA",
		Section: "custom_ca CLI",
		Manual:  "CustomCA",
	}))

	pkgcmd.AddFlags(rootCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Errora(err)
		os.Exit(-1)
	}
}

// fqdn returns the k8s cluster dns name for the Citadel service.
func fqdn() string {
	return fmt.Sprintf("customca.default.svc.cluster.local")
}

var (
	listenedNamespaceKeyVar = env.RegisterStringVar(cmd.ListenedNamespaceKey, "", "")
)

func runCA() {
	if value, exists := listenedNamespaceKeyVar.Lookup(); exists {
		// When -namespace is not set, try to read the namespace from environment variable.
		if opts.listenedNamespaces == "" {
			opts.listenedNamespaces = value
		}
	}

	listenedNamespaces := strings.Split(opts.listenedNamespaces, ",")

	cs, err := kubelib.CreateClientset(opts.kubeConfigFile, "")
	if err != nil {
		fatalf("Could not create k8s clientset: %v", err)
	}
	stopCh := make(chan struct{})
	ca := createCA(stopCh)

	log.Info("CustomCA is running in server only mode.")
	if opts.grpcPort <= 0 {
		fatalf("The gRPC port must be set.")
	}

	// start registry if gRPC server is to be started
	reg := registry.GetIdentityRegistry()

	// add certificate identity to the identity registry for the liveness probe check
	if registryErr := reg.AddMapping(probecontroller.LivenessProbeClientIdentity,
		probecontroller.LivenessProbeClientIdentity); registryErr != nil {
		log.Errorf("Failed to add identity mapping: %v", registryErr)
	}

	ch := make(chan struct{})

	// monitor service objects with "alpha.istio.io/kubernetes-serviceaccounts" and
	// "alpha.istio.io/canonical-serviceaccounts" annotations
	serviceController := kube.NewServiceController(cs.CoreV1(), listenedNamespaces, reg)
	serviceController.Run(ch)

	// monitor service account objects for istio mesh expansion
	serviceAccountController := kube.NewServiceAccountController(cs.CoreV1(), listenedNamespaces, reg)
	serviceAccountController.Run(ch)

	hostnames := append(strings.Split(opts.grpcHosts, ","), fqdn())
	caServer, startErr := caserver.New(ca, MaxDefaultMaxWorkloadCertTTL,
		true, hostnames, opts.grpcPort, spiffe.GetTrustDomain(),
		true)
	if startErr != nil {
		fatalf("Failed to create Custom CA ca server: %v", startErr)
	}
	if serverErr := caServer.Run(); serverErr != nil {
		// stop the registry-related controllers
		ch <- struct{}{}

		log.Warnf("Failed to start GRPC server with error: %v", serverErr)
	}

	log.Info("CustomCA has started")

	// Capture termination and close the stop channel
	go pkgcmd.WaitSignal(stopCh)

	// Blocking until receives error.
	select {
	case <-stopCh:
		fmt.Println("Stopping CA server, termination signal received")
		return
	}
}

func createCA(stopCh chan struct{}) *customca.CustomCA {
	// TODO: Create your custom CA here
	customCA, err := customca.NewCustomCA()
	if err != nil {
		log.Errorf("Failed to create the CustomCA (error: %v)", err)
	}

	return customCA
}
