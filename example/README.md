# Example

This example is intended to deploy a Big Bang repo on a k3d / k3s server. You can setup a local k3d cluster with the following command

```bash
k3d cluster create
  # Required by the PLG stack
  --volume /etc/machine-id:/etc/machine-id

  # Required for Istio ingress
  --k3s-arg "--disable=traefik@server:0"
  --port 80:80@loadbalancer
  --port 443:443@loadbalancer

  # Required for TLS to work correctly with kubectl
  --k3s-arg "--tls-san=$SERVER_IP@server:0"
  --api-port 6443
```