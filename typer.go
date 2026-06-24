package main

import (
	"math/rand"
	"time"
	"unicode"
	"unsafe"
)

// ── Konstanta delay (ms) ──────────────────────────────────────────

const (
	// Dalam kata: cepat, natural
	inWordMin = 30
	inWordMax = 60

	// Antar kata (setelah spasi)
	wordPauseMin = 80
	wordPauseMax = 250

	// Setelah tanda baca (akhir kalimat/frasa)
	punctPauseMin = 180
	punctPauseMax = 450

	// Micro-pause acak di tengah kata (jari ragu sejenak)
	microPauseMin   = 120
	microPauseMax   = 300
	microPauseEvery = 8 // rata-rata 1 dari N karakter dalam kata

	// Jeda sebelum huruf kapital (simulasi tahan Shift)
	shiftDelayMin = 20
	shiftDelayMax = 55

	// Penambahan delay untuk karakter berulang (ll, ss, dst)
	repeatDelayAdd = 25

	// Angka / simbol di tengah teks
	symbolDelayMin = 10
	symbolDelayMax = 25

	// ── Typo timing ───────────────────────────────────────────────
	// Delay setelah mengetik karakter salah (user "sadar" sesaat)
	typoRealizePauseMin = 120
	typoRealizePauseMax = 280

	// Delay setelah backspace (tangan kembali ke posisi huruf yang benar)
	typoAfterBackspaceMin = 80
	typoAfterBackspaceMax = 180

	// Startup: jeda setelah INS ditekan sebelum mulai typing
	startupDelayMs = 500
)

var isPunct = map[rune]bool{
	'.': true, ',': true, '?': true, '!': true,
	';': true, ':': true,
}

// ── Windows INPUT struct (SendInput, 64-bit, 40 bytes) ────────────
//
// Layout offset (64-bit):
//   [0]  inputType  uint32   — INPUT_KEYBOARD = 1
//   [4]  _          uint32   — padding (Go align union ke 8)
//   [8]  wVk        uint16   — virtual key (0 jika pakai Unicode)
//  [10]  wScan      uint16   — Unicode codepoint
//  [12]  dwFlags    uint32   — KEYEVENTF_UNICODE | KEYEVENTF_KEYUP
//  [16]  kbTime     uint32   — 0 = sistem isi sendiri
//  [20]  _          uint32   — padding (align uintptr ke 8)
//  [24]  dwExtra    uintptr  — 0
//  [32]  _          [8]byte  — pad union = 32 bytes (MOUSEINPUT size)
//  Total: 40 bytes = sizeof(INPUT) ✓

type keyboardInput struct {
	wVk     uint16
	wScan   uint16
	dwFlags uint32
	kbTime  uint32
	_       uint32 // padding sebelum uintptr
	dwExtra uintptr
}

type winInput struct {
	inputType uint32
	_         uint32 // padding sebelum union
	ki        keyboardInput
	_         [8]byte // pad union ke 32 bytes
}

const (
	inputKeyboard    = 1
	keyeventfUnicode = 0x0004
	keyeventfKeyup   = 0x0002
	vkBack           = uint16(0x08) // VK_BACK
)

// sendChar mengirim satu karakter Unicode via SendInput (down + up).
func sendChar(ch rune) {
	inputs := [2]winInput{
		{
			inputType: inputKeyboard,
			ki:        keyboardInput{wScan: uint16(ch), dwFlags: keyeventfUnicode},
		},
		{
			inputType: inputKeyboard,
			ki:        keyboardInput{wScan: uint16(ch), dwFlags: keyeventfUnicode | keyeventfKeyup},
		},
	}
	procSendInput.Call(
		2,
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(unsafe.Sizeof(inputs[0])),
	)
}

// sendBackspace mengirim satu penekanan Backspace (VK_BACK).
func sendBackspace() {
	inputs := [2]winInput{
		{inputType: inputKeyboard, ki: keyboardInput{wVk: vkBack, dwFlags: 0}},
		{inputType: inputKeyboard, ki: keyboardInput{wVk: vkBack, dwFlags: keyeventfKeyup}},
	}
	procSendInput.Call(
		2,
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(unsafe.Sizeof(inputs[0])),
	)
}

// ── Keyboard neighbor map untuk typo realistis ────────────────────

