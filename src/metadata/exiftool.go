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

// FindExifTool locates the exiftool binary on Windows or Unix systems.
func FindExifTool() (string, error) {
	// 1. Environment variable override
	if p := os.Getenv("EXIFTOOL_PATH"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	// 2. Next to executable
	if exePath, err := os.Executable(); err == nil {
		dir := filepath.Dir(exePath)
		candidates := []string{
			filepath.Join(dir, "exiftool.exe"),
			filepath.Join(dir, "exiftool(-k).exe"),
			filepath.Join(dir, "exiftool"),
			filepath.Join(dir, "exiftool_files", "exiftool.pl"),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c, nil
			}
		}
	}

	// 3. Standard Windows locations
	if runtime.GOOS == "windows" {
		candidates := []string{
			`C:\Program Files\ExifTool\exiftool.exe`,
			`C:\Program Files (x86)\ExifTool\exiftool.exe`,
			`C:\ExifTool\exiftool.exe`,
			`C:\exiftool.exe`,
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c, nil
			}
		}
	}

	// 4. System PATH
	if p, err := exec.LookPath("exiftool"); err == nil {
		return p, nil
	}
	if runtime.GOOS == "windows" {
		if p, err := exec.LookPath("exiftool.exe"); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("exiftool not found. Ensure exiftool is installed and available in PATH")
}

// ExifToolSession wraps a persistent stay_open exiftool process for high performance.
type ExifToolSession struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	reader *bufio.Reader
}

// StartExifToolSession launches an exiftool daemon in stay_open mode.
func StartExifToolSession(exiftoolPath string) (*ExifToolSession, error) {
	cmd := exec.Command(exiftoolPath, "-stay_open", "True", "-@", "-")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, err
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		return nil, err
	}

	return &ExifToolSession{
		cmd:    cmd,
		stdin:  stdin,
		reader: bufio.NewReader(stdout),
	}, nil
}

// Execute sends arguments to the stay_open exiftool process and waits for {ready}.
func (s *ExifToolSession) Execute(args []string) (string, error) {
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
			return strings.Join(outputLines, "\n"), err
		}
		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == "{ready}" {
			break
		}
		outputLines = append(outputLines, trimmed)
	}

	return strings.Join(outputLines, "\n"), nil
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
