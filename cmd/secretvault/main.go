package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"syscall"
	"text/tabwriter"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/matteospanio/secret-vault/pkg/clipboard"
	"github.com/matteospanio/secret-vault/pkg/git"
	"github.com/matteospanio/secret-vault/pkg/sync"
	"github.com/matteospanio/secret-vault/pkg/tui"
	"github.com/matteospanio/secret-vault/pkg/vault"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var (
	vaultPath       string
	password        string
	copyToClipboard bool
	addCategory     string
	addTags         string
	filterCategory  string
	filterTag       string
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
	Long: `Retrieve and print a secret value to stdout (suitable for piping).
Use --copy to copy the value to the clipboard instead of printing it.`,
	Args: cobra.ExactArgs(1),
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

// syncCmd represents the sync command group
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Cloud sync commands",
	Long:  `Commands for synchronizing vault across cloud storage providers.`,
}

// syncConfigureCmd represents the sync configure command
var syncConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure cloud sync",
	Long:  `Configure cloud sync provider (currently supports Nextcloud).`,
	Run: func(cmd *cobra.Command, args []string) {
		handleSyncConfigure()
	},
}

// syncPushCmd represents the sync push command
var syncPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Upload vault to cloud",
	Long:  `Upload the local vault to the configured cloud storage.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleSyncPush()
	},
}

// syncPullCmd represents the sync pull command
var syncPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Download vault from cloud",
	Long:  `Download the vault from the configured cloud storage.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleSyncPull()
	},
}

