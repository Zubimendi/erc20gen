package config

import (
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"unicode"
)

// AccessControlType defines the access control model for the token.
type AccessControlType string

const (
	AccessOwnable AccessControlType = "ownable"
	AccessRoles   AccessControlType = "roles"
	AccessNone    AccessControlType = "none"
)

// TokenConfig holds all parameters for ERC-20 token generation.
type TokenConfig struct {
	// Core ERC-20 fields
	Name            string
	Symbol          string
	Decimals        uint8
	InitialSupply   string // human-readable, e.g. "1000000"
	MaxSupply       string // empty = unlimited

	// Feature flags
	Mintable  bool
	Burnable  bool
	Pausable  bool
	Permit    bool // EIP-2612
	Snapshot  bool
	Votes     bool

	// Access control
	AccessControl AccessControlType

	// Metadata
	License         string
	SolidityVersion string

	// Output options
	WithDeploy bool
	WithTest   bool
}

var (
	validSymbolRe   = regexp.MustCompile(`^[A-Z0-9]{1,11}$`)
	validNameRe     = regexp.MustCompile(`^[A-Za-z0-9 _\-]{1,64}$`)
	validDecimalNum = regexp.MustCompile(`^\d+$`)
)

// Validate performs comprehensive input validation with clear error messages.
func (c *TokenConfig) Validate() error {
	var errs []string

	// Name
	if strings.TrimSpace(c.Name) == "" {
		errs = append(errs, "token name is required")
	} else if !validNameRe.MatchString(c.Name) {
		errs = append(errs, "token name must be 1-64 alphanumeric characters (spaces, hyphens, underscores allowed)")
	}

	// Symbol
	if strings.TrimSpace(c.Symbol) == "" {
		errs = append(errs, "token symbol is required")
	} else if !validSymbolRe.MatchString(strings.ToUpper(c.Symbol)) {
		errs = append(errs, "token symbol must be 1-11 uppercase letters/digits (e.g. MTK, USDC)")
	}

	// Decimals
	if c.Decimals > 18 {
		errs = append(errs, "decimals must be between 0 and 18")
	}

	// Initial supply
	if c.InitialSupply != "" {
		if err := validateSupplyString(c.InitialSupply); err != nil {
			errs = append(errs, fmt.Sprintf("initial supply: %s", err))
		}
	}

	// Max supply
	if c.MaxSupply != "" {
		if err := validateSupplyString(c.MaxSupply); err != nil {
			errs = append(errs, fmt.Sprintf("max supply: %s", err))
		}
		// Ensure max >= initial
		if c.InitialSupply != "" {
			initial, _ := new(big.Int).SetString(c.InitialSupply, 10)
			max, _ := new(big.Int).SetString(c.MaxSupply, 10)
			if initial != nil && max != nil && initial.Cmp(max) > 0 {
				errs = append(errs, "initial supply cannot exceed max supply")
			}
		}
	}

	// Access control
	switch c.AccessControl {
	case AccessOwnable, AccessRoles, AccessNone:
		// valid
	case "":
		c.AccessControl = AccessOwnable
	default:
		errs = append(errs, fmt.Sprintf("invalid access control type %q — must be: ownable, roles, or none", c.AccessControl))
	}

	// Votes requires Snapshot (OpenZeppelin coupling)
	if c.Votes && !c.Snapshot {
		// auto-enable snapshot when votes is on
		c.Snapshot = true
	}

	// License
	if c.License == "" {
		c.License = "MIT"
	}

	// Solidity version
	if c.SolidityVersion == "" {
		c.SolidityVersion = "^0.8.24"
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n  - "))
	}
	return nil
}

func validateSupplyString(s string) error {
	s = strings.TrimSpace(s)
	if !validDecimalNum.MatchString(s) {
		return fmt.Errorf("%q is not a valid positive integer", s)
	}
	n, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return fmt.Errorf("%q cannot be parsed as an integer", s)
	}
	if n.Sign() < 0 {
		return errors.New("must be a non-negative integer")
	}
	return nil
}

// ContractFileName returns the expected Solidity filename.
func (c *TokenConfig) ContractFileName() string {
	return c.SafeName() + ".sol"
}

// SafeName returns a filesystem-safe version of the token name for use in filenames.
func (c *TokenConfig) SafeName() string {
	safe := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return '_'
	}, c.Name)
	return safe
}

// HasAccessControl returns true if any access control is active.
func (c *TokenConfig) HasAccessControl() bool {
	return c.AccessControl != AccessNone
}

// NeedsOwnable returns true if Ownable should be imported.
func (c *TokenConfig) NeedsOwnable() bool {
	return c.AccessControl == AccessOwnable
}

// NeedsRoles returns true if AccessControl (roles) should be imported.
func (c *TokenConfig) NeedsRoles() bool {
	return c.AccessControl == AccessRoles
}

// ImportPaths returns all required OpenZeppelin import paths.
func (c *TokenConfig) ImportPaths() []string {
	var imports []string

	imports = append(imports, "@openzeppelin/contracts/token/ERC20/ERC20.sol")

	if c.Burnable {
		imports = append(imports, "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol")
	}
	if c.Pausable {
		imports = append(imports, "@openzeppelin/contracts/token/ERC20/extensions/ERC20Pausable.sol")
		imports = append(imports, "@openzeppelin/contracts/utils/Pausable.sol")
	}
	if c.Permit {
		imports = append(imports, "@openzeppelin/contracts/token/ERC20/extensions/ERC20Permit.sol")
	}
	if c.Snapshot {
		imports = append(imports, "@openzeppelin/contracts/token/ERC20/extensions/ERC20Snapshot.sol")
	}
	if c.Votes {
		imports = append(imports, "@openzeppelin/contracts/token/ERC20/extensions/ERC20Votes.sol")
	}
	if c.MaxSupply != "" {
		imports = append(imports, "@openzeppelin/contracts/token/ERC20/extensions/ERC20Capped.sol")
	}
	if c.NeedsOwnable() {
		imports = append(imports, "@openzeppelin/contracts/access/Ownable.sol")
	}
	if c.NeedsRoles() {
		imports = append(imports, "@openzeppelin/contracts/access/AccessControl.sol")
	}

	return imports
}

// InheritanceList returns the Solidity inheritance list (excluding base ERC20).
func (c *TokenConfig) InheritanceList() []string {
	var list []string

	if c.MaxSupply != "" {
		list = append(list, "ERC20Capped")
	}
	if c.Burnable {
		list = append(list, "ERC20Burnable")
	}
	if c.Pausable {
		list = append(list, "ERC20Pausable")
	}
	if c.Permit {
		list = append(list, "ERC20Permit")
	}
	if c.Snapshot {
		list = append(list, "ERC20Snapshot")
	}
	if c.Votes {
		list = append(list, "ERC20Votes")
	}
	if c.NeedsOwnable() {
		list = append(list, "Ownable")
	}
	if c.NeedsRoles() {
		list = append(list, "AccessControl")
	}

	return list
}