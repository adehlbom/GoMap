// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/src/mac_helpers.go
package main

import (
	"os/exec"
	"strings"
)

// tryAlternativeHostnameMethods attempts to get device hostname using methods other than DNS lookup
func tryAlternativeHostnameMethods(ip string) string {
	// Try ARP table lookup for the MAC address
	cmd := exec.Command("arp", "-n", ip)
	output, err := cmd.Output()

	if err == nil && len(output) > 0 {
		// Parse the MAC address from ARP output (format varies by OS)
		outputStr := string(output)

		// Simple MAC extraction - this is Mac OS specific
		// Example: ? (192.168.1.1) at a0:63:91:f0:cc:4b on en0 ifscope permanent [ethernet]
		parts := strings.Fields(outputStr)
		for i, part := range parts {
			if part == "at" && i+1 < len(parts) {
				mac := parts[i+1]

				// Try to map MAC to a vendor
				vendor := getMACVendor(mac)
				if vendor != "" {
					return vendor + "-" + strings.Replace(mac, ":", "-", -1)
				}

				return "MAC-" + strings.Replace(mac, ":", "-", -1)
			}
		}
	}

	return ""
}

// getMACVendor returns a manufacturer name based on MAC address prefix (simplified version)
func getMACVendor(mac string) string {
	mac = strings.ToUpper(mac)

	// This is a simplified lookup with just a few common vendors
	// In a real application, you might want to use a complete OUI database
	prefixToVendor := map[string]string{
		"00:0C:29": "VMware",
		"00:50:56": "VMware",
		"00:1A:11": "Google",
		"3C:22:FB": "Apple",
		"A8:66:7F": "Apple",
		"D8:BB:2C": "Apple",
		"34:36:3B": "Apple",
		"F4:5C:89": "Apple",
		"C8:2A:14": "Apple",
		"B8:27:EB": "Raspberry",
		"DC:A6:32": "Raspberry",
		"E4:5F:01": "Raspberry",
		"F0:B4:29": "Xiaomi",
		"18:FE:34": "Espressif",
		"3C:71:BF": "Espressif",
		"8C:AA:B5": "Samsung",
		"5C:F3:70": "Samsung",
		"58:A0:23": "Google",
		"E4:F0:42": "Google",
		"00:17:88": "Philips",
		"EC:B5:FA": "Philips",
		"00:14:A8": "Sonos",
		"B8:E9:37": "Sonos",
		"B8:F0:09": "Amazon",
		"FC:65:DE": "Amazon",
	}

	// Check if the MAC starts with any known prefix
	for prefix, vendor := range prefixToVendor {
		if strings.HasPrefix(mac, prefix) {
			return vendor
		}
	}

	return ""
}
