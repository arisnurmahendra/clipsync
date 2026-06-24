package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/getlantern/systray"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procSendInput        = user32.NewProc("SendInput")
	procMessageBoxW      = user32.NewProc("MessageBoxW")
	procRegisterHotKey   = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey = user32.NewProc("UnregisterHotKey")
	procGetMessage       = user32.NewProc("GetMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procDispatchMessage  = user32.NewProc("DispatchMessageW")
)

// ── SendInput ─────────────────────────────────────────────────────
//
// Struktur harus 40 bytes pada 64-bit Windows agar sesuai sizeof(INPUT).
// Layout (64-bit):
//
//	offset  0: Type      uint32  (4)
//	offset  4: (padding)         (4)  ← Go align ki ke 8
//	offset  8: ki.Vk    uint16  (2)
//	offset 10: ki.Scan  uint16  (2)
//	offset 12: ki.Flags uint32  (4)
//	offset 16: ki.Time  uint32  (4)
//	offset 20: (padding)         (4)  ← Go align ExtraInfo ke 8
//	offset 24: ki.Extra uintptr (8)
//	offset 32: Pad      [8]byte (8)   ← pad union ke 32 (ukuran MOUSEINPUT)
//	Total: 40 bytes ✓

// ── MessageBox ────────────────────────────────────────────────────

func showMsgBox(title, msg string) {
	t, _ := syscall.UTF16PtrFromString(title)
	m, _ := syscall.UTF16PtrFromString(msg)
	procMessageBoxW.Call(0,
		uintptr(unsafe.Pointer(m)),
		uintptr(unsafe.Pointer(t)),
		0, // MB_OK
	)
}

// ── Input Dialog via PowerShell VB InputBox ───────────────────────

func showInputDialog(title, prompt, def string) (string, bool) {
	// Escape single quotes untuk PowerShell
	esc := func(s string) string {
		return strings.ReplaceAll(s, "'", "''")
	}
	script := fmt.Sprintf(
		`Add-Type -AssemblyName Microsoft.VisualBasic; `+
			`[Microsoft.VisualBasic.Interaction]::InputBox('%s','%s','%s')`,
		esc(prompt), esc(title), esc(def),
	)
	out, err := exec.Command(
		"powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", script,
	).Output()
	if err != nil {
		return "", false
	}
	result := strings.TrimSpace(string(out))
	if result == "" {
		return "", false
	}
	return result, true
}

// parseIntInput membaca string ke int, fallback ke def jika gagal.
func parseIntInput(s string, def int) int {
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || v <= 0 {
		return def
	}
	return v
}

// ── Flash tray icon ───────────────────────────────────────────────
//
// Bergantian antara iconNormal dan iconAlert sebanyak N kali.
// Ini adalah sinyal visual bahwa data fetch hasilnya kosong.

func flashTrayIcon() {
	go func() {
		for i := 0; i < 6; i++ {
			if i%2 == 0 {
				systray.SetIcon(iconAlert)
			} else {
				systray.SetIcon(iconNormal)
			}
			time.Sleep(350 * time.Millisecond)
		}
		systray.SetIcon(iconNormal)
	}()
}
