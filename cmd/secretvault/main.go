package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/matteospanio/secret-vault-cli/pkg/vault"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var (
	vaultPath string
	password  string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "secretvault",
	Short: "Secure token and secret storage",
	Long: `Secret Vault CLI - A command-line application for securely storing 
and retrieving API tokens and credentials with encryption.`,
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new vault",
	Long:  `Create a new encrypted vault with a master password.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleInit()
	},
}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add or update a secret",
	Long:  `Add a new secret or update an existing one in the vault.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleAdd(args[0])
	},
}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Retrieve a secret value",
	Long:  `Retrieve and print a secret value to stdout (suitable for piping).`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleGet(args[0])
	},
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all secret names",
	Long:  `Display all secrets in the vault with metadata in a formatted table.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleList()
	},
}

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a secret",
	Long:  `Delete a secret from the vault.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleRemove(args[0])
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&vaultPath, "vault-path", "", "path to vault file (default: ~/.secret-vault/vault.enc)")
	rootCmd.PersistentFlags().StringVar(&password, "password", "", "master password (if not set, will prompt)")

	// Bind flags to viper
	viper.BindPFlag("vault-path", rootCmd.PersistentFlags().Lookup("vault-path"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))

	// Add commands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(removeCmd)
}

func initConfig() {
	// Environment variables
	viper.SetEnvPrefix("vault")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("vault-path", "")
	viper.SetDefault("password", "")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getVaultPath() (string, error) {
	// Check flag/config first
	if path := viper.GetString("vault-path"); path != "" {
		return path, nil
	}
	// Check environment variable
	if path := os.Getenv("VAULT_PATH"); path != "" {
		return path, nil
	}
	// Use default
	return vault.GetDefaultVaultPath()
}

func getPassword(prompt string) (string, error) {
	// Check flag/config first
	if pwd := viper.GetString("password"); pwd != "" {
		return pwd, nil
	}
	// Check environment variable
	if pwd := os.Getenv("VAULT_PASSWORD"); pwd != "" {
		return pwd, nil
	}

	// Prompt user
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
		fmt.Println("Error: failed to read password")
		os.Exit(1)
	}

	if password == "" {
		fmt.Println("Error: password cannot be empty")
		os.Exit(1)
	}

	confirmPassword, err := getPassword("Confirm master password: ")
	if err != nil {
		fmt.Println("Error: failed to read password")
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
		fmt.Println("Error: failed to read password")
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
	reader := bufio.NewReader(os.Stdin)
	description, err := reader.ReadString('\n')
	if err != nil {
		description = ""
	}
	description = strings.TrimSpace(description)

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
		fmt.Println("Error: failed to read password")
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
		fmt.Println("Error: failed to read password")
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
		fmt.Println("Error: failed to read password")
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
