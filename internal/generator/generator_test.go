package generator_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/Zubimendi/erc20gen/internal/config"
	"github.com/Zubimendi/erc20gen/internal/generator"
)

// ─── Helper ──────────────────────────────────────────────────────────────────

func baseConfig() *config.TokenConfig {
	return &config.TokenConfig{
		Name:            "TestToken",
		Symbol:          "TST",
		Decimals:        18,
		InitialSupply:   "1000000",
		AccessControl:   config.AccessOwnable,
		License:         "MIT",
		SolidityVersion: "^0.8.24",
	}
}

// ─── Config Validation Tests ──────────────────────────────────────────────────

func TestTokenConfig_Validate_HappyPath(t *testing.T) {
	cfg := baseConfig()
	assert.NoError(t, cfg.Validate())
}

func TestTokenConfig_Validate_EmptyName(t *testing.T) {
	cfg := baseConfig()
	cfg.Name = ""
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token name is required")
}

func TestTokenConfig_Validate_InvalidSymbol(t *testing.T) {
	tests := []struct {
		name   string
		symbol string
	}{
		{"lowercase", "mtk"},
		{"too long", "TOOLONGSYMBOL"},
		{"special chars", "MT K"},
		{"empty", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := baseConfig()
			cfg.Symbol = tt.symbol
			err := cfg.Validate()
			require.Error(t, err)
		})
	}
}

func TestTokenConfig_Validate_ValidSymbols(t *testing.T) {
	valid := []string{"MTK", "USDC", "BTC", "A", "TOKEN123", "T1"}
	for _, s := range valid {
		cfg := baseConfig()
		cfg.Symbol = s
		assert.NoError(t, cfg.Validate(), "symbol %q should be valid", s)
	}
}

func TestTokenConfig_Validate_DecimalsOutOfRange(t *testing.T) {
	cfg := baseConfig()
	cfg.Decimals = 19
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decimals must be between 0 and 18")
}

func TestTokenConfig_Validate_InitialSupplyExceedsCap(t *testing.T) {
	cfg := baseConfig()
	cfg.InitialSupply = "10000000"
	cfg.MaxSupply = "1000000"
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "initial supply cannot exceed max supply")
}

func TestTokenConfig_Validate_InvalidAccessControl(t *testing.T) {
	cfg := baseConfig()
	cfg.AccessControl = "superadmin"
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid access control type")
}

func TestTokenConfig_Validate_VotesAutoEnablesSnapshot(t *testing.T) {
	cfg := baseConfig()
	cfg.Votes = true
	cfg.Snapshot = false
	require.NoError(t, cfg.Validate())
	assert.True(t, cfg.Snapshot, "Votes should auto-enable Snapshot")
}

func TestTokenConfig_Validate_EmptyAccessControlDefaultsToOwnable(t *testing.T) {
	cfg := baseConfig()
	cfg.AccessControl = ""
	require.NoError(t, cfg.Validate())
	assert.Equal(t, config.AccessOwnable, cfg.AccessControl)
}

// ─── ContractFileName Tests ───────────────────────────────────────────────────

func TestTokenConfig_ContractFileName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "MyToken", "MyToken.sol"},
		{"with spaces", "My Token", "My_Token.sol"},
		{"with hyphen", "My-Token", "My_Token.sol"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.TokenConfig{Name: tt.input}
			assert.Equal(t, tt.expected, cfg.ContractFileName())
		})
	}
}

// ─── ImportPaths Tests ────────────────────────────────────────────────────────

func TestTokenConfig_ImportPaths_BaseOnly(t *testing.T) {
	cfg := baseConfig()
	paths := cfg.ImportPaths()
	assert.Contains(t, paths, "@openzeppelin/contracts/token/ERC20/ERC20.sol")
	assert.Contains(t, paths, "@openzeppelin/contracts/access/Ownable.sol")
	assert.Len(t, paths, 2)
}

