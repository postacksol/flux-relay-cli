package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install or update the Flux Relay CLI",
	Long: `Install or update the Flux Relay CLI to the latest version.

This command will:
- Use 'go install' if Go is available
- Or download and run the platform-specific installer

Examples:
  flux-relay install          # Install/update using Go
  flux-relay install --force  # Force reinstall`,
	RunE: runInstall,
}

var forceInstall bool

func init() {
	installCmd.Flags().BoolVar(&forceInstall, "force", false, "Force reinstall even if already installed")
	rootCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) error {
	fmt.Println("Flux Relay CLI Installer")
	fmt.Println("========================")
	fmt.Println()

	// Check if Go is installed
	goInstalled := checkGoInstalled()
	if goInstalled {
		fmt.Println("✅ Go found - using 'go install' method")
		fmt.Println()
		return installViaGo()
	}

	// Fall back to platform-specific installer
	fmt.Println("⚠️  Go not found - using platform-specific installer")
	fmt.Println()
	return installViaScript()
}

func checkGoInstalled() bool {
	cmd := exec.Command("go", "version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func installViaGo() error {
	modulePath := "github.com/postacksol/flux-relay-cli@latest"
	
	fmt.Printf("Installing %s...\n", modulePath)
	fmt.Println()
	
	cmd := exec.Command("go", "install", modulePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install via go install: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Installation complete!")
	fmt.Println()
	
	// Find where Go installed it
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		home, _ := os.UserHomeDir()
		goPath = filepath.Join(home, "go")
	}
	binPath := filepath.Join(goPath, "bin", "flux-relay")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	binDir := filepath.Join(goPath, "bin")
	if _, err := os.Stat(binPath); err == nil {
		fmt.Printf("Binary installed to: %s\n", binPath)
		fmt.Println()
		
		// Check if bin directory is in PATH
		pathEnv := os.Getenv("PATH")
		if pathEnv != "" {
			pathList := filepath.SplitList(pathEnv)
			inPath := false
			for _, p := range pathList {
				if p == binDir {
					inPath = true
					break
				}
			}
			
			if !inPath {
				fmt.Println("⚠️  Warning: The Go bin directory is not in your PATH")
				fmt.Println()
				fmt.Printf("Add this to your ~/.bashrc or ~/.zshrc:\n")
				fmt.Printf("  export PATH=\"$PATH:%s\"\n", binDir)
				fmt.Println()
				fmt.Println("Then run:")
				fmt.Printf("  source ~/.bashrc  # or ~/.zshrc\n")
				fmt.Println()
				fmt.Println("Or add it temporarily for this session:")
				fmt.Printf("  export PATH=\"$PATH:%s\"\n", binDir)
			} else {
				fmt.Println("✅ Go bin directory is already in your PATH")
			}
		}
		
		fmt.Println()
		fmt.Println("To verify, run: flux-relay --version")
	} else {
		fmt.Println("Installation completed, but binary location could not be determined.")
		fmt.Println()
		fmt.Println("Make sure $GOPATH/bin or $HOME/go/bin is in your PATH:")
		fmt.Printf("  export PATH=\"$PATH:%s\"\n", binDir)
	}

	return nil
}

func installViaScript() error {
	installURL := "https://raw.githubusercontent.com/postacksol/flux-relay-cli/main/install.ps1"
	
	switch runtime.GOOS {
	case "windows":
		fmt.Println("Running Windows installer...")
		fmt.Println()
		fmt.Printf("If this doesn't work automatically, run:\n")
		fmt.Printf("  irm %s | iex\n", installURL)
		fmt.Println()
		
		// Try to download and run the installer
		cmd := exec.Command("powershell", "-Command", 
			fmt.Sprintf("irm %s | iex", installURL))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run installer: %w\n\nTry running manually:\n  irm %s | iex", err, installURL)
		}
		
	case "linux", "darwin":
		installURL = "https://raw.githubusercontent.com/postacksol/flux-relay-cli/main/install.sh"
		fmt.Println("Running installer script...")
		fmt.Println()
		fmt.Printf("If this doesn't work automatically, run:\n")
		fmt.Printf("  curl -fsSL %s | bash\n", installURL)
		fmt.Println()
		
		// Try to download and run the installer
		cmd := exec.Command("bash", "-c", 
			fmt.Sprintf("curl -fsSL %s | bash", installURL))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run installer: %w\n\nTry running manually:\n  curl -fsSL %s | bash", err, installURL)
		}
		
	default:
		return fmt.Errorf("unsupported platform: %s\n\nPlease install manually. See: https://github.com/postacksol/flux-relay-cli", runtime.GOOS)
	}

	fmt.Println()
	fmt.Println("✅ Installation complete!")
	return nil
}
