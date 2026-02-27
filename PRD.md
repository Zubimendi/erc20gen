# PRD: ERC-20 Token Generator CLI (erc20gen)

**Version:** 1.0  
**Author:** [Your Name]  
**Status:** In Development  
**Last Updated:** 2025

---

## 1. Problem Statement

### Background
The Ethereum ecosystem has seen hundreds of millions of dollars lost due to poorly configured or vulnerable ERC-20 token contracts. Common failure modes:

- **No access control on `mint()`** — anyone can mint unlimited tokens (seen in multiple rugpull schemes)
- **Incorrect decimal math** — teams specify supply without accounting for decimals, creating either dust amounts or overflows
- **Missing emergency mechanisms** — no pause, no ownership transfer path, no multi-sig
- **Manual copy-paste workflows** — developers copy contracts from Stack Overflow or old tutorials, inheriting outdated patterns

### Current Landscape
| Tool | Limitations |
|---|---|
| OpenZeppelin Wizard | Browser-only, no CLI, no test generation, no deploy scripts |
| Remix IDE | No automation, manual steps, not CI-friendly |
| Thirdweb | Requires account, paid features, black box |
| Manual | Error-prone, inconsistent, no audit trail |

### Target Users
1. **Solo developers** launching governance tokens, project tokens, or stablecoins
2. **Startups** who need a repeatable, auditable token generation process
3. **Security engineers** who need to quickly spin up test tokens for protocol testing
4. **CTF players** who need specific token configurations for challenges

---

## 2. Goals

### Primary Goals
- Generate production-ready, OpenZeppelin v5-based ERC-20 contracts in < 60 seconds
- Zero paid services required — fully open source, self-hostable
- Security-first output: every contract follows best practices by default
- Cross-platform: Linux, macOS, Windows

### Non-Goals (v1.0)
- GUI or web interface
- On-chain deployment (erc20gen generates files; Hardhat/Foundry handles deployment)
- Vyper support
- ERC-721 / ERC-1155 (separate tools)

---

## 3. User Stories

| ID | As a... | I want to... | So that... |
|---|---|---|---|
| US-1 | Developer | Run `erc20gen generate` interactively | I don't need to memorize all flags |
| US-2 | DevOps engineer | Pass all options as flags | I can automate this in CI/CD |
| US-3 | Security engineer | See a printed security checklist | I don't forget pre-deployment steps |
| US-4 | Solo founder | Generate a capped, mintable token | I can control token supply growth |
| US-5 | Protocol developer | Generate a Votes-enabled token | I can use it for on-chain governance |
| US-6 | Developer | Get a Hardhat test skeleton | I start TDD immediately |
| US-7 | Developer | Get a deploy script | I don't have to write boilerplate |

---

## 4. Functional Requirements

### 4.1 Core Contract Generation
| ID | Requirement | Priority |
|---|---|---|
| FR-1 | Generate valid Solidity ERC-20 contract from config | P0 |
| FR-2 | Support all OpenZeppelin v5 ERC-20 extensions | P0 |
| FR-3 | Output file named `{TokenName}.sol` in specified directory | P0 |
| FR-4 | Include SPDX license identifier | P0 |
| FR-5 | Support configurable Solidity version pragma | P1 |
| FR-6 | Support 0-18 decimals | P0 |
| FR-7 | Apply correct decimal math to initial/max supply | P0 |

### 4.2 Features
| ID | Requirement | Priority |
|---|---|---|
| FR-8 | Mintable with Ownable guard | P0 |
| FR-9 | Mintable with AccessControl/MINTER_ROLE guard | P0 |
| FR-10 | Burnable (ERC20Burnable) | P0 |
| FR-11 | Pausable with authorized pause/unpause | P0 |
| FR-12 | Permit — EIP-2612 (ERC20Permit) | P1 |
| FR-13 | Snapshot — ERC20Snapshot | P1 |
| FR-14 | Votes — ERC20Votes with delegation | P2 |
| FR-15 | Capped supply — ERC20Capped | P1 |

### 4.3 Access Control
| ID | Requirement | Priority |
|---|---|---|
| FR-16 | Ownable model (single owner) | P0 |
| FR-17 | AccessControl model (multi-role) | P0 |
| FR-18 | None model (for testing or permissionless) | P1 |

