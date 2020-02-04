# Istio custom CA skeleton

This sample CA creates a gRPC service to accept CSR requests from the Istio Node agent. Upon receiving the CSR request, it should validate the request - if valid, CSR should be approved and signed. Otherwise, an error should be returned.
The `pkg/pki/ca/myca.go` is the skeleton CA that should be modified by the user.

## Running the sample

### Build 

```shell
$ go build -o mycustomca cmd/main.go
```

### Run 

```shell
$ ./mycustomca
```
