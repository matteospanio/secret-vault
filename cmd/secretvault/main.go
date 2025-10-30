package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/matteospanio/secret-vault-cli/pkg/export"
	"github.com/matteospanio/secret-vault-cli/pkg/git"
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

// gitCmd represents the git command group
var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "Git workflow integration commands",
	Long:  `Commands for integrating Secret Vault CLI with Git workflows via hooks.`,
}

// gitInstallHooksCmd represents the git install-hooks command
var gitInstallHooksCmd = &cobra.Command{
	Use:   "install-hooks",
	Short: "Install Git hooks for Secret Vault integration",
	Long: `Install Git hooks that integrate Secret Vault CLI with Git operations.
The hooks will notify you about token opportunities when pushing to remotes.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleGitInstallHooks()
	},
}

// gitUninstallHooksCmd represents the git uninstall-hooks command
var gitUninstallHooksCmd = &cobra.Command{
	Use:   "uninstall-hooks",
	Short: "Uninstall Git hooks",
	Long:  `Remove Git hooks installed by Secret Vault CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleGitUninstallHooks()
	},
}

// Export/Import command flags
var (
	exportFormat     string
	exportOutput     string
	exportEncrypted  bool
	exportConfirm    bool
	importFormat     string
	importMerge      bool
	importOverwrite  bool
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export vault secrets to a file",
	Long: `Export vault secrets to JSON or YAML format.
By default, only metadata is exported (names, descriptions, timestamps).
Use --confirm flag to include secret values in the export.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleExport()
	},
}

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import secrets from a file",
	Long: `Import secrets from a JSON or YAML file.
The format is auto-detected from the file extension or can be specified with --format.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleImport(args[0])
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
	rootCmd.AddCommand(gitCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)

	// Add git subcommands
	gitCmd.AddCommand(gitInstallHooksCmd)
	gitCmd.AddCommand(gitUninstallHooksCmd)

	// Export command flags
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "json", "output format (json, yaml)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "output file path (default: vault-export.<format>)")
	exportCmd.Flags().BoolVar(&exportEncrypted, "encrypted", false, "export in encrypted form (not yet implemented)")
	exportCmd.Flags().BoolVar(&exportConfirm, "confirm", false, "confirm export of secret values")

	// Import command flags
	importCmd.Flags().StringVarP(&importFormat, "format", "f", "", "input format (json, yaml) - auto-detected if not specified")
	importCmd.Flags().BoolVar(&importMerge, "merge", true, "merge with existing secrets (default)")
	importCmd.Flags().BoolVar(&importOverwrite, "overwrite", false, "overwrite existing secrets without prompting")
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

func handleGitInstallHooks() {
	// Get vault path for hooks
	vaultPath, err := getVaultPath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Check if we're in a git repository
	_, err = git.GetGitRootDir()
	if err != nil {
		fmt.Println("Error: not in a Git repository")
		fmt.Println("Navigate to a Git repository and try again")
		os.Exit(1)
	}

	// Install hooks
	err = git.InstallHooks("", vaultPath)
	if err != nil {
		fmt.Printf("Error installing hooks: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Git hooks installed successfully")
	fmt.Println()
	fmt.Println("The following hooks have been installed:")
	fmt.Println("  • pre-push: Notifies about token opportunities when pushing")
	fmt.Println()
	fmt.Println("Try pushing to a remote to see the integration in action!")
}

func handleGitUninstallHooks() {
	// Check if we're in a git repository
	_, err := git.GetGitRootDir()
	if err != nil {
		fmt.Println("Error: not in a Git repository")
		fmt.Println("Navigate to a Git repository and try again")
		os.Exit(1)
	}

	// Uninstall hooks
	err = git.UninstallHooks()
	if err != nil {
		fmt.Printf("Error uninstalling hooks: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Git hooks uninstalled successfully")
}

func handleExport() {
	vaultPath, err := getVaultPath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !vault.VaultExists(vaultPath) {
		fmt.Println("Error: vault not initialized. Run 'secretvault init' first")
		os.Exit(1)
	}

	// Get formatter
	formatter, err := export.GetFormatter(exportFormat)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Determine output file
	outputFile := exportOutput
	if outputFile == "" {
		outputFile = "vault-export" + formatter.FileExtension()
	}

	// Check if including secret values
	includeValues := exportConfirm
	if !includeValues {
		fmt.Println("⚠️  Exporting metadata only (names, descriptions, timestamps)")
		fmt.Println("   Use --confirm to include secret values")
	} else {
		fmt.Println("⚠️  Warning: Exporting with secret values!")
		fmt.Print("   Type 'yes' to confirm: ")
		reader := bufio.NewReader(os.Stdin)
		confirmation, _ := reader.ReadString('\n')
		confirmation = strings.TrimSpace(strings.ToLower(confirmation))
		if confirmation != "yes" {
			fmt.Println("Export cancelled")
			os.Exit(0)
		}
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

	// Export data
	data, err := formatter.Marshal(v, includeValues)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Write to file
	err = os.WriteFile(outputFile, data, 0600)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Vault exported to %s (%s format)\n", outputFile, formatter.Name())
	if includeValues {
		fmt.Println("⚠️  Warning: File contains secret values - keep it secure!")
	}
}

func handleImport(filePath string) {
	vaultPath, err := getVaultPath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !vault.VaultExists(vaultPath) {
		fmt.Println("Error: vault not initialized. Run 'secretvault init' first")
		os.Exit(1)
	}

	// Read import file
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Determine format
	format := importFormat
	if format == "" {
		// Auto-detect from file extension
		if strings.HasSuffix(filePath, ".json") {
			format = "json"
		} else if strings.HasSuffix(filePath, ".yaml") || strings.HasSuffix(filePath, ".yml") {
			format = "yaml"
		} else {
			fmt.Println("Error: cannot auto-detect format, please specify with --format")
			os.Exit(1)
		}
	}

	// Get formatter
	formatter, err := export.GetFormatter(format)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Parse import data
	importedVault, err := formatter.Unmarshal(data)
	if err != nil {
		fmt.Printf("Error parsing import file: %v\n", err)
		os.Exit(1)
	}

	// Validate that secrets have values
	hasEmptyValues := false
	for _, secret := range importedVault.Secrets {
		if secret.Value == "" {
			hasEmptyValues = true
			break
		}
	}

	if hasEmptyValues {
		fmt.Println("⚠️  Warning: Some secrets in the import file have empty values")
		fmt.Println("   These may be metadata-only exports")
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

	// Merge secrets
	conflictCount := 0
	addedCount := 0
	skippedCount := 0

	for name, importedSecret := range importedVault.Secrets {
		if existingSecret, exists := v.Secrets[name]; exists {
			conflictCount++
			
			if !importOverwrite {
				fmt.Printf("\n⚠️  Secret '%s' already exists\n", name)
				fmt.Printf("   Existing: created %s, updated %s\n", 
					existingSecret.CreatedAt.Format("2006-01-02"),
					existingSecret.UpdatedAt.Format("2006-01-02"))
				fmt.Printf("   Import:   created %s, updated %s\n", 
					importedSecret.CreatedAt.Format("2006-01-02"),
					importedSecret.UpdatedAt.Format("2006-01-02"))
				fmt.Print("   Overwrite? [y/N]: ")
				
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				response = strings.TrimSpace(strings.ToLower(response))
				
				if response != "y" && response != "yes" {
					fmt.Println("   Skipped")
					skippedCount++
					continue
				}
			}
		} else {
			addedCount++
		}

		v.Secrets[name] = importedSecret
	}

	// Save vault
	err = vault.SaveVault(v, vaultPath, password)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ Import complete:\n")
	fmt.Printf("   Added: %d secret(s)\n", addedCount)
	if conflictCount > 0 {
		fmt.Printf("   Updated: %d secret(s)\n", conflictCount-skippedCount)
		if skippedCount > 0 {
			fmt.Printf("   Skipped: %d secret(s)\n", skippedCount)
		}
	}
}
