package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	pluginName  = flag.String("name", "", "Plugin name")
	outputDir   = flag.String("output", "dist", "Output directory")
	verbose     = flag.Bool("verbose", false, "Enable verbose output")
	releaseMode = flag.Bool("release", false, "Build in release mode")
)

func main() {
	flag.Parse()

	if *pluginName == "" {
		fmt.Println("Error: Plugin name is required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Create the output directory if it doesn't exist
	err := os.MkdirAll(*outputDir, 0755)
	if err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Build the shared library
	buildSharedLibrary()

	// Build the C wrapper
	buildCWrapper()

	// Package the plugin
	packagePlugin()

	fmt.Printf("Successfully built plugin: %s\n", *pluginName)
}

func buildSharedLibrary() {
	fmt.Println("Building Go shared library...")

	buildFlags := []string{
		"build",
		"-buildmode=c-shared",
	}

	if *releaseMode {
		buildFlags = append(buildFlags, "-ldflags=-s -w")
	} else {
		buildFlags = append(buildFlags, "-gcflags=all=-N -l")
	}

	// Target directory for the shared library
	targetFile := filepath.Join(*outputDir, fmt.Sprintf("lib%s.so", *pluginName))
	buildFlags = append(buildFlags, "-o", targetFile)

	// Add the main package to build
	buildFlags = append(buildFlags, fmt.Sprintf("./examples/%s", *pluginName))

	cmd := exec.Command("go", buildFlags...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if *verbose {
		fmt.Printf("Running: go %s\n", strings.Join(buildFlags, " "))
	}

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error building shared library: %v\n", err)
		os.Exit(1)
	}
}

func buildCWrapper() {
	fmt.Println("Building C wrapper...")

	// Determine the file extension based on the platform
	ext := ".so"
	if runtime.GOOS == "windows" {
		ext = ".dll"
	} else if runtime.GOOS == "darwin" {
		ext = ".dylib"
	}

	cFlags := []string{
		"-g",
		"-shared",
		"-fPIC",
		"-I./include/clap/include", // CLAP include directory
		"-o", filepath.Join(*outputDir, *pluginName+ext),
		"./src/c/plugin.c",
		"-L", *outputDir,
		fmt.Sprintf("-l%s", *pluginName),
	}

	if runtime.GOOS == "darwin" {
		cFlags = append(cFlags, "-undefined", "dynamic_lookup")
	}

	if *releaseMode {
		cFlags = append([]string{"-O3"}, cFlags...)
	}

	if *verbose {
		fmt.Printf("Running: gcc %s\n", strings.Join(cFlags, " "))
	}

	cmd := exec.Command("gcc", cFlags...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error building C wrapper: %v\n", err)
		os.Exit(1)
	}
}

func packagePlugin() {
	fmt.Println("Packaging plugin...")

	// The final .clap package name
	clapFile := filepath.Join(*outputDir, *pluginName+".clap")

	// For macOS, create a bundle
	if runtime.GOOS == "darwin" {
		clapDir := filepath.Join(*outputDir, *pluginName+".clap")
		
		// Create the bundle directory structure
		contentsDir := filepath.Join(clapDir, "Contents")
		macosDir := filepath.Join(contentsDir, "MacOS")
		
		os.MkdirAll(macosDir, 0755)
		
		// Copy the plugin binary to the bundle
		src := filepath.Join(*outputDir, *pluginName+".dylib")
		dst := filepath.Join(macosDir, *pluginName)
		
		copyFile(src, dst)
		
		// Create Info.plist
		infoPlist := filepath.Join(contentsDir, "Info.plist")
		createInfoPlist(infoPlist)
	} else {
		// For other platforms, just rename the file
		src := filepath.Join(*outputDir, *pluginName)
		if runtime.GOOS == "windows" {
			src += ".dll"
		} else {
			src += ".so"
		}
		
		copyFile(src, clapFile)
	}
}

func copyFile(src, dst string) {
	if *verbose {
		fmt.Printf("Copying %s to %s\n", src, dst)
	}
	
	input, err := os.ReadFile(src)
	if err != nil {
		fmt.Printf("Error reading source file: %v\n", err)
		os.Exit(1)
	}
	
	err = os.WriteFile(dst, input, 0644)
	if err != nil {
		fmt.Printf("Error writing destination file: %v\n", err)
		os.Exit(1)
	}
}

func createInfoPlist(path string) {
	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleDevelopmentRegion</key>
	<string>English</string>
	<key>CFBundleExecutable</key>
	<string>%s</string>
	<key>CFBundleIdentifier</key>
	<string>org.clapgo.%s</string>
	<key>CFBundleInfoDictionaryVersion</key>
	<string>6.0</string>
	<key>CFBundleName</key>
	<string>%s</string>
	<key>CFBundlePackageType</key>
	<string>BNDL</string>
	<key>CFBundleVersion</key>
	<string>1.0</string>
	<key>CFBundleShortVersionString</key>
	<string>1.0</string>
</dict>
</plist>`, *pluginName, *pluginName, *pluginName)

	if *verbose {
		fmt.Printf("Creating Info.plist at %s\n", path)
	}
	
	err := os.WriteFile(path, []byte(plist), 0644)
	if err != nil {
		fmt.Printf("Error writing Info.plist: %v\n", err)
		os.Exit(1)
	}
}