var neighborKeys = map[rune][]rune{
	'q': {'w', 'a'},
	'w': {'q', 'e', 'a', 's'},
	'e': {'w', 'r', 's', 'd'},
	'r': {'e', 't', 'd', 'f'},
	't': {'r', 'y', 'f', 'g'},
	'y': {'t', 'u', 'g', 'h'},
	'u': {'y', 'i', 'h', 'j'},
	'i': {'u', 'o', 'j', 'k'},
	'o': {'i', 'p', 'k', 'l'},
	'p': {'o', 'l'},
	'a': {'q', 'w', 's', 'z'},
	's': {'a', 'w', 'e', 'd', 'z', 'x'},
	'd': {'s', 'e', 'r', 'f', 'x', 'c'},
	'f': {'d', 'r', 't', 'g', 'c', 'v'},
	'g': {'f', 't', 'y', 'h', 'v', 'b'},
	'h': {'g', 'y', 'u', 'j', 'b', 'n'},
	'j': {'h', 'u', 'i', 'k', 'n', 'm'},
	'k': {'j', 'i', 'o', 'l', 'm'},
	'l': {'k', 'o', 'p'},
	'z': {'a', 's', 'x'},
	'x': {'z', 's', 'd', 'c'},
	'c': {'x', 'd', 'f', 'v'},
	'v': {'c', 'f', 'g', 'b'},
	'b': {'v', 'g', 'h', 'n'},
	'n': {'b', 'h', 'j', 'm'},
	'm': {'n', 'j', 'k'},
}

// neighborOf mengembalikan satu karakter tetangga keyboard.
// Huruf kapital tetap kapital pada hasil.
func neighborOf(ch rune) rune {
	lower := unicode.ToLower(ch)
	neighbors, ok := neighborKeys[lower]
	if !ok || len(neighbors) == 0 {
		return ch
	}
	pick := neighbors[rand.Intn(len(neighbors))]
	if unicode.IsUpper(ch) {
		return unicode.ToUpper(pick)
	}
	return pick
}

// ── Helper RNG ────────────────────────────────────────────────────

func randInt(min, max int) int {
	if min >= max {
		return min
	}
	return min + rand.Intn(max-min+1)
}

func sleepMs(min, max int) {
	time.Sleep(time.Duration(randInt(min, max)) * time.Millisecond)
}

// ── typeText: engine utama ────────────────────────────────────────

func typeText(text string) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Baca config sekali — konsisten selama sesi typing berlangsung
	typoRate := getCfg().TypoRate

	// Jeda startup: beri waktu user pindah fokus ke field target
	time.Sleep(startupDelayMs * time.Millisecond)

	chars := []rune(text)
	total := len(chars)
	var prev rune
	posInWord := 0 // posisi karakter dalam kata saat ini (reset saat spasi/punct)

	for i, ch := range chars {
		isFirst := i == 0
		isLast := i == total-1

		// ── Typo simulation ───────────────────────────────────
		// Hanya huruf, bukan posisi pertama/terakhir, dan hanya
		// jika typoRate > 0 dan RNG menghendaki.
		if typoRate > 0 &&
			unicode.IsLetter(ch) &&
			!isFirst && !isLast &&
			rng.Intn(100) < typoRate {

			wrong := neighborOf(ch)
			if wrong != ch {
				// 1. Ketik karakter salah
				sendChar(wrong)

				// 2. Jeda: user "membaca" sebentar dan sadar ada typo
				sleepMs(typoRealizePauseMin, typoRealizePauseMax)

				// 3. Backspace
				sendBackspace()

				// 4. Jeda: tangan balik ke posisi huruf yang benar
				sleepMs(typoAfterBackspaceMin, typoAfterBackspaceMax)
			}
		}

		// ── Kirim karakter asli ───────────────────────────────
		sendChar(ch)

		// ── Delay setelah karakter ────────────────────────────
		switch {

		case isPunct[ch]:
			// Akhir kalimat / frasa — jeda paling panjang
			sleepMs(punctPauseMin, punctPauseMax)
			posInWord = 0

		case ch == ' ':
			// Batas antar kata — jeda natural
			sleepMs(wordPauseMin, wordPauseMax)
			posInWord = 0

		default:
			// Dalam kata — cepat
			delay := randInt(inWordMin, inWordMax)

			// Huruf kapital: tahan Shift sebentar sebelum huruf berikutnya
			if unicode.IsUpper(ch) && !unicode.IsUpper(prev) && !isFirst {
				delay += randInt(shiftDelayMin, shiftDelayMax)
			}

			// Huruf berulang: jari harus "reset" ke posisi yang sama
			if ch == prev && ch != ' ' {
				delay += repeatDelayAdd
			}

			// Angka / simbol di tengah teks
			if !unicode.IsLetter(ch) && ch != ' ' && !isPunct[ch] {
				delay += randInt(symbolDelayMin, symbolDelayMax)
			}

			// Micro-pause acak: sesekali jari berhenti sejenak
			// Dimulai setelah karakter ke-2 dalam kata agar tidak
			// muncul di awal kata (terasa aneh).
			posInWord++
			if posInWord > 2 && rng.Intn(microPauseEvery) == 0 {
				delay += randInt(microPauseMin, microPauseMax)
			}

			time.Sleep(time.Duration(delay) * time.Millisecond)
		}

		prev = ch
	}

	// ── Tanda selesai: double space ───────────────────────────────
	// GAS sudah normalisasi input → tidak ada double space dari teks asli.
	// Double space ini hanya bisa datang dari daemon — sinyal "done" untuk user.
	sleepMs(wordPauseMin, wordPauseMax)
	sendChar(' ')
	sleepMs(40, 90)
	sendChar(' ')
}
