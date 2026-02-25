package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	appName    = "erc20gen"
	appVersion = "1.0.0"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   appName,
	Short: "ERC-20 Token Generator CLI — production-ready smart contracts in seconds",
	Long: `
███████╗██████╗  ██████╗    ██████╗  ██████╗  ██████╗ ███████╗███╗   ██╗
██╔════╝██╔══██╗██╔════╝    ╚════██╗██╔═████╗██╔════╝ ██╔════╝████╗  ██║
█████╗  ██████╔╝██║          █████╔╝██║██╔██║██║  ███╗█████╗  ██╔██╗ ██║
██╔══╝  ██╔══██╗██║         ██╔═══╝ ████╔╝██║██║   ██║██╔══╝  ██║╚██╗██║
███████╗██║  ██║╚██████╗    ███████╗╚██████╔╝╚██████╔╝███████╗██║ ╚████║
╚══════╝╚═╝  ╚═╝ ╚═════╝    ╚══════╝ ╚═════╝  ╚═════╝ ╚══════╝╚═╝  ╚═══╝

Generate production-ready, security-audited ERC-20 smart contracts.
Supports OpenZeppelin patterns: Ownable, Roles, Mintable, Burnable,
Pausable, Permit (EIP-2612), Snapshots, and Capped supply.

Built with security-first principles. No paid services required.
`,
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $HOME/.erc20gen.yaml)")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable colored output")
	_ = viper.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".erc20gen")
	}
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()
}