func TestTokenConfig_ImportPaths_AllFeatures(t *testing.T) {
	cfg := baseConfig()
	cfg.Mintable = true
	cfg.Burnable = true
	cfg.Pausable = true
	cfg.Permit = true
	cfg.Snapshot = true
	cfg.Votes = true
	cfg.MaxSupply = "10000000"

	paths := cfg.ImportPaths()
	assert.Contains(t, paths, "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol")
	assert.Contains(t, paths, "@openzeppelin/contracts/token/ERC20/extensions/ERC20Pausable.sol")
	assert.Contains(t, paths, "@openzeppelin/contracts/token/ERC20/extensions/ERC20Permit.sol")
	assert.Contains(t, paths, "@openzeppelin/contracts/token/ERC20/extensions/ERC20Snapshot.sol")
	assert.Contains(t, paths, "@openzeppelin/contracts/token/ERC20/extensions/ERC20Votes.sol")
	assert.Contains(t, paths, "@openzeppelin/contracts/token/ERC20/extensions/ERC20Capped.sol")
}

// ─── Generator Tests ──────────────────────────────────────────────────────────

func TestGenerator_GenerateContract_ContainsRequiredElements(t *testing.T) {
	cfg := baseConfig()
	require.NoError(t, cfg.Validate())

	gen := generator.New(cfg)
	contract, err := gen.GenerateContract()
	require.NoError(t, err)

	assert.Contains(t, contract, "SPDX-License-Identifier: MIT")
	assert.Contains(t, contract, "pragma solidity ^0.8.24")
	assert.Contains(t, contract, `"TestToken"`)
	assert.Contains(t, contract, `"TST"`)
	assert.Contains(t, contract, "contract TestToken is ERC20")
	assert.Contains(t, contract, "Ownable")
	assert.Contains(t, contract, "_mint(")
	assert.Contains(t, contract, "1000000")
}

func TestGenerator_GenerateContract_MintableIncludesMintFunction(t *testing.T) {
	cfg := baseConfig()
	cfg.Mintable = true
	require.NoError(t, cfg.Validate())

	gen := generator.New(cfg)
	contract, err := gen.GenerateContract()
	require.NoError(t, err)

	assert.Contains(t, contract, "function mint(")
	assert.Contains(t, contract, "onlyOwner")
}

func TestGenerator_GenerateContract_BurnableIncludesBurnableInheritance(t *testing.T) {
	cfg := baseConfig()
	cfg.Burnable = true
	require.NoError(t, cfg.Validate())

	gen := generator.New(cfg)
	contract, err := gen.GenerateContract()
	require.NoError(t, err)

	assert.Contains(t, contract, "ERC20Burnable")
	assert.Contains(t, contract, "ERC20Burnable.sol")
}

func TestGenerator_GenerateContract_PausableIncludesPauseFunctions(t *testing.T) {
	cfg := baseConfig()
	cfg.Pausable = true
	require.NoError(t, cfg.Validate())

	gen := generator.New(cfg)
	contract, err := gen.GenerateContract()
	require.NoError(t, err)

	assert.Contains(t, contract, "function pause()")
	assert.Contains(t, contract, "function unpause()")
	assert.Contains(t, contract, "_pause()")
	assert.Contains(t, contract, "_unpause()")
}

func TestGenerator_GenerateContract_RolesAccessControl(t *testing.T) {
	cfg := baseConfig()
	cfg.AccessControl = config.AccessRoles
	cfg.Mintable = true
	require.NoError(t, cfg.Validate())

	gen := generator.New(cfg)
	contract, err := gen.GenerateContract()
	require.NoError(t, err)

	assert.Contains(t, contract, "MINTER_ROLE")
	assert.Contains(t, contract, "onlyRole(MINTER_ROLE)")
	assert.Contains(t, contract, "AccessControl")
	assert.NotContains(t, contract, "onlyOwner")
}

func TestGenerator_GenerateContract_NoAccessControl(t *testing.T) {
	cfg := baseConfig()
	cfg.AccessControl = config.AccessNone
	cfg.Mintable = true
	require.NoError(t, cfg.Validate())

	gen := generator.New(cfg)
	contract, err := gen.GenerateContract()
	require.NoError(t, err)

	assert.NotContains(t, contract, "onlyOwner")
	assert.NotContains(t, contract, "onlyRole")
	assert.NotContains(t, contract, "Ownable")
}

func TestGenerator_GenerateContract_CustomDecimals(t *testing.T) {
	cfg := baseConfig()
	cfg.Decimals = 6
	require.NoError(t, cfg.Validate())

	gen := generator.New(cfg)
	contract, err := gen.GenerateContract()
	require.NoError(t, err)

	assert.Contains(t, contract, "function decimals()")
	assert.Contains(t, contract, "return 6;")
}

