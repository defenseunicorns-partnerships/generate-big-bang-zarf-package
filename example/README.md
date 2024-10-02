# Example

This example was generated using the following command

```bash
generate-big-bang-zarf-package 2.34.0 --values-file-manifests=values-files/kyverno.yaml,values-files/loki.yaml,values-files/neuvector.yaml
```

Note that the values-files directory was created before the generate command was run. The generated zarf.yaml changes depending on the values file manifests submitted as different values can enable different helm releases in Big Bang, which in turn require different repos and images. 

This example is intended to deploy a Big Bang on a k3d or k3s server. You can setup a local k3d cluster with the following command

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