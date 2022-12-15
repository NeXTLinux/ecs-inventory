package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate Completion script",
	Long: `To load completions:

Bash:

$ source <(ecg completion bash)

# To load completions for each session, execute once:
Linux:
  $ ecg completion bash > /etc/bash_completion.d/ecg
MacOS:
  $ ecg completion bash > /usr/local/etc/bash_completion.d/ecg

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ ecg completion zsh > "${fpath[1]}/_ecg"

# You will need to start a new shell for this setup to take effect.

Fish:

$ ecg completion fish | source

# To load completions for each session, execute once:
$ ecg completion fish > ~/.config/fish/completions/ecg.fish
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		switch args[0] {
		case "bash":
			err = cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			err = cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			err = cmd.Root().GenFishCompletion(os.Stdout, true)
		}
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}