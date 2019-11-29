# secretgen

Requires kustomize 3.2.2 or higher.

A [kustomize](http://kustomize.io) plugin to generate Kubernetes secrets from sops encrypted files.

Env files generate a secret value per line, other file types use the file as value. Supports combining multiple files into a single secret.

## Installation

Kustomize looks for plugins in `~/.config/kustomize/plugin`.

```bash
go install sigs.k8s.io/kustomize/kustomize/v3
mkdir -p ~/.config/kustomize/plugin/secretgen
curl -Ls https://github.com/julienp/secretgen/releases/download/v0.1.1/secretgen_0.1.1_Darwin_x86_64.tar.gz | tar -C ~/.config/kustomize/plugin/secretgen -xz
```



### Manual Installation

```bash
mkdir -p ~/.config/kustomize/plugin/secretgen
go build -o ~/.config/kustomize/plugin/secretgen/secretgen cmd/main.go
```

## Usage

You have to pass the `enable_alpha_plugins` flag to kustumize when building:

```bash
kustomize build --enable_alpha_plugins my-manifests/
 ```

## Testing

Import the gpg key:

```bash
gpg --import test-key.asc
```
and run the tests:

```bash
go test -v .
```
