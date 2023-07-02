# HC Vault Util 

`hc-vault-util` is a companion CLI tool for managing Hashicorp Vault.

## Features

- Vault transit backend import private key using key wrapping 
    - See [transit-import-key Tutorial](https://github.com/vdbulcke/terraform-vault-sample/blob/main/tutorial/transit-import-key/README.md)
- Generate CSR from Vault transit key using [cfssl json csr format](https://github.com/cloudflare/cfssl#signing)
    - See [transit-gencsr Tutorial](https://github.com/vdbulcke/terraform-vault-sample/blob/main/tutorial/transit-gencsr/README.md)

- Vault Kv2 TUI: using vim key bindings (`h`, `j`, `k`, `l`) for quickly navigating your Vault kv2 secrets in your terminal.

[Changelog](./CHANGELOG.md)



## Install & Documentation 

- [Install](https://vdbulcke.github.io/hc-vault-util/install/) instruction
- [CLI Doc](./doc/hc-vault-util.md)
- [Documentation](https://vdbulcke.github.io/hc-vault-util/)

### Validate Signature With Cosign

Make sure you have `cosign` installed locally (see [Cosign Install](https://docs.sigstore.dev/cosign/installation/)).


Then you can use the `./verify_signature.sh` in this repo: 

```bash
./verify_signature.sh PATH_TO_DOWNLOADED_ARCHIVE TAG_VERSION
```
for example
```bash
$ ./verify_signature.sh  ~/Downloads/hc-vault-util_0.2.0_Linux_x86_64.tar.gz v0.3.0

Checking Signature for version: v0.3.0
Verified OK

```

## Vault Kv2 Secret TUI

`hc-vault ui` relies on [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) to display and navigate your Vault kv2 secrets in your terminal.

```bash
hc-vault-util ui
```

> NOTE: you must have `VAULT_ADDR` and `VAULT_TOKEN` environment variables 

<img  src=./example/demo.gif width="700"/>

> The above example was generated with VHS ([view source](./example/demo.tape)).

Key Binding

| Key | Action |
|  -- | -- | 
| `k` | Move up the list |
| `j` | Move Down the list |
| `h` | Move to previous page |
| `l` or ENTER | Move to next page  |
| Arrow Keys | navigate in pager |
| / | Trigger fuzzy filter  |
| ? | Help | 
| q | Quit | 
| CTRL+C | Quit |
