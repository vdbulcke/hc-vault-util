## hc-vault-util completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(hc-vault-util completion bash)

To load completions for every new session, execute once:

#### Linux:

	hc-vault-util completion bash > /etc/bash_completion.d/hc-vault-util

#### macOS:

	hc-vault-util completion bash > $(brew --prefix)/etc/bash_completion.d/hc-vault-util

You will need to start a new shell for this setup to take effect.


```
hc-vault-util completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
  -d, --debug      debug mode enabled
      --no-color   disable color output
```

### SEE ALSO

* [hc-vault-util completion](hc-vault-util_completion.md)	 - Generate the autocompletion script for the specified shell

###### Auto generated by spf13/cobra on 31-Oct-2022
