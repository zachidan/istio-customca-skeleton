# Istio custom CA skeleton

This sample CA creates a gRPC service to accept CSR requests from the Istio Node agent. Upon receiving the CSR request, the CA should validate the request - if valid, the CSR should be approved and signed. Otherwise, an error should be returned.

The [skeleton CA](pkg/pki/ca/myca.go) class should be modified by the user with the specific required logic.

## Running the sample

### Build 

```shell
$ go build -o mycustomca cmd/main.go
```

### Run 

```shell
$ ./mycustomca
```
