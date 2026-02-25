# ERC-20 Token Generator CLI

> **Generate production-ready, security-audited ERC-20 smart contracts in seconds.**  
> Built with Go. No browser, no paid services, no boilerplate.

[![CI](https://github.com/Zubimendi/erc20gen/actions/workflows/ci.yml/badge.svg)](https://github.com/Zubimendi/erc20gen/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/Zubimendi/erc20gen)](https://goreportcard.com/report/github.com/Zubimendi/erc20gen)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## The Problem

Every week, developers deploy ERC-20 tokens with critical vulnerabilities:

- Unrestricted `mint()` functions (no access control)
- Missing overflow protection
- Incorrect decimal handling causing permanent fund loss
- No emergency pause mechanism
- Long-lived admin keys with no transfer path

Existing tools (Remix IDE, OpenZeppelin Wizard) are browser-based, require copy-pasting, and produce no deployment or test scaffolding. Teams waste hours wiring up the same patterns over and over.

**erc20gen** solves this from your terminal — reproducible, versionable, CI-ready.

---

## Features

| Feature                 | Description                                                  |
| ----------------------- | ------------------------------------------------------------ |
| 🎯 Interactive mode     | Survey-driven prompts guide you through every option         |
| 🏗️ Non-interactive mode | Full flag support for scripting and CI                       |
| 🔐 Mintable             | `onlyOwner` or `MINTER_ROLE` guarded `mint()`                |
| 🔥 Burnable             | Holders can burn their own tokens                            |
| ⏸️ Pausable             | Emergency pause with access-controlled `pause()`/`unpause()` |
| ✍️ Permit (EIP-2612)    | Gasless approvals via off-chain signatures                   |
| 📸 Snapshot             | Balance snapshots for governance voting                      |
| 🗳️ Votes                | On-chain voting delegation (EIP-5805)                        |
| 🔒 Access Control       | `Ownable` or `AccessControl` (roles) or `none`               |
| 🪙 Capped Supply        | Hard supply cap via `ERC20Capped`                            |
| 📜 Deploy Script        | Hardhat deployment JS, Etherscan verification ready          |
| 🧪 Test Skeleton        | Full Hardhat test suite with security edge cases             |
| 🔍 Security Checklist   | Printed to stdout after every generation                     |
| 🌍 Cross-platform       | Binaries for Linux, macOS, Windows                           |

---

## Installation

### From source (requires Go 1.22+)

```bash
git clone https://github.com/Zubimendi/erc20gen.git
cd erc20gen
make install
```

### Download binary

```bash
# Linux amd64
curl -L https://github.com/Zubimendi/erc20gen/releases/latest/download/erc20gen-linux-amd64 \
  -o /usr/local/bin/erc20gen && chmod +x /usr/local/bin/erc20gen
```

---

## Usage

### Interactive mode (recommended)

```bash
erc20gen generate
```

You'll be walked through:

1. Token name & symbol
2. Decimals (18, 6, 8, or 0)
3. Initial supply
4. Optional max supply cap
5. Feature selection (Mintable, Burnable, Pausable, Permit, Snapshot, Votes)
6. Access control model
7. Output options (deploy script, test skeleton)

### Non-interactive mode

```bash
erc20gen generate \
  --name "GovToken" \
  --symbol "GOV" \
  --decimals 18 \
  --initial-supply 100000000 \
  --max-supply 1000000000 \
  --mintable \
  --burnable \
  --pausable \
  --permit \
  --snapshot \
  --votes \
  --access roles \
  --with-deploy \
  --with-test \
  --out ./contracts
```

### Output structure

```
contracts/
└── GovToken.sol          # Production-ready Solidity contract

scripts/
└── deploy_GovToken.js    # Hardhat deployment script

test/
└── GovToken.test.js      # Hardhat test suite with 15+ test cases
```

---

## Example Output

Running `erc20gen generate --name "StableToken" --symbol "STB" --mintable --pausable --access roles` generates:

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Pausable.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";

contract StableToken is ERC20, ERC20Pausable, AccessControl {
    bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE");
    bytes32 public constant PAUSER_ROLE = keccak256("PAUSER_ROLE");

    constructor(address defaultAdmin) ERC20("StableToken", "STB") {
        _grantRole(DEFAULT_ADMIN_ROLE, defaultAdmin);
        _grantRole(MINTER_ROLE, defaultAdmin);
        _grantRole(PAUSER_ROLE, defaultAdmin);
    }

    function mint(address to, uint256 amount) external onlyRole(MINTER_ROLE) {
        _mint(to, amount);
    }

    function pause() external onlyRole(PAUSER_ROLE) { _pause(); }
    function unpause() external onlyRole(PAUSER_ROLE) { _unpause(); }

    function _update(address from, address to, uint256 value)
        internal override(ERC20, ERC20Pausable) {
        super._update(from, to, value);
    }
}
```

---

## Security Model

erc20gen is built by a security-first engineer. Every generated contract:

- Uses **OpenZeppelin v5** contracts — the industry standard
- Applies **principle of least privilege** via access control
- Avoids **tx.origin** authentication
- Correctly handles **decimal precision** (no integer truncation bugs)
- Generates a **security checklist** for pre-deployment review
- Recommends **Slither** and **Echidna** for post-generation auditing

### Recommended audit workflow

```bash
# 1. Generate contract
erc20gen generate --name "MyToken" --symbol "MTK" --mintable --out ./contracts

# 2. Static analysis
pip install slither-analyzer
slither contracts/MyToken.sol

# 3. Fuzzing
echidna-test . --contract MyToken --config echidna.yaml

# 4. Manual review
# Check the security checklist printed by erc20gen
```

---

## Development

### Prerequisites

- Go 1.22+
- golangci-lint (`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)
- govulncheck (`go install golang.org/x/vuln/cmd/govulncheck@latest`)

### Commands

```bash
make test        # Run tests with race detector
make lint        # golangci-lint
make sec         # gosec security scan
make vuln        # vulnerability check
make coverage    # HTML coverage report
make build       # Build binary
make release     # Cross-compile all platforms
```

### Test coverage target: 80%+

---

## Contributing

1. Fork & clone
2. `git checkout -b feat/your-feature`
3. Write tests first (TDD)
4. `make test && make lint`
5. Submit PR with clear description

---

## License

MIT © 2025 Zubimendi

---

## Roadmap

- [ ] Foundry deployment script support
- [ ] ERC-4626 Vault token extension
- [ ] Multi-chain deployment config (Hardhat Networks)
- [ ] Vyper contract generation
- [ ] `erc20gen audit` subcommand — run Slither automatically
- [ ] Token config as YAML (`erc20gen generate --config token.yaml`)

---

## Author

Built as part of a 12-week portfolio sprint focused on security-first engineering.

Follow the build: [Twitter](#) | [LinkedIn](#) | [Medium](#)
