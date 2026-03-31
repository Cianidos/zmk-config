package main

import (
	"archive/zip"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var (
	quiet    bool
	pollMs   int
	mountDir string
	pollWait time.Duration
)

func log(format string, args ...any) {
	if !quiet {
		fmt.Printf(format+"\n", args...)
	}
}

func main() {
	flag.BoolVar(&quiet, "q", false, "quiet mode, suppress all informational output")
	flag.IntVar(&pollMs, "t", 150, "poll interval in milliseconds for device detection")
	flag.StringVar(&mountDir, "mount", "", "exact mount path of the keyboard drive (skips auto-detection)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <directory or .zip with .uf2 files>\n\nFlashes ZMK firmware to a split nice!nano keyboard.\n\nThe input must contain 3 .uf2 files with \"reset\", \"left\", and \"right\" in their names.\nThe script will flash in sequence: reset → left → reset → right.\n\nFor each step:\n  1. Connect/reset your nice!nano so it appears as a USB drive (NICENANO)\n  2. The script detects the mount, copies the firmware\n  3. The device reboots automatically\n  4. Repeat for the next step\n\nDevice auto-detection checks:\n  Linux:   /proc/mounts, /media/<user>/NICENANO, /run/media/<user>/NICENANO\n  macOS:   /Volumes/NICENANO\n  Windows: drive letters D-Z (reads INFO_UF2.TXT to confirm nice!nano)\n\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	pollWait = time.Duration(pollMs) * time.Millisecond

	src := flag.Arg(0)
	info, err := os.Stat(src)
	if err != nil {
		fatal("cannot access %q: %v", src, err)
	}

	var dir string
	var cleanup func()

	if !info.IsDir() && strings.HasSuffix(strings.ToLower(src), ".zip") {
		log("Extracting zip archive: %s", src)
		dir, err = os.MkdirTemp("", "flash-zmk-*")
		if err != nil {
			fatal("mktemp: %v", err)
		}
		cleanup = func() {
			log("Cleaning up temp directory: %s", dir)
			os.RemoveAll(dir)
		}
		if err := unzip(src, dir); err != nil {
			cleanup()
			fatal("unzip: %v", err)
		}
		log("Extracted .uf2 files to %s", dir)
	} else if info.IsDir() {
		dir = src
		log("Using firmware directory: %s", dir)
	} else {
		fatal("%q is not a directory or .zip file", src)
	}

	if cleanup != nil {
		defer cleanup()
	}

	resetFile, leftFile, rightFile, err := findUF2Files(dir)
	if err != nil {
		fatal("%v", err)
	}

	log("Found firmware files:")
	log("  reset: %s", filepath.Base(resetFile))
	log("  left:  %s", filepath.Base(leftFile))
	log("  right: %s", filepath.Base(rightFile))

	steps := []struct {
		label       string
		file        string
		instruction string
	}{
		{"reset (for left half)", resetFile, "Connect or double-tap reset on the LEFT half of your keyboard"},
		{"left", leftFile, "The left half should re-appear in bootloader mode after reset"},
		{"reset (for right half)", resetFile, "Now connect or double-tap reset on the RIGHT half of your keyboard"},
		{"right", rightFile, "The right half should re-appear in bootloader mode after reset"},
	}

	for i, step := range steps {
		log("")
		log("═══ Step %d/%d: Flash %s ═══", i+1, len(steps), step.label)
		log("→ %s", step.instruction)
		log("Waiting for NICENANO device to appear (polling every %dms)...", pollMs)

		mountPoint := waitForDevice()
		log("✓ Device found at: %s", mountPoint)
		log("  Copying %s → %s", filepath.Base(step.file), mountPoint)

		dest := filepath.Join(mountPoint, filepath.Base(step.file))
		if err := copyFile(step.file, dest); err != nil {
			if isExpectedCopyError(err, mountPoint) {
				log("  I/O error during copy (expected — device reboots mid-transfer)")
			} else {
				fatal("copy failed: %v", err)
			}
		}

		log("  File copied. Waiting for device to disconnect (it reboots after flashing)...")
		waitForDeviceGone(mountPoint)
		log("✓ Device disconnected.")

		// small delay before next cycle
		time.Sleep(1 * time.Second)
	}

	log("")
	log("══════════════════════════════════════")
	log("All done! Both halves flashed successfully.")
	log("══════════════════════════════════════")
}

// isExpectedCopyError returns true when a copy error is caused by the device
// rebooting mid-transfer, which is normal UF2 bootloader behavior.
func isExpectedCopyError(err error, mountPoint string) bool {
	// Linux/macOS: POSIX EIO when device vanishes during write
	if errors.Is(err, syscall.EIO) {
		return true
	}
	// macOS: ENXIO when device node disappears
	if errors.Is(err, syscall.ENXIO) {
		return true
	}
	// Any OS: if the mount point is already gone, the reboot happened
	if _, statErr := os.Stat(mountPoint); os.IsNotExist(statErr) {
		return true
	}
	return false
}

func findUF2Files(dir string) (reset, left, right string, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", "", "", fmt.Errorf("read dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".uf2") {
			continue
		}
		lower := strings.ToLower(e.Name())
		full := filepath.Join(dir, e.Name())
		switch {
		case strings.Contains(lower, "reset"):
			reset = full
		case strings.Contains(lower, "left"):
			left = full
		case strings.Contains(lower, "right"):
			right = full
		}
	}

	var missing []string
	if reset == "" {
		missing = append(missing, "reset")
	}
	if left == "" {
		missing = append(missing, "left")
	}
	if right == "" {
		missing = append(missing, "right")
	}
	if len(missing) > 0 {
		return "", "", "", fmt.Errorf("missing .uf2 files containing: %s", strings.Join(missing, ", "))
	}
	return
}

