package ui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/nitintf/openport/internal/client"
)

var (
	// Brand
	logoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	versionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	// Info
	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Width(12)

	urlStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))

	arrowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("236"))

	// Request log
	methodStyle = lipgloss.NewStyle().
			Bold(true).
			Width(7)

	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	durationStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	tsStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("239"))

	dotOK = lipgloss.NewStyle().
		Foreground(lipgloss.Color("76")).
		Render("●")

	dotRedirect = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Render("●")

	dotClientErr = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")).
			Render("●")

	dotServerErr = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")).
			Render("●")

	statusOKStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("76"))

	statusRedirectStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214"))

	statusClientErrStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("203"))

	statusServerErrStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("196"))

	// Errors
	errIcon = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true).
		Render("✕")

	errTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	errMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	errHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	// Shutdown
	checkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("76")).
			Render("✓")
)

// PrintBanner displays the startup tunnel information.
func PrintBanner(tunnelURL, localAddr string) {
	fmt.Println()
	fmt.Printf("  %s %s\n",
		logoStyle.Render("openport"),
		versionStyle.Render("v0.1.0"),
	)
	fmt.Println()

	fmt.Printf("  %s %s  %s  %s\n",
		labelStyle.Render("Forwarding"),
		urlStyle.Render(tunnelURL),
		arrowStyle.Render("→"),
		urlStyle.Render("http://"+localAddr),
	)
	fmt.Println()
	fmt.Printf("  %s\n", hintStyle.Render("Press Ctrl+C to stop"))

	divider := dividerStyle.Render("  " + strings.Repeat("─", 52))
	fmt.Println(divider)
	fmt.Println()
}

// PrintRequestLog renders a single styled request log line.
func PrintRequestLog(r client.RequestLog) {
	dot := statusDot(r.StatusCode)
	status := statusCodeStyle(r.StatusCode).Render(strconv.Itoa(r.StatusCode))
	method := methodStyle.Render(r.Method)
	path := pathStyle.Render(r.Path)
	dur := formatDuration(r.Duration)
	ts := tsStyle.Render(r.Timestamp.Format("15:04:05"))

	fmt.Printf("  %s %s %s %s %s %s\n", dot, ts, status, method, path, dur)
}

// PrintError displays a human-friendly error message.
func PrintError(err error) {
	fmt.Println()

	var ce *client.ConnectError
	if errors.As(err, &ce) {
		switch {
		case errors.Is(ce.Kind, client.ErrLocalNotReachable):
			printErrorBlock(
				"Port not reachable",
				fmt.Sprintf("Nothing is running on %s.", ce.Detail),
				"Start your local server first, then try again.",
			)
		case errors.Is(ce.Kind, client.ErrServerUnreachable):
			printErrorBlock(
				"Cannot reach server",
				fmt.Sprintf("Could not connect to the openport server at %s.", ce.Addr),
				"Make sure the server is running and the address is correct.",
			)
		case errors.Is(ce.Kind, client.ErrSubdomainTaken):
			printErrorBlock(
				"Subdomain unavailable",
				fmt.Sprintf("The subdomain \"%s\" is already in use.", ce.Detail),
				"Try a different subdomain with --subdomain or omit it for a random one.",
			)
		case errors.Is(ce.Kind, client.ErrConnectionLost):
			printErrorBlock(
				"Connection lost",
				"The tunnel connection was interrupted.",
				"Check your network and try reconnecting.",
			)
		default:
			printErrorBlock("Something went wrong", ce.Detail, "")
		}
	} else {
		printErrorBlock("Something went wrong", err.Error(), "")
	}

	fmt.Println()
}

func printErrorBlock(title, message, hint string) {
	fmt.Printf("  %s %s\n", errIcon, errTitleStyle.Render(title))
	fmt.Printf("    %s\n", errMsgStyle.Render(message))
	if hint != "" {
		fmt.Printf("    %s\n", errHintStyle.Render(hint))
	}
}

// PrintShutdown displays a clean disconnection message.
func PrintShutdown() {
	fmt.Println()
	fmt.Printf("  %s %s\n", checkStyle, versionStyle.Render("Tunnel closed."))
	fmt.Println()
}

func statusDot(code int) string {
	switch {
	case code >= 200 && code < 300:
		return dotOK
	case code >= 300 && code < 400:
		return dotRedirect
	case code >= 400 && code < 500:
		return dotClientErr
	default:
		return dotServerErr
	}
}

func statusCodeStyle(code int) lipgloss.Style {
	switch {
	case code >= 200 && code < 300:
		return statusOKStyle
	case code >= 300 && code < 400:
		return statusRedirectStyle
	case code >= 400 && code < 500:
		return statusClientErrStyle
	default:
		return statusServerErrStyle
	}
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Millisecond)
	if d < time.Second {
		return durationStyle.Render(fmt.Sprintf("%dms", d.Milliseconds()))
	}
	return durationStyle.Render(fmt.Sprintf("%.1fs", d.Seconds()))
}
