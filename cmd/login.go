package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/postacksol/flux-relay-cli/internal/api"
	"github.com/postacksol/flux-relay-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type DadJoke struct {
	ID     string `json:"id"`
	Joke   string `json:"joke"`
	Status int    `json:"status"`
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Flux Relay",
	Long: `Login to Flux Relay using device code authentication.
This will open your browser to complete the authentication flow.

If browser cannot be opened (e.g., in WSL or headless environments),
use --headless flag to get a URL and device code to paste manually.`,
	RunE: runLogin,
}

var headlessMode bool

func init() {
	loginCmd.Flags().BoolVar(&headlessMode, "headless", false, "Headless mode: show URL and code instead of opening browser")
	rootCmd.AddCommand(loginCmd)
}

func printLogo() {
	// Set UTF-8 output for Windows console
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "chcp", "65001", ">", "nul")
		cmd.Run()
	}
	
	logo := `
███████╗██╗     ██╗   ██╗██╗  ██╗     ██████╗ ███████╗██╗      █████╗ ██╗   ██╗
██╔════╝██║     ██║   ██║╚██╗██╔╝     ██╔══██╗██╔════╝██║     ██╔══██╗╚██╗ ██╔╝
█████╗  ██║     ██║   ██║ ╚███╔╝      ██████╔╝█████╗  ██║     ███████║ ╚████╔╝ 
██╔══╝  ██║     ██║   ██║ ██╔██╗      ██╔══██╗██╔══╝  ██║     ██╔══██║  ╚██╔╝  
██║     ███████╗╚██████╔╝██╔╝ ██╗     ██║  ██║███████╗███████╗██║  ██║   ██║   
╚═╝     ╚══════╝ ╚═════╝ ╚═╝  ╚═╝     ╚═╝  ╚═╝╚══════╝╚══════╝╚═╝  ╚═╝   ╚═╝   
 
`
	fmt.Print(logo)
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Get API URL from flag, config, or default
	apiURL := apiBaseURL
	if apiURL == "" {
		// Try config file
		apiURL = viper.GetString("api_url")
		if apiURL == "" {
			// Default to localhost for development
			apiURL = "http://localhost:3000"
		}
	}

	// Check if already logged in
	cfg := config.New()
	accessToken := cfg.GetAccessToken()
	if accessToken != "" {
		// Try to validate the token by getting user info
		client := api.NewClient(apiURL)
		userInfo, err := client.GetCurrentUser(accessToken)
		if err == nil && userInfo != nil {
			// Already logged in!
			printLogo()
			fmt.Println("Already logged in!")
			fmt.Println()
			fmt.Printf("   Email: %s\n", userInfo.Email())
			if userInfo.Username() != "" {
				fmt.Printf("   Username: %s\n", userInfo.Username())
			}
			fmt.Printf("   User ID: %s\n", userInfo.ID())
			fmt.Println()
			fmt.Println("To log in as a different user, run 'flux-relay logout' first.")
			fmt.Println()
			return nil
		}
		// Token is invalid/expired, continue with login flow
	}

	printLogo()
	fmt.Println("Starting authentication flow...")
	fmt.Println()

	// Step 1: Request device code
	client := api.NewClient(apiURL)
	deviceCode, err := client.InitiateDeviceCode()
	if err != nil {
		return fmt.Errorf("failed to initiate device code: %w", err)
	}

	fmt.Printf("Device Code: %s\n", deviceCode.UserCode)
	fmt.Printf("Verification URL: %s\n", deviceCode.VerificationURI)
	fmt.Println()

	// Handle headless mode or browser open failure
	useHeadless := headlessMode
	if !useHeadless {
		fmt.Println("Opening browser...")
		if err := openBrowser(deviceCode.VerificationURI); err != nil {
			// Browser failed to open, fall back to headless mode
			fmt.Println()
			fmt.Println("WARNING: Could not open browser automatically.")
			fmt.Println("Switching to headless mode...")
			fmt.Println()
			useHeadless = true
		}
	}

	if useHeadless {
		// Headless mode: show URL and instructions, then exit
		fmt.Println("Visit the following URL to login:")
		fmt.Printf("   %s\n", deviceCode.VerificationURI)
		fmt.Println()
		fmt.Println("After logging in, you will receive an access token on the success page.")
		fmt.Println("Copy the token and run:")
		fmt.Printf("   flux-relay config set token \"YOUR_TOKEN_HERE\"\n")
		fmt.Println()
		return nil
	}

	// Normal mode: browser opened successfully
	fmt.Println("Waiting for authentication...")
	fmt.Println("   (Press Ctrl+C to cancel)")
	fmt.Println()
	fmt.Println("Tip: After logging in on the browser, wait a moment for authorization to complete.")

	// Step 2: Wait a moment before starting to poll (give browser time to open and user to see the page)
	fmt.Println("   Waiting 3 seconds before checking...")
	time.Sleep(3 * time.Second)

	// Step 3: Poll for token
	tokenResponse, err := pollForToken(client, deviceCode.DeviceCode, deviceCode.Interval)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Step 4: Save token
	if err := cfg.SaveToken(tokenResponse); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Println()
	fmt.Println("Authentication complete!")
	fmt.Println("   Token saved to:", cfg.ConfigPath())
	fmt.Println()

	return nil
}

var lastJokeLineCount int = 0

