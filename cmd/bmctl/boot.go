package main

import (
	"github.com/GSI-HPC/bmctl/pkg/bmc"
	"github.com/spf13/cobra"
)

func newBootCmd(bmcClientConfig *bmc.ClientConfig) *cobra.Command {
	cmd := cobra.Command{
		Use:   "boot IMAGE",
		Short: "Boot an image file",
		Long: `Out-of-band initiated device boot

1. Create SSH tunnel, if SSH proxy is set (expects OpenSSH ssh command in $PATH)
2. Connect to and authenticate with BMC
3. Start HTTPS server serving given image file
4. Insert image URL as virtual medium
5. Set next boot target to this virtual medium
6. Reboot the device
7. Wait until Ctrl+C, after which the HTTPS server is shut down

Limitations (currently):
* Works for the first device per BMC only
* local SSH Proxy Port hardcoded to 5555
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			img := args[0]

			client, err := bmc.NewClient(ctx, *bmcClientConfig)
			if err != nil {
				return err
			}
			defer client.Close()

			return client.Boot(ctx, img)
		},
	}

	addBmcClientConfigFlags(&cmd, bmcClientConfig)

	return &cmd
}