func TestGenerator_GenerateContract_Standard18DecimalsNoOverride(t *testing.T) {
	cfg := baseConfig()
	cfg.Decimals = 18
	require.NoError(t, cfg.Validate())

	gen := generator.New(cfg)
	contract, err := gen.GenerateContract()
	require.NoError(t, err)

	// Standard decimals should NOT generate an override function
	assert.NotContains(t, contract, "function decimals()")
}

func TestGenerator_GenerateContract_WithCap(t *testing.T) {
	cfg := baseConfig()
	cfg.MaxSupply = "10000000"
	require.NoError(t, cfg.Validate())

	gen := generator.New(cfg)
	contract, err := gen.GenerateContract()
	require.NoError(t, err)

	assert.Contains(t, contract, "ERC20Capped")
	assert.Contains(t, contract, "10000000")
}

func TestGenerator_GenerateContract_PermitIncluded(t *testing.T) {
	cfg := baseConfig()
	cfg.Permit = true
	require.NoError(t, cfg.Validate())

	gen := generator.New(cfg)
	contract, err := gen.GenerateContract()
	require.NoError(t, err)

	assert.Contains(t, contract, "ERC20Permit")
	assert.Contains(t, contract, "ERC20Permit.sol")
}

func TestGenerator_GenerateContract_NoInitialSupply(t *testing.T) {
	cfg := baseConfig()
	cfg.InitialSupply = ""
	require.NoError(t, cfg.Validate())

	gen := generator.New(cfg)
	contract, err := gen.GenerateContract()
	require.NoError(t, err)

	assert.NotContains(t, contract, "_mint(")
}

func TestGenerator_GenerateDeployScript_ContainsEssentials(t *testing.T) {
	cfg := baseConfig()
	require.NoError(t, cfg.Validate())

	gen := generator.New(cfg)
	script, err := gen.GenerateDeployScript()
	require.NoError(t, err)

	assert.Contains(t, script, "const { ethers } = require(\"hardhat\")")
	assert.Contains(t, script, "TestToken")
	assert.Contains(t, script, "deployer.address")
	assert.Contains(t, script, "waitForDeployment")
}

func TestGenerator_GenerateTestSkeleton_ContainsEssentials(t *testing.T) {
	cfg := baseConfig()
	require.NoError(t, cfg.Validate())

	gen := generator.New(cfg)
	test, err := gen.GenerateTestSkeleton()
	require.NoError(t, err)

	assert.Contains(t, test, `describe("TestToken"`)
	assert.Contains(t, test, "Should have correct name and symbol")
	assert.Contains(t, test, `"TestToken"`)
	assert.Contains(t, test, `"TST"`)
	assert.True(t, strings.Contains(test, "deployFixture"))
}

func TestGenerator_GenerateTestSkeleton_MintableAddsTests(t *testing.T) {
	cfg := baseConfig()
	cfg.Mintable = true
	require.NoError(t, cfg.Validate())

	gen := generator.New(cfg)
	test, err := gen.GenerateTestSkeleton()
	require.NoError(t, err)

	assert.Contains(t, test, "Minting")
	assert.Contains(t, test, "authorized minting")
	assert.Contains(t, test, "unauthorized minting")
}

func TestGenerator_GenerateTestSkeleton_PausableAddsTests(t *testing.T) {
	cfg := baseConfig()
	cfg.Pausable = true
	require.NoError(t, cfg.Validate())

	gen := generator.New(cfg)
	test, err := gen.GenerateTestSkeleton()
	require.NoError(t, err)

	assert.Contains(t, test, "Pausable")
	assert.Contains(t, test, "paused")
}

// ─── InheritanceList Tests ────────────────────────────────────────────────────

func TestTokenConfig_InheritanceList_OrderMatters(t *testing.T) {
	cfg := baseConfig()
	cfg.Burnable = true
	cfg.Pausable = true
	cfg.Permit = true

	list := cfg.InheritanceList()
	assert.Equal(t, []string{"ERC20Burnable", "ERC20Pausable", "ERC20Permit", "Ownable"}, list)
}

func TestTokenConfig_InheritanceList_CappedFirst(t *testing.T) {
	cfg := baseConfig()
	cfg.MaxSupply = "1000000000"
	cfg.Burnable = true

	list := cfg.InheritanceList()
	assert.Equal(t, "ERC20Capped", list[0], "ERC20Capped should be first in inheritance")
}