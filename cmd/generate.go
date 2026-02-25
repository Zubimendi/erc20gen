package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/Zubimendi/erc20gen/internal/config"
	"github.com/Zubimendi/erc20gen/internal/generator"
	"github.com/Zubimendi/erc20gen/internal/prompts"
)

var generateCmd = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"gen", "new"},
	Short:   "Generate a new ERC-20 token smart contract",
	Long: `Interactively generate a production-ready ERC-20 Solidity token contract.

Examples:
  # Interactive mode (recommended)
  erc20gen generate

  # Non-interactive with flags
  erc20gen generate \
    --name "MyToken" \
    --symbol "MTK" \
    --decimals 18 \
    --initial-supply 1000000 \
    --mintable \
    --burnable \
    --pausable \
    --out ./contracts

  # From a config file
  erc20gen generate --config token.yaml`,
	RunE: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	f := generateCmd.Flags()
	f.String("name", "", "Token name (e.g. MyToken)")
	f.String("symbol", "", "Token symbol (e.g. MTK)")
	f.Uint8("decimals", 18, "Number of decimals (0-18)")
	f.String("initial-supply", "", "Initial supply (in whole tokens, e.g. 1000000)")
	f.String("max-supply", "", "Maximum supply cap (leave empty for unlimited)")
	f.Bool("mintable", false, "Allow minting new tokens after deployment")
	f.Bool("burnable", false, "Allow token holders to burn their tokens")
	f.Bool("pausable", false, "Allow owner to pause all token transfers")
	f.Bool("permit", false, "Add EIP-2612 permit() for gasless approvals")
	f.Bool("snapshot", false, "Add snapshot capability for governance")
	f.Bool("votes", false, "Add ERC-20 Votes for on-chain governance")
	f.String("access", "ownable", "Access control: ownable | roles | none")
	f.String("license", "MIT", "SPDX license identifier")
	f.String("solidity-version", "^0.8.24", "Solidity compiler version pragma")
	f.String("out", "./contracts", "Output directory for generated files")
	f.Bool("with-deploy", false, "Also generate a Hardhat deployment script")
	f.Bool("with-test", false, "Also generate a Hardhat test file skeleton")
	f.Bool("interactive", true, "Use interactive prompts (disable with --interactive=false)")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	var cfg *config.TokenConfig
	var err error

	interactive, _ := cmd.Flags().GetBool("interactive")
	nameFlag, _ := cmd.Flags().GetString("name")

	// If no name flag is provided and interactive mode is on, use prompts
	if interactive && nameFlag == "" {
		cfg, err = prompts.CollectTokenConfig()
		if err != nil {
			return fmt.Errorf("prompt error: %w", err)
		}
	} else {
		// Build config from flags
		cfg, err = buildConfigFromFlags(cmd)
		if err != nil {
			return err
		}
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Generate
	outDir, _ := cmd.Flags().GetString("out")
	if err := os.MkdirAll(outDir, 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	gen := generator.New(cfg)

	// Write contract
	contractPath := filepath.Join(outDir, cfg.ContractFileName())
	contract, err := gen.GenerateContract()
	if err != nil {
		return fmt.Errorf("contract generation failed: %w", err)
	}
	if err := os.WriteFile(contractPath, []byte(contract), 0640); err != nil {
		return fmt.Errorf("failed to write contract: %w", err)
	}
	fmt.Printf("✅ Contract generated: %s\n", contractPath)

	// Optional deploy script
	withDeploy, _ := cmd.Flags().GetBool("with-deploy")
	if cfg.WithDeploy || withDeploy {
		deployPath := filepath.Join(outDir, "..", "scripts", "deploy_"+cfg.SafeName()+".js")
		_ = os.MkdirAll(filepath.Dir(deployPath), 0750)
		deploy, err := gen.GenerateDeployScript()
		if err != nil {
			return fmt.Errorf("deploy script generation failed: %w", err)
		}
		if err := os.WriteFile(deployPath, []byte(deploy), 0640); err != nil {
			return fmt.Errorf("failed to write deploy script: %w", err)
		}
		fmt.Printf("✅ Deploy script generated: %s\n", deployPath)
	}

	// Optional test skeleton
	withTest, _ := cmd.Flags().GetBool("with-test")
	if cfg.WithTest || withTest {
		testPath := filepath.Join(outDir, "..", "test", cfg.SafeName()+".test.js")
		_ = os.MkdirAll(filepath.Dir(testPath), 0750)
		test, err := gen.GenerateTestSkeleton()
		if err != nil {
			return fmt.Errorf("test skeleton generation failed: %w", err)
		}
		if err := os.WriteFile(testPath, []byte(test), 0640); err != nil {
			return fmt.Errorf("failed to write test skeleton: %w", err)
		}
		fmt.Printf("✅ Test skeleton generated: %s\n", testPath)
	}

	fmt.Printf("\n🔐 Security checklist printed to stdout:\n")
	printSecurityChecklist(cfg)
	return nil
}

func buildConfigFromFlags(cmd *cobra.Command) (*config.TokenConfig, error) {
	name, _ := cmd.Flags().GetString("name")
	symbol, _ := cmd.Flags().GetString("symbol")
	decimals, _ := cmd.Flags().GetUint8("decimals")
	initialSupply, _ := cmd.Flags().GetString("initial-supply")
	maxSupply, _ := cmd.Flags().GetString("max-supply")
	mintable, _ := cmd.Flags().GetBool("mintable")
	burnable, _ := cmd.Flags().GetBool("burnable")
	pausable, _ := cmd.Flags().GetBool("pausable")
	permit, _ := cmd.Flags().GetBool("permit")
	snapshot, _ := cmd.Flags().GetBool("snapshot")
	votes, _ := cmd.Flags().GetBool("votes")
	access, _ := cmd.Flags().GetString("access")
	license, _ := cmd.Flags().GetString("license")
	solidityVersion, _ := cmd.Flags().GetString("solidity-version")
	withDeploy, _ := cmd.Flags().GetBool("with-deploy")
	withTest, _ := cmd.Flags().GetBool("with-test")

	return &config.TokenConfig{
		Name:            name,
		Symbol:          symbol,
		Decimals:        decimals,
		InitialSupply:   initialSupply,
		MaxSupply:       maxSupply,
		Mintable:        mintable,
		Burnable:        burnable,
		Pausable:        pausable,
		Permit:          permit,
		Snapshot:        snapshot,
		Votes:           votes,
		AccessControl:   config.AccessControlType(access),
		License:         license,
		SolidityVersion: solidityVersion,
		WithDeploy:      withDeploy,
		WithTest:        withTest,
	}, nil
}

func printSecurityChecklist(cfg *config.TokenConfig) {
	checks := []string{
		"[ ] Review OpenZeppelin version in package.json — use latest stable",
		"[ ] Audit mint() access control before mainnet deployment",
		"[ ] Run Slither static analysis: slither contracts/" + cfg.ContractFileName(),
		"[ ] Run Echidna fuzzer on token invariants",
		"[ ] Verify initial supply is correct (decimals applied in contract)",
		"[ ] Consider front-running risks if using Pausable",
		"[ ] Test all edge cases: zero transfers, max uint256 approvals",
	}
	if cfg.Permit {
		checks = append(checks, "[ ] Validate EIP-712 domain separator is network-specific")
	}
	if cfg.Snapshot {
		checks = append(checks, "[ ] Snapshot IDs should not be guessable — avoid sequential abuse")
	}
	if cfg.Votes {
		checks = append(checks, "[ ] Governance voting delay and quorum must be reviewed carefully")
	}
	for _, c := range checks {
		fmt.Println(" ", c)
	}
}