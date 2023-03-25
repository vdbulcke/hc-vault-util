# HC Vault Util 

`hc-vault-util` is a companion CLI tool for managing Hashicorp Vault.

## Features

- Vault transit backend import private key using key wrapping 
    - See [transit-import-key Tutorial](https://github.com/vdbulcke/terraform-vault-sample/blob/main/tutorial/transit-import-key/README.md)
- Generate CSR from Vault transit key using [cfssl json csr format](https://github.com/cloudflare/cfssl#signing)
    - See [transit-gencsr Tutorial](https://github.com/vdbulcke/terraform-vault-sample/blob/main/tutorial/transit-gencsr/README.md)

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