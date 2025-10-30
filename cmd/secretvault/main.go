package main

import (
	"fmt"
	"os"
	"sort"
	"syscall"
	"text/tabwriter"

	"github.com/matteospanio/secret-vault-cli/pkg/vault"
	"golang.org/x/term"
)

const usage = `Secret Vault CLI - Secure token and secret storage

Usage:
  secretvault <command> [options]

Commands:
  init              Initialize a new vault
  add <name>        Add or update a secret
  get <name>        Retrieve a secret value
  list              List all secret names
  remove <name>     Remove a secret
  help              Show this help message

Examples:
  secretvault init
  secretvault add github-token
  secretvault get github-token
  secretvault list
  secretvault remove github-token

Environment Variables:
  VAULT_PASSWORD    Master password (if not set, will prompt)
  VAULT_PATH        Path to vault file (default: ~/.secret-vault/vault.enc)
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		handleInit()
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Error: secret name required")
			fmt.Println("Usage: secretvault add <name>")
			os.Exit(1)
		}
		handleAdd(os.Args[2])
	case "get":
		if len(os.Args) < 3 {
			fmt.Println("Error: secret name required")
			fmt.Println("Usage: secretvault get <name>")
			os.Exit(1)
		}
		handleGet(os.Args[2])
	case "list":
		handleList()
	case "remove":
		if len(os.Args) < 3 {
			fmt.Println("Error: secret name required")
			fmt.Println("Usage: secretvault remove <name>")
			os.Exit(1)
		}
		handleRemove(os.Args[2])
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Printf("Error: unknown command '%s'\n\n", command)
		fmt.Print(usage)
		os.Exit(1)
	}
}

func getVaultPath() (string, error) {
	if path := os.Getenv("VAULT_PATH"); path != "" {
		return path, nil
	}
	return vault.GetDefaultVaultPath()
}

func getPassword(prompt string) (string, error) {
	if password := os.Getenv("VAULT_PASSWORD"); password != "" {
		return password, nil
	}

	fmt.Print(prompt)
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}

	return string(passwordBytes), nil
}

func handleInit() {
	vaultPath, err := getVaultPath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if vault.VaultExists(vaultPath) {
		fmt.Printf("Error: vault already exists at %s\n", vaultPath)
		fmt.Println("Use 'add', 'get', 'list', or 'remove' to manage secrets")
		os.Exit(1)
	}

	password, err := getPassword("Enter master password: ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if password == "" {
		fmt.Println("Error: password cannot be empty")
		os.Exit(1)
	}

	confirmPassword, err := getPassword("Confirm master password: ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if password != confirmPassword {
		fmt.Println("Error: passwords do not match")
		os.Exit(1)
	}

	v := vault.NewVault()
	err = vault.SaveVault(v, vaultPath, password)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Vault initialized at %s\n", vaultPath)
}

func handleAdd(name string) {
	vaultPath, err := getVaultPath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !vault.VaultExists(vaultPath) {
		fmt.Println("Error: vault not initialized. Run 'secretvault init' first")
		os.Exit(1)
	}

	password, err := getPassword("Enter master password: ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	v, err := vault.LoadVault(vaultPath, password)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print("Enter secret value: ")
	valueBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	value := string(valueBytes)
	if value == "" {
		fmt.Println("Error: secret value cannot be empty")
		os.Exit(1)
	}

	fmt.Print("Enter description (optional): ")
	var description string
	fmt.Scanln(&description)

	v.AddSecret(name, value, description)

	err = vault.SaveVault(v, vaultPath, password)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Secret '%s' added successfully\n", name)
}

func handleGet(name string) {
	vaultPath, err := getVaultPath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !vault.VaultExists(vaultPath) {
		fmt.Println("Error: vault not initialized. Run 'secretvault init' first")
		os.Exit(1)
	}

	password, err := getPassword("Enter master password: ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	v, err := vault.LoadVault(vaultPath, password)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	secret, err := v.GetSecret(name)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(secret.Value)
}

func handleList() {
	vaultPath, err := getVaultPath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !vault.VaultExists(vaultPath) {
		fmt.Println("Error: vault not initialized. Run 'secretvault init' first")
		os.Exit(1)
	}

	password, err := getPassword("Enter master password: ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	v, err := vault.LoadVault(vaultPath, password)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	names := v.ListSecrets()
	if len(names) == 0 {
		fmt.Println("No secrets stored in vault")
		return
	}

	sort.Strings(names)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tDESCRIPTION\tCREATED\tUPDATED")
	fmt.Fprintln(w, "----\t-----------\t-------\t-------")

	for _, name := range names {
		secret, _ := v.GetSecret(name)
		description := secret.Description
		if description == "" {
			description = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			name,
			description,
			secret.CreatedAt.Format("2006-01-02"),
			secret.UpdatedAt.Format("2006-01-02"),
		)
	}
	w.Flush()
}

func handleRemove(name string) {
	vaultPath, err := getVaultPath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !vault.VaultExists(vaultPath) {
		fmt.Println("Error: vault not initialized. Run 'secretvault init' first")
		os.Exit(1)
	}

	password, err := getPassword("Enter master password: ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	v, err := vault.LoadVault(vaultPath, password)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	err = v.RemoveSecret(name)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	err = vault.SaveVault(v, vaultPath, password)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Secret '%s' removed successfully\n", name)
}