// syncStatusCmd represents the sync status command
var syncStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sync status",
	Long:  `Display the current sync status and differences between local and remote.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleSyncStatus()
	},
}

// clearClipboardCmd represents the clear-clipboard command
var clearClipboardCmd = &cobra.Command{
	Use:   "clear-clipboard",
	Short: "Clear the system clipboard",
	Long:  `Remove all content from the system clipboard. Useful after copying sensitive secrets.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleClearClipboard()
	},
}

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch terminal user interface",
	Long:  `Launch an interactive terminal user interface for managing secrets.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleTUI()
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

	// Get command flags
	getCmd.Flags().BoolVarP(&copyToClipboard, "copy", "c", false, "copy secret value to clipboard instead of printing")

	// Add command flags
	addCmd.Flags().StringVar(&addCategory, "category", "", "category for the secret")
	addCmd.Flags().StringVar(&addTags, "tags", "", "comma-separated tags for the secret")

	// List command flags
	listCmd.Flags().StringVar(&filterCategory, "filter-category", "", "filter secrets by category")
	listCmd.Flags().StringVar(&filterTag, "filter-tag", "", "filter secrets by tag")

	// Add commands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(gitCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(clearClipboardCmd)
	rootCmd.AddCommand(tuiCmd)

	// Add git subcommands
	gitCmd.AddCommand(gitInstallHooksCmd)
	gitCmd.AddCommand(gitUninstallHooksCmd)

	// Add sync subcommands
	syncCmd.AddCommand(syncConfigureCmd)
	syncCmd.AddCommand(syncPushCmd)
	syncCmd.AddCommand(syncPullCmd)
	syncCmd.AddCommand(syncStatusCmd)
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

	// Parse tags from comma-separated string
	var tags []string
	if addTags != "" {
		for _, tag := range strings.Split(addTags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	v.AddSecret(name, value, description, addCategory, tags)

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

	if copyToClipboard {
		err := clipboard.CopyToClipboard(secret.Value)
		if err != nil {
			fmt.Printf("Error: failed to copy to clipboard: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ Secret copied to clipboard")
	} else {
		fmt.Println(secret.Value)
	}
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

	// Get secrets, optionally filtered
	var secrets []vault.Secret

	if filterCategory != "" || filterTag != "" {
		// Apply filters
		for _, s := range v.Secrets {
			secrets = append(secrets, s)
		}
		if filterCategory != "" {
			filtered := make([]vault.Secret, 0)
			categoryLower := strings.ToLower(filterCategory)
			for _, s := range secrets {
				if strings.ToLower(s.Category) == categoryLower {
					filtered = append(filtered, s)
				}
			}
			secrets = filtered
		}
		if filterTag != "" {
			filtered := make([]vault.Secret, 0)
			tagLower := strings.ToLower(filterTag)
			for _, s := range secrets {
				for _, t := range s.Tags {
					if strings.ToLower(t) == tagLower {
						filtered = append(filtered, s)
						break
					}
				}
			}
			secrets = filtered
		}
	} else {
		for _, s := range v.Secrets {
			secrets = append(secrets, s)
		}
	}

	if len(secrets) == 0 {
		if filterCategory != "" || filterTag != "" {
			fmt.Println("No secrets match the given filters")
		} else {
			fmt.Println("No secrets stored in vault")
		}
		return
	}

	// Sort by name
	sort.Slice(secrets, func(i, j int) bool {
		return secrets[i].Name < secrets[j].Name
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tCATEGORY\tTAGS\tDESCRIPTION\tCREATED\tUPDATED")
	fmt.Fprintln(w, "----\t--------\t----\t-----------\t-------\t-------")

	for _, secret := range secrets {
		description := secret.Description
		if description == "" {
			description = "-"
		}
		category := secret.Category
		if category == "" {
			category = "-"
		}
		tags := "-"
		if len(secret.Tags) > 0 {
			tags = strings.Join(secret.Tags, ", ")
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			secret.Name,
			category,
			tags,
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

func handleSyncConfigure() {
	fmt.Println("Configure Cloud Sync")
	fmt.Println("====================")
	fmt.Println()

	// Get provider
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Provider (nextcloud): ")
	provider, _ := reader.ReadString('\n')
	provider = strings.TrimSpace(provider)
	if provider == "" {
		provider = "nextcloud"
	}

	if provider != "nextcloud" {
		fmt.Printf("Error: unsupported provider '%s'. Currently only 'nextcloud' is supported.\n", provider)
		os.Exit(1)
	}

	// Get Nextcloud configuration
	fmt.Print("Nextcloud URL (e.g., https://cloud.example.com): ")
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)

	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		fmt.Println("Error: failed to read password")
		os.Exit(1)
	}
	ncPassword := string(passwordBytes)

	fmt.Print("Remote path (e.g., /Vaults/vault.enc): ")
	path, _ := reader.ReadString('\n')
	path = strings.TrimSpace(path)

	// Validate inputs
	if url == "" || username == "" || ncPassword == "" || path == "" {
		fmt.Println("Error: all fields are required")
		os.Exit(1)
	}

	// Create config
	config := &sync.SyncConfig{
		Provider: provider,
		Settings: map[string]interface{}{
			"url":      url,
			"username": username,
			"password": ncPassword,
			"path":     path,
		},
	}

	// Get config path
	configPath, err := sync.GetDefaultConfigPath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Save config
	if err := sync.SaveSyncConfig(configPath, config); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("✓ Sync configured successfully\n")
	fmt.Printf("  Provider: %s\n", provider)
	fmt.Printf("  URL: %s\n", url)
	fmt.Printf("  Username: %s\n", username)
	fmt.Printf("  Remote path: %s\n", path)
	fmt.Println()
	fmt.Println("Use 'secretvault sync push' to upload your vault")
	fmt.Println("Use 'secretvault sync pull' to download your vault")
}

func handleSyncPush() {
	// Load sync config
	configPath, err := sync.GetDefaultConfigPath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	config, err := sync.LoadSyncConfig(configPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get vault path
	vaultPath, err := getVaultPath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !vault.VaultExists(vaultPath) {
		fmt.Println("Error: vault not initialized. Run 'secretvault init' first")
		os.Exit(1)
	}

	// Create provider
	var provider sync.SyncProvider
	if config.Provider == "nextcloud" {
		ncConfig, err := sync.ParseNextcloudConfig(config)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		provider = sync.NewNextcloudProvider(ncConfig)
	} else {
		fmt.Printf("Error: unsupported provider '%s'\n", config.Provider)
		os.Exit(1)
	}

	// Connect
	fmt.Println("Connecting to cloud provider...")
	if err := provider.Connect(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer provider.Disconnect()

	// Create sync manager
	manager := sync.NewManager(provider, vaultPath)

	// Push
	fmt.Println("Uploading vault...")
	if err := manager.Push(false); err != nil {
		// Check if it's a conflict error
		if conflictErr, ok := err.(*sync.ErrConflict); ok {
			fmt.Println("Error: Sync conflict detected!")
			fmt.Printf("  Local modified: %v\n", conflictErr.Local.LastModified)
			fmt.Printf("  Remote modified: %v\n", conflictErr.Remote.LastModified)
			fmt.Println()
			fmt.Println("To resolve, you can:")
			fmt.Println("  1. Pull remote changes: secretvault sync pull")
			fmt.Println("  2. Force push (overwrites remote): secretvault sync push --force")
			os.Exit(1)
		}
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Vault uploaded successfully")
}

func handleSyncPull() {
	// Load sync config
	configPath, err := sync.GetDefaultConfigPath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	config, err := sync.LoadSyncConfig(configPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get vault path
	vaultPath, err := getVaultPath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Create provider
	var provider sync.SyncProvider
	if config.Provider == "nextcloud" {
		ncConfig, err := sync.ParseNextcloudConfig(config)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		provider = sync.NewNextcloudProvider(ncConfig)
	} else {
		fmt.Printf("Error: unsupported provider '%s'\n", config.Provider)
		os.Exit(1)
	}

	// Connect
	fmt.Println("Connecting to cloud provider...")
	if err := provider.Connect(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer provider.Disconnect()

	// Create sync manager
	manager := sync.NewManager(provider, vaultPath)

	// Pull
	fmt.Println("Downloading vault...")
	if err := manager.Pull(false); err != nil {
		// Check if it's a conflict error
		if conflictErr, ok := err.(*sync.ErrConflict); ok {
			fmt.Println("Error: Sync conflict detected!")
			fmt.Printf("  Local modified: %v\n", conflictErr.Local.LastModified)
			fmt.Printf("  Remote modified: %v\n", conflictErr.Remote.LastModified)
			fmt.Println()
			fmt.Println("To resolve, you can:")
			fmt.Println("  1. Force pull (overwrites local): secretvault sync pull --force")
			fmt.Println("  2. Push local changes: secretvault sync push")
			os.Exit(1)
		}
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Vault downloaded successfully")
}

func handleSyncStatus() {
	// Load sync config
	configPath, err := sync.GetDefaultConfigPath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	config, err := sync.LoadSyncConfig(configPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get vault path
	vaultPath, err := getVaultPath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Create provider
	var provider sync.SyncProvider
	if config.Provider == "nextcloud" {
		ncConfig, err := sync.ParseNextcloudConfig(config)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		provider = sync.NewNextcloudProvider(ncConfig)
	} else {
		fmt.Printf("Error: unsupported provider '%s'\n", config.Provider)
		os.Exit(1)
	}

	// Connect
	if err := provider.Connect(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer provider.Disconnect()

	// Create sync manager
	manager := sync.NewManager(provider, vaultPath)

	// Get status
	status, err := manager.Status()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Display status
	fmt.Println("Sync Status")
	fmt.Println("===========")
	fmt.Println()
	fmt.Printf("Provider: %s\n", config.Provider)
	fmt.Println()

	if status.LocalExists {
		fmt.Println("Local Vault:")
		fmt.Printf("  Last modified: %v\n", status.LocalMetadata.LastModified.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Size: %d bytes\n", status.LocalMetadata.Size)
		fmt.Printf("  Checksum: %s\n", status.LocalMetadata.Checksum[:16]+"...")
	} else {
		fmt.Println("Local Vault: Not found")
	}

	fmt.Println()

	if status.RemoteExists {
		fmt.Println("Remote Vault:")
		fmt.Printf("  Last modified: %v\n", status.RemoteMetadata.LastModified.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Size: %d bytes\n", status.RemoteMetadata.Size)
		fmt.Printf("  Checksum: %s\n", status.RemoteMetadata.Checksum[:16]+"...")
	} else {
		fmt.Println("Remote Vault: Not found")
	}

	fmt.Println()

	if status.HasConflict {
		fmt.Println("Status: ⚠️  CONFLICT - Both local and remote have changed")
		fmt.Println("Action: Use 'sync pull' or 'sync push' to resolve")
	} else if status.InSync {
		fmt.Println("Status: ✓ In sync")
	} else if status.LocalExists && !status.RemoteExists {
		fmt.Println("Status: Local only - Use 'sync push' to upload")
	} else if !status.LocalExists && status.RemoteExists {
		fmt.Println("Status: Remote only - Use 'sync pull' to download")
	} else {
		fmt.Println("Status: Out of sync")
	}
}

func handleClearClipboard() {
	err := clipboard.ClearClipboard()
	if err != nil {
		fmt.Printf("Error: failed to clear clipboard: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Clipboard cleared")
}

func handleTUI() {
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

	// Launch TUI
	p := tea.NewProgram(tui.NewModel(v))
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}

	// Save vault after TUI exits to persist any changes
	if m, ok := finalModel.(tui.Model); ok {
		err = vault.SaveVault(m.GetVault(), vaultPath, password)
		if err != nil {
			fmt.Printf("Error saving vault: %v\n", err)
			os.Exit(1)
		}
	}
}
