package fritzbox

import (
	"fmt"
	"testing"
)

func Test_Bitmask(t *testing.T) {
	var p uint32 = 40960

	// Check if bit 6 and bit 8 are set
	const (
		bit6 = 1 << 5 // Bit 6 (zero-based)
		bit8 = 1 << 7 // Bit 8
	)

	functions := toDeviceFunctions(p)

	if len(functions) == 0 {
		t.Error("functions is empty")
	}

	bit8Test := 1 << (TemperatureSensor - 1)

	if bit8 != bit8Test {
		t.Error("bit8 != bit8Test")
	}

	if (p&bit6 != 0) && (p&bit8 != 0) {
		fmt.Println("✅ Bit 6 and Bit 8 are set.")
	} else {
		fmt.Println("❌ Bit 6 and/or Bit 8 are not set.")
	}
}
