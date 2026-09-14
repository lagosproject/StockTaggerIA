package metadata

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ExifToolCommand defines the executable and base arguments needed to launch ExifTool.
type ExifToolCommand struct {
	Binary string
	Args   []string
}

// BuildCommand constructs an *exec.Cmd with the base command plus any additional arguments.
func (c *ExifToolCommand) BuildCommand(extraArgs ...string) *exec.Cmd {
	args := make([]string, 0, len(c.Args)+len(extraArgs))
	args = append(args, c.Args...)
	args = append(args, extraArgs...)
	return exec.Command(c.Binary, args...)
}

// ResolveExifToolCommand locates exiftool as a standalone binary or Perl-wrapped script.
func ResolveExifToolCommand() (*ExifToolCommand, error) {
	// 1. Environment variable override
	if p := os.Getenv("EXIFTOOL_PATH"); p != "" {
		if _, err := os.Stat(p); err == nil {
			if strings.HasSuffix(strings.ToLower(p), ".pl") && runtime.GOOS == "windows" {
				if perlPath, err := exec.LookPath("perl.exe"); err == nil {
					return &ExifToolCommand{Binary: perlPath, Args: []string{p}}, nil
				}
			}
			return &ExifToolCommand{Binary: p}, nil
		}
	}

	// 2. Next to executable
	if exePath, err := os.Executable(); err == nil {
		dir := filepath.Dir(exePath)
		candidates := []string{
			filepath.Join(dir, "exiftool.exe"),
			filepath.Join(dir, "exiftool(-k).exe"),
			filepath.Join(dir, "exiftool"),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return &ExifToolCommand{Binary: c}, nil
			}
		}

		// Check for exiftool_files package (e.g. Strawberry Perl bundle on Windows)
		plPath := filepath.Join(dir, "exiftool_files", "exiftool.pl")
		if _, err := os.Stat(plPath); err == nil {
			if runtime.GOOS == "windows" {
				perlCand := filepath.Join(dir, "exiftool_files", "perl.exe")
				if _, err := os.Stat(perlCand); err == nil {
					return &ExifToolCommand{Binary: perlCand, Args: []string{plPath}}, nil
				}
				if p, err := exec.LookPath("perl.exe"); err == nil {
					return &ExifToolCommand{Binary: p, Args: []string{plPath}}, nil
				}
			} else {
				return &ExifToolCommand{Binary: plPath}, nil
			}
		}
	}

	// 3. Standard platform locations
	if runtime.GOOS == "windows" {
		candidates := []string{
			`C:\Program Files\ExifTool\exiftool.exe`,
			`C:\Program Files (x86)\ExifTool\exiftool.exe`,
			`C:\ExifTool\exiftool.exe`,
			`C:\exiftool.exe`,
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return &ExifToolCommand{Binary: c}, nil
			}
		}
	} else if runtime.GOOS == "darwin" {
		// macOS Homebrew paths (Apple Silicon ARM64 & Intel)
		candidates := []string{
			"/opt/homebrew/bin/exiftool",
			"/usr/local/bin/exiftool",
			"/opt/local/bin/exiftool",
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return &ExifToolCommand{Binary: c}, nil
			}
		}
	} else {
		// Linux standard locations
		candidates := []string{
			"/usr/bin/exiftool",
			"/usr/local/bin/exiftool",
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return &ExifToolCommand{Binary: c}, nil
			}
		}
	}

	// 4. System PATH
	if p, err := exec.LookPath("exiftool"); err == nil {
		return &ExifToolCommand{Binary: p}, nil
	}
	if runtime.GOOS == "windows" {
		if p, err := exec.LookPath("exiftool.exe"); err == nil {
			return &ExifToolCommand{Binary: p}, nil
		}
	}

	return nil, fmt.Errorf("exiftool not found. Ensure exiftool is installed and available in PATH")
}

// FindExifTool locates the exiftool binary on Windows or Unix systems.
func FindExifTool() (string, error) {
	cmd, err := ResolveExifToolCommand()
	if err != nil {
		return "", err
	}
	return cmd.Binary, nil
}

type safeBuffer struct {
	strings.Builder
}

// ExifToolSession wraps a persistent stay_open exiftool process for high performance.
type ExifToolSession struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	reader    *bufio.Reader
	stderrBuf *safeBuffer
	cmdConfig *ExifToolCommand
}

// StartExifToolSession launches an exiftool daemon in stay_open mode.
func StartExifToolSession(exiftoolPath string) (*ExifToolSession, error) {
	return StartExifToolSessionWithCmd(&ExifToolCommand{Binary: exiftoolPath})
}

// StartExifToolSessionWithCmd launches an exiftool daemon using the configured command.
func StartExifToolSessionWithCmd(cmdConfig *ExifToolCommand) (*ExifToolSession, error) {
	cmd := cmdConfig.BuildCommand("-stay_open", "True", "-@", "-")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		stdout.Close()
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		return nil, err
	}

	sess := &ExifToolSession{
		cmd:       cmd,
		stdin:     stdin,
		reader:    bufio.NewReader(stdout),
		stderrBuf: &safeBuffer{},
		cmdConfig: cmdConfig,
	}

	// Capture stderr continuously in background
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			sess.stderrBuf.WriteString(scanner.Text() + "\n")
		}
	}()

	return sess, nil
}

// Execute sends arguments to the stay_open exiftool process and waits for {ready}.
func (s *ExifToolSession) Execute(args []string) (string, error) {
	// Clear stderr buffer before running command
	s.stderrBuf.Reset()

	for _, arg := range args {
		if _, err := fmt.Fprintln(s.stdin, arg); err != nil {
			return "", err
		}
	}
	if _, err := fmt.Fprintln(s.stdin, "-execute"); err != nil {
		return "", err
	}

	var outputLines []string
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			stderrMsg := strings.TrimSpace(s.stderrBuf.String())
			if stderrMsg != "" {
				return strings.Join(outputLines, "\n"), fmt.Errorf("%w: %s", err, stderrMsg)
			}
			return strings.Join(outputLines, "\n"), err
		}
		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == "{ready}" {
			break
		}
		outputLines = append(outputLines, trimmed)
	}

	combined := strings.Join(outputLines, "\n")
	stderrMsg := strings.TrimSpace(s.stderrBuf.String())

	// Check if ExifTool reported an error
	if strings.Contains(combined, "0 image files updated") ||
		strings.Contains(combined, "weren't updated due to errors") ||
		strings.Contains(stderrMsg, "Error:") {
		errDetail := stderrMsg
		if errDetail == "" {
			errDetail = combined
		}
		return combined, fmt.Errorf("exiftool write failed: %s", errDetail)
	}

	return combined, nil
}

// Close gracefully terminates the stay_open exiftool process.
func (s *ExifToolSession) Close() {
	if s.stdin != nil {
		fmt.Fprintln(s.stdin, "-stay_open")
		fmt.Fprintln(s.stdin, "False")
		s.stdin.Close()
	}
	if s.cmd != nil {
		s.cmd.Wait()
	}
}
