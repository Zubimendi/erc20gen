package prompts

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/Zubimendi/erc20gen/internal/config"
)

func CollectTokenConfig() (*config.TokenConfig, error) {
	cfg := &config.TokenConfig{}

	// --- Core identity ---
	var answers struct {
		Name          string
		Symbol        string
		DecimalsStr   string
		InitialSupply string
	}

	if err := survey.Ask([]*survey.Question{
		{
			Name:     "name",
			Prompt:   &survey.Input{Message: "Token Name:", Help: "e.g. MyAwesomeToken"},
			Validate: survey.Required,
		},
		{
			Name:     "symbol",
			Prompt:   &survey.Input{Message: "Token Symbol (uppercase):", Help: "e.g. MTK — max 11 chars"},
			Validate: survey.Required,
		},
		{
			Name: "decimalsStr",
			Prompt: &survey.Select{
				Message: "Decimals:",
				Options: []string{"18", "6", "8", "0"},
				Default: "18",
				Help:    "18 is the Ethereum standard. Use 6 for stablecoins like USDC.",
			},
		},
		{
			Name:   "initialSupply",
			Prompt: &survey.Input{Message: "Initial Supply (whole tokens):", Default: "1000000"},
		},
	}, &answers); err != nil {
		return nil, err
	}

	cfg.Name = answers.Name
	cfg.Symbol = answers.Symbol
	cfg.InitialSupply = answers.InitialSupply

	switch answers.DecimalsStr {
	case "6":
		cfg.Decimals = 6
	case "8":
		cfg.Decimals = 8
	case "0":
		cfg.Decimals = 0
	default:
		cfg.Decimals = 18
	}

	// --- Supply cap ---
	var hasCap bool
	if err := survey.AskOne(
		&survey.Confirm{Message: "Set a maximum supply cap?", Default: false},
		&hasCap,
	); err != nil {
		return nil, err
	}

	if hasCap {
		var cap string
		if err := survey.AskOne(
			&survey.Input{Message: "Maximum Supply (whole tokens):", Default: "10000000"},
			&cap,
		); err != nil {
			return nil, err
		}
		cfg.MaxSupply = cap
	}

	// --- Feature flags ---
	var features []string
	if err := survey.AskOne(&survey.MultiSelect{
		Message: "Select token features:",
		Options: []string{
			"Mintable     — owner can mint new tokens",
			"Burnable     — holders can burn their tokens",
			"Pausable     — owner can pause all transfers",
			"Permit       — EIP-2612 gasless approvals",
			"Snapshot     — balance snapshots for governance",
			"Votes        — on-chain voting power",
		},
		Help: "Space to select, Enter to confirm.",
	}, &features); err != nil {
		return nil, err
	}

	for _, f := range features {
		switch f[:8] {
		case "Mintable":
			cfg.Mintable = true
		case "Burnable":
			cfg.Burnable = true
		case "Pausable":
			cfg.Pausable = true
		case "Permit  ":
			cfg.Permit = true
		case "Snapshot":
			cfg.Snapshot = true
		case "Votes   ":
			cfg.Votes = true
		}
	}

	// --- Access control ---
	var accessStr string
	if err := survey.AskOne(&survey.Select{
		Message: "Access Control Model:",
		Options: []string{"ownable", "roles", "none"},
		Default: "ownable",
		Help:    "ownable = single owner. roles = multi-role with AccessControl. none = no restrictions.",
	}, &accessStr); err != nil {
		return nil, err
	}
	cfg.AccessControl = config.AccessControlType(accessStr)

	// --- Output options ---
	var outputAnswers struct {
		WithDeploy bool
		WithTest   bool
		License    string
	}

	if err := survey.Ask([]*survey.Question{
		{Name: "withDeploy", Prompt: &survey.Confirm{Message: "Generate Hardhat deployment script?", Default: true}},
		{Name: "withTest", Prompt: &survey.Confirm{Message: "Generate Hardhat test skeleton?", Default: true}},
		{Name: "license", Prompt: &survey.Select{
			Message: "License:",
			Options: []string{"MIT", "GPL-3.0", "UNLICENSED", "Apache-2.0"},
			Default: "MIT",
		}},
	}, &outputAnswers); err != nil {
		return nil, err
	}

	cfg.WithDeploy = outputAnswers.WithDeploy
	cfg.WithTest = outputAnswers.WithTest
	cfg.License = outputAnswers.License
	cfg.SolidityVersion = "^0.8.24"

	return cfg, nil
}