//go:build darwin && cgo

package main

/*
#include <stdint.h>

// osxcross ships no compiler-rt, so cross builds from Linux miss this symbol
// (emitted for @available checks in webview). Weak, so native builds keep the real one.
__attribute__((weak)) int32_t __isPlatformVersionAtLeast(uint32_t platform, uint32_t major, uint32_t minor, uint32_t patch) {
	return 1;
}
*/
import "C"