### 4.4 CLI Interface
| ID | Requirement | Priority |
|---|---|---|
| FR-19 | Interactive prompt mode (default) | P0 |
| FR-20 | Full flag-based non-interactive mode | P0 |
| FR-21 | `--help` on all commands | P0 |
| FR-22 | `version` subcommand | P1 |
| FR-23 | Config file support (YAML) | P2 |

### 4.5 Output Files
| ID | Requirement | Priority |
|---|---|---|
| FR-24 | Hardhat deployment script (JS) with Etherscan verification | P1 |
| FR-25 | Hardhat test skeleton with 15+ test cases | P1 |
| FR-26 | Security checklist printed to stdout | P0 |

---

## 5. Non-Functional Requirements

| Category | Requirement |
|---|---|
| Security | All generated contracts use OpenZeppelin v5. No external calls in generated code that could introduce reentrancy. |
| Performance | Contract generation must complete in < 500ms |
| Testing | Unit test coverage ≥ 80%. All templates tested via generator_test.go |
| Portability | Binary under 20MB. Statically linked. No runtime dependencies |
| Compatibility | Solidity ^0.8.24, OpenZeppelin 5.x, Hardhat 2.x |
| Auditability | Generated contracts include a comment indicating they were tool-generated and must be reviewed |

---

## 6. Security Requirements

- No network calls at runtime — fully offline capable
- File permissions: output files written with `0640` (owner r/w, group r, no world access)
- Directory creation uses `0750`
- Input validation: all user inputs sanitized before template injection
- No shell injection possible — Go's `text/template` is not a shell

---

## 7. Technical Architecture

```
erc20gen/
├── main.go                          # Entry point
├── cmd/
│   ├── root.go                      # Root cobra command + global flags
│   ├── generate.go                  # `generate` subcommand
│   └── version.go                   # `version` subcommand
├── internal/
│   ├── config/
│   │   └── config.go                # TokenConfig model + validation
│   ├── generator/
│   │   ├── generator.go             # Template rendering engine
│   │   └── generator_test.go        # TDD unit tests
│   └── prompts/
│       └── prompts.go               # Survey-based interactive prompts
├── .github/workflows/ci.yml         # CI: test, lint, build, security scan
├── .golangci.yml                    # Linter config
├── Makefile                         # Developer tasks
└── README.md
```

### Dependencies
| Package | Purpose |
|---|---|
| `spf13/cobra` | CLI framework |
| `spf13/viper` | Config management |
| `AlecAivazis/survey/v2` | Interactive prompts |
| `stretchr/testify` | Test assertions |

---

## 8. Test Plan

### Unit Tests (Go)
- `config.Validate()` — 15+ cases covering all validation branches
- `generator.GenerateContract()` — all feature combinations
- `generator.GenerateDeployScript()` — output correctness
- `generator.GenerateTestSkeleton()` — conditional test sections
- `TokenConfig.ImportPaths()` — correct OZ imports per feature
- `TokenConfig.InheritanceList()` — correct ordering

### Integration Tests (Hardhat — manual)
```bash
# After generating a contract:
cd my-hardhat-project
npm install
npx hardhat compile           # Must succeed with 0 errors
npx hardhat test              # Must pass all generated tests
```

### Security Tests
```bash
slither contracts/MyToken.sol # Must return 0 critical findings
govulncheck ./...              # Must return 0 vulnerabilities in Go code
```

---

## 9. Launch Checklist

- [ ] All unit tests passing (`make test`)
- [ ] Coverage ≥ 80% (`make coverage`)
- [ ] Lint clean (`make lint`)
- [ ] Security scan clean (`make sec`, `make vuln`)
- [ ] README complete with examples
- [ ] Generated contract tested against Hardhat compile + test
- [ ] Cross-platform binaries built (`make release`)
- [ ] GitHub Release created with binaries attached
- [ ] LinkedIn post published
- [ ] Twitter/X thread published
- [ ] Medium article published

---

## 10. Social Content

### Twitter/X Thread