func waitForDevice() string {
	if mountDir != "" {
		for {
			info, err := os.Stat(mountDir)
			if err == nil && info.IsDir() {
				time.Sleep(pollWait)
				return mountDir
			}
			time.Sleep(pollWait)
		}
	}
	for {
		if mp := findNiceNanoMount(); mp != "" {
			time.Sleep(pollWait)
			return mp
		}
		time.Sleep(pollWait)
	}
}

func waitForDeviceGone(mountPoint string) {
	for {
		if _, err := os.Stat(mountPoint); os.IsNotExist(err) {
			return
		}
		entries, err := os.ReadDir(mountPoint)
		if err != nil || len(entries) == 0 {
			return
		}
		time.Sleep(pollWait)
	}
}

func findNiceNanoMount() string {
	switch runtime.GOOS {
	case "windows":
		return findNiceNanoWindows()
	case "darwin":
		return findNiceNanoDarwin()
	default:
		return findNiceNanoLinux()
	}
}

func findNiceNanoLinux() string {
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		upper := strings.ToUpper(line)
		if strings.Contains(upper, "NICENANO") || strings.Contains(upper, "NICE_NANO") || strings.Contains(upper, "NICE!NANO") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1]
			}
		}
	}

	// fallback: scan common mount locations for directory name match
	for _, base := range []string{"/media", "/run/media"} {
		matches, _ := filepath.Glob(filepath.Join(base, "*", "*NICENANO*"))
		if len(matches) == 0 {
			matches, _ = filepath.Glob(filepath.Join(base, "*", "*NICE?NANO*"))
		}
		if len(matches) == 0 {
			matches, _ = filepath.Glob(filepath.Join(base, "*NICENANO*"))
		}
		if len(matches) == 0 {
			matches, _ = filepath.Glob(filepath.Join(base, "*NICE?NANO*"))
		}
		for _, m := range matches {
			info, err := os.Stat(m)
			if err == nil && info.IsDir() {
				return m
			}
		}
	}

	return ""
}

func findNiceNanoDarwin() string {
	// macOS auto-mounts USB drives under /Volumes/<label>
	for _, pattern := range []string{
		"/Volumes/*NICENANO*",
		"/Volumes/*NICE_NANO*",
		"/Volumes/*NICE!NANO*",
	} {
		matches, _ := filepath.Glob(pattern)
		for _, m := range matches {
			info, err := os.Stat(m)
			if err == nil && info.IsDir() {
				return m
			}
		}
	}
	return ""
}

func findNiceNanoWindows() string {
	// Windows assigns a drive letter to USB mass storage devices.
	// UF2 bootloaders always place INFO_UF2.TXT in the drive root;
	// read it to confirm it belongs to a nice!nano.
	for _, letter := range "DEFGHIJKLMNOPQRSTUVWXYZ" {
		root := string(letter) + ":\\"
		data, err := os.ReadFile(filepath.Join(root, "INFO_UF2.TXT"))
		if err != nil {
			continue
		}
		upper := strings.ToUpper(string(data))
		if strings.Contains(upper, "NICENANO") || strings.Contains(upper, "NICE_NANO") || strings.Contains(upper, "NICE!NANO") {
			return root
		}
	}
	return ""
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		name := filepath.Base(f.Name)
		if !strings.HasSuffix(strings.ToLower(name), ".uf2") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		outPath := filepath.Join(dest, name)
		out, err := os.Create(outPath)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		rc.Close()
		out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