func printJokeOnOneLine(joke string) {
	// Clear previous joke lines first
	for i := 0; i < lastJokeLineCount; i++ {
		fmt.Print("\033[A\033[K") // Move up one line and clear it
	}
	lastJokeLineCount = 0

	// Get terminal width (default to 80 if we can't determine it)
	width := 80
	if w, err := getTerminalWidth(); err == nil && w > 0 {
		width = w
	}

	// Account for the "[Joke] " prefix (8 characters)
	availableWidth := width - 8
	if availableWidth < 20 {
		availableWidth = 60 // Minimum width
	}

	// Format joke to fit on one line if possible
	jokeText := strings.TrimSpace(joke)
	if len(jokeText) <= availableWidth {
		// Fits on one line
		fmt.Printf("   [Joke] %s\n", jokeText)
		lastJokeLineCount = 1
	} else {
		// Too long, wrap it but try to keep it minimal
		words := strings.Fields(jokeText)
		line := "   [Joke] "
		for _, word := range words {
			if len(line)+len(word)+1 > width {
				// Current line is full, start new line
				fmt.Println(line)
				lastJokeLineCount++
				line = "          " + word // Indent continuation lines
			} else {
				if line != "   [Joke] " && line != "          " {
					line += " "
				}
				line += word
			}
		}
		if line != "   [Joke] " && line != "          " {
			fmt.Println(line)
			lastJokeLineCount++
		}
	}
}

func getTerminalWidth() (int, error) {
	// Try to get terminal width
	if runtime.GOOS == "windows" {
		// On Windows, try to get console width
		// This is a simple approach - in production you might want to use a library
		return 80, nil // Default fallback
	}
	
	// On Unix-like systems, try to get from environment or stty
	// For now, return default
	return 80, nil
}

func fetchDadJoke() string {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", "https://icanhazdadjoke.com/", nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Flux Relay CLI (https://github.com/postacksol/flux-relay-cli)")

	resp, err := client.Do(req)
	if err != nil {
		return "" // Silently fail - jokes are optional
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "" // Silently fail
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var joke DadJoke
	if err := json.Unmarshal(body, &joke); err != nil {
		return ""
	}

	// The API returns status 200 in the JSON, but we should check the joke field
	if joke.Joke == "" {
		return ""
	}

	return joke.Joke
}

func pollForToken(client *api.Client, deviceCode string, interval int) (*api.TokenResponse, error) {
	// Start with immediate poll, then use interval
	firstPoll := true
	pollCount := 0
	maxPolls := 120 // 10 minutes max (120 * 5 seconds)
	lastJokeTime := time.Now()
	jokeInterval := 8 * time.Second // Show a new joke every 8 seconds

	fmt.Println() // New line before starting

	// Show first joke IMMEDIATELY while waiting
	joke := fetchDadJoke()
	if joke == "" {
		// Fallback if API fails - show a default joke
		joke = "Why did the developer go broke? Because he used up all his cache!"
	}
	printJokeOnOneLine(joke)
	lastJokeTime = time.Now()
	fmt.Print("   Polling")

	for pollCount < maxPolls {
		// Wait before polling (except first time)
		if !firstPoll {
			time.Sleep(time.Duration(interval) * time.Second)
		}
		firstPoll = false
		pollCount++

		// Check for new joke (every 8 seconds after the first)
		elapsedSinceLastJoke := time.Since(lastJokeTime)
		if elapsedSinceLastJoke >= jokeInterval {
			// Clear the polling dots line first
			fmt.Print("\r\033[K") // Clear current line
			
			joke := fetchDadJoke()
			if joke != "" {
				// Show joke (will clear previous joke lines)
				printJokeOnOneLine(joke)
				fmt.Print("   Polling") // Restart polling indicator
				lastJokeTime = time.Now()
			} else {
				// If joke fetch failed, still update time to avoid spamming
				lastJokeTime = time.Now()
				// Restore polling indicator
				fmt.Print("   Polling")
			}
		}

		token, err := client.GetToken(deviceCode)
		if err == nil {
			fmt.Print("\r\033[K") // Clear polling line
			
			// Show a celebration joke on success!
			joke := fetchDadJoke()
			if joke != "" {
				fmt.Printf("   Success! Here's a joke to celebrate:\n")
				fmt.Printf("   %s\n", joke)
			}
			
			fmt.Println()
			return token, nil
		}

		// Check if it's an authorization_pending error (expected)
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.Code() == "authorization_pending" {
				// Just add a dot, don't reprint "Polling"
				fmt.Print(".")
				continue
			}
			if apiErr.Code() == "access_denied" {
				fmt.Println() // New line
				return nil, fmt.Errorf("authorization was denied")
			}
			// For "Invalid device code" - this might be a timing issue, retry a few times
			if apiErr.Code() == "Invalid device code" || apiErr.Code() == "invalid_device_code" {
				if pollCount <= 3 {
					// Retry a few times in case of timing issues
					fmt.Print(".")
					continue
				}
				fmt.Println() // New line
				return nil, fmt.Errorf("device code not found after multiple attempts. Please make sure:\n   1. You've opened the verification URL in your browser\n   2. You've logged in successfully\n   3. The device code hasn't expired (10 minutes)\n   4. Try running 'flux-relay login' again")
			}
			// For expired device code
			if apiErr.Code() == "Device code expired" || apiErr.Code() == "device_code_expired" {
				fmt.Println() // New line
				return nil, fmt.Errorf("device code expired. Please run 'flux-relay login' again to get a new code")
			}
			// Log other API errors but continue polling (might be temporary)
			if pollCount%10 == 0 { // Only log every 10th poll to avoid spam
				fmt.Printf("\n   [*] Still waiting... (attempt %d/%d)\n", pollCount, maxPolls)
			}
			continue
		}

		// For non-API errors, log occasionally and continue (might be network issues)
		if pollCount%10 == 0 {
			fmt.Printf("\n   [*] Network issue, retrying... (attempt %d/%d)\n", pollCount, maxPolls)
		}
		// Continue polling - might be temporary network issue
		continue
	}

	fmt.Println() // New line
	return nil, fmt.Errorf("authentication timeout after 10 minutes. Please try again")
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Start()
}
