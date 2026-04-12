//go:build 386 || amd64 || amd64p32

package platform

const (
	ecx1SSE3    = 0
	ecx1SSSE3   = 9
	ecx1FMA     = 12
	ecx1CX16    = 13
	ecx1SSE4_1  = 19
	ecx1SSE4_2  = 20
	ecx1MOVBE   = 22
	ecx1POPCNT  = 23
	ecx1XSAVE   = 26
	ecx1OSXSAVE = 27
	ecx1AVX     = 28
	ecx1F16C    = 29

	ebx7BMI1     = 3
	ebx7AVX2     = 5
	ebx7BMI2     = 8
	ebx7AVX512F  = 16
	ebx7AVX512DQ = 17
	ebx7AVX512CD = 28
	ebx7AVX512BW = 30
	ebx7AVX512VL = 31

	ecxxLAHF  = 0
	ecxxLZCNT = 5

	eaxOSXMM      = 1
	eaxOSYMM      = 2
	eaxOSOpMask   = 5
	eaxOSZMMHi16  = 6
	eaxOSZMMHi256 = 7
)

var (
	// GOAMD64=v1 (default): The baseline. Exclusively generates instructions that all 64-bit x86 processors can execute.
	// GOAMD64=v2: all v1 instructions, plus CX16, LAHF-SAHF, POPCNT, SSE3, SSE4.1, SSE4.2, SSSE3.
	// GOAMD64=v3: all v2 instructions, plus AVX, AVX2, BMI1, BMI2, F16C, FMA, LZCNT, MOVBE, OSXSAVE.
	// GOAMD64=v4: all v3 instructions, plus AVX512F, AVX512BW, AVX512CD, AVX512DQ, AVX512VL.
	ecx1FeaturesV2  = bitSet(ecx1CX16) | bitSet(ecx1POPCNT) | bitSet(ecx1SSE3) | bitSet(ecx1SSE4_1) | bitSet(ecx1SSE4_2) | bitSet(ecx1SSSE3)
	ecx1FeaturesV3  = ecx1FeaturesV2 | bitSet(ecx1AVX) | bitSet(ecx1F16C) | bitSet(ecx1FMA) | bitSet(ecx1MOVBE) | bitSet(ecx1OSXSAVE)
	ebx7FeaturesV3  = bitSet(ebx7AVX2) | bitSet(ebx7BMI1) | bitSet(ebx7BMI2)
	ebx7FeaturesV4  = ebx7FeaturesV3 | bitSet(ebx7AVX512F) | bitSet(ebx7AVX512BW) | bitSet(ebx7AVX512CD) | bitSet(ebx7AVX512DQ) | bitSet(ebx7AVX512VL)
	ecxxFeaturesV2  = bitSet(ecxxLAHF)
	ecxxFeaturesV3  = ecxxFeaturesV2 | bitSet(ecxxLZCNT)
	eaxOSFeaturesV3 = bitSet(eaxOSXMM) | bitSet(eaxOSYMM)
	eaxOSFeaturesV4 = eaxOSFeaturesV3 | bitSet(eaxOSOpMask) | bitSet(eaxOSZMMHi16) | bitSet(eaxOSZMMHi256)
)

// cpuid is implemented in cpuinfo_x86.s.
func cpuid(eaxArg, ecxArg uint32) (eax, ebx, ecx, edx uint32)

// xgetbv with ecx = 0 is implemented in cpu_x86.s.
func xgetbv() (eax, edx uint32)

func lookupCPUVariant() string {
	variant := "v1"
	maxID, _, _, _ := cpuid(0, 0)
	if maxID < 7 {
		return variant
	}
	_, _, ecx1, _ := cpuid(1, 0)
	_, ebx7, _, _ := cpuid(7, 0)
	maxX, _, _, _ := cpuid(0x80000000, 0)
	_, _, ecxx, _ := cpuid(0x80000001, 0)

	if maxX < 0x80000001 || !bitIsSet(ecx1FeaturesV2, ecx1) || !bitIsSet(ecxxFeaturesV2, ecxx) {
		return variant
	}
	variant = "v2"

	if !bitIsSet(ecx1FeaturesV3, ecx1) || !bitIsSet(ebx7FeaturesV3, ebx7) || !bitIsSet(ecxxFeaturesV3, ecxx) {
		return variant
	}
	// For XGETBV, OSXSAVE bit is required and verified by ecx1FeaturesV3.
	eaxOS, _ := xgetbv()
	if !bitIsSet(eaxOSFeaturesV3, eaxOS) {
		return variant
	}
	variant = "v3"

	// Darwin support for AVX-512 appears to have issues.
	if isDarwin || !bitIsSet(ebx7FeaturesV4, ebx7) || !bitIsSet(eaxOSFeaturesV4, eaxOS) {
		return variant
	}
	variant = "v4"

	return variant
}

func bitSet(bitpos uint) uint32 {
	return 1 << bitpos
}

func bitIsSet(bits, value uint32) bool {
	return (value & bits) == bits
}
