package cffunc

import (
	"net/http"
	"os"
	"testing"
)

func TestGetPublicIP(t *testing.T) {
	ipinfoclient := http.Client{}

	// Test getting the public IP address
	publicip, err := Get_public_ip(&ipinfoclient)
	if err != nil {
		t.Error(err)
	}
	if publicip == "" {
		t.Error("Expected a non-empty public IP address, but got an empty string")
	}
}

func TestOpenAndRead(t *testing.T) {
	// Create a test file with some data
	testFile, err := os.Create("test_file.txt")
	if err != nil {
		t.Error(err)
	}
	defer testFile.Close()
	testData := "test data\n"
	testFile.WriteString(testData)

	// Test opening and reading the file
	lines, err := Openandread("test_file.txt")
	if err != nil {
		t.Error(err)
	}
	if len(lines) != 1 {
		t.Errorf("Expected 1 line, but got %d", len(lines))
	}
	if lines[0] != testData {
		t.Errorf("Expected %q, but got %q", testData, lines[0])
	}
}