```
🧵 I just built an ERC-20 Token Generator CLI in Go — and it generates 
production-ready, security-audited smart contracts in < 60 seconds.

Here's why I built it and what makes it different: 👇

1/ The problem: Every week, developers deploy ERC-20 tokens with critical bugs.
   - No access control on mint()
   - Wrong decimal math (permanent fund loss)
   - No emergency pause
   
   Copy-paste from tutorials = copy-paste the vulnerabilities.

2/ Existing tools (OpenZeppelin Wizard, Remix) are browser-based.
   No CLI. No deploy scripts. No test generation. Not CI-friendly.

3/ erc20gen solves this from your terminal:
   
   erc20gen generate \
     --name "GovToken" \
     --symbol "GOV" \
     --mintable \
     --pausable \
     --access roles \
     --with-deploy \
     --with-test

4/ Every generated contract includes:
   ✅ OpenZeppelin v5 (industry standard)
   ✅ Access control (Ownable or multi-role)
   ✅ Correct decimal math
   ✅ Hardhat deploy script + test skeleton
   ✅ Pre-deployment security checklist

5/ Built with:
   - Go + Cobra (CLI)
   - survey/v2 (interactive prompts)
   - text/template (contract generation)
   - TDD throughout (testify)
   - GitHub Actions CI with gosec + govulncheck

6/ Supports:
   🔐 Mintable, Burnable, Pausable
   ✍️ EIP-2612 Permit (gasless approvals)
   📸 Snapshots
   🗳️ On-chain Votes
   🪙 Supply caps

7/ Repo: github.com/yourhandle/erc20gen
   
   Star ⭐ if you found this useful.
   
   Week 1 of my 12-week portfolio sprint.
   Next: On-chain Transaction Monitor + Alert System.
   
   Follow along 🚀

#Go #Golang #Ethereum #Solidity #Web3 #ERC20 #OpenSource #DevTools
```

### LinkedIn Post

```
I spent this week building an ERC-20 Token Generator CLI in Go — and I 
want to show you why this tool matters and what I learned building it.

THE PROBLEM
Hundreds of millions of dollars have been lost to poorly configured ERC-20 
tokens. Unrestricted mint() functions, incorrect decimal handling, no 
emergency pause. Teams copy-paste contracts from tutorials and inherit 
every vulnerability in them.

THE SOLUTION: erc20gen
A command-line tool that generates production-ready, security-audited 
ERC-20 smart contracts from a single command.

It supports:
🔐 Mintable with proper access control (Ownable or Role-based)
🔥 Burnable tokens
⏸️ Emergency pausable transfers
✍️ EIP-2612 Permit for gasless approvals
📸 Snapshot for governance
🗳️ On-chain Votes (EIP-5805)
🪙 Hard supply caps

And generates:
📄 The Solidity contract (OpenZeppelin v5)
🚀 A Hardhat deployment script
🧪 A full test skeleton with security edge cases
✅ A printed security checklist for pre-deployment

WHAT I LEARNED
• Go's text/template is incredibly powerful for code generation
• TDD made refactoring the template logic painless
• CLI UX matters — interactive mode vs flags for different use cases
• Security in developer tools is often an afterthought — it shouldn't be

This is Week 1 of my 12-week open source portfolio sprint.
Goal: 70 production-ready repositories across blockchain, DevOps, and 
cybersecurity — fully documented and publicly available.

GitHub: github.com/yourhandle/erc20gen

What would you add to this tool? Drop it in the comments 👇

#Golang #Ethereum #Solidity #Web3 #OpenSource #DevTools #CyberSecurity 
#SoftwareEngineering #PortfolioProject
```

### Medium Article Outline

**Title:** "Building a Production-Safe ERC-20 Token Generator CLI in Go"

1. Introduction — the hidden danger of copy-paste Solidity
2. Real-world token vulnerabilities (with examples)
3. Tool design decisions (Cobra vs urfave/cli, template engine choice)
4. Architecture walkthrough (config → validation → template rendering)
5. TDD approach — writing tests before templates
6. The templates explained (each OZ extension and why)
7. Security checklist — what to audit post-generation
8. CI/CD pipeline — gosec, govulncheck, cross-compilation
9. Live demo + output walkthrough
10. What's next — roadmap
```
