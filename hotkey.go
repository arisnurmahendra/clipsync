package main

import (
	"log"
	"runtime"
	"unsafe"
)

const (
	wmHotkey = 0x0312 // WM_HOTKEY
	vkInsert = 0x2D   // VK_INSERT
	hotkeyID = 1
)

// MSG adalah struktur Windows MSG untuk GetMessage.
type wMsg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	PtX     int32
	PtY     int32
}

// startHotkeyListener mendaftarkan tombol INS sebagai global hotkey
// dan memanggil onIns() setiap kali ditekan.
//
// Catatan: RegisterHotKey mengirim WM_HOTKEY ke message queue
// thread yang mendaftarkannya. Karena itu goroutine ini di-lock
// ke satu OS thread (LockOSThread) dan punya GetMessage loop sendiri,
// terpisah dari message loop systray.
func startHotkeyListener(onIns func()) {
	go func() {
		// Kunci goroutine ke OS thread ini agar WM_HOTKEY
		// masuk ke queue thread yang sama dengan RegisterHotKey.
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		r, _, err := procRegisterHotKey.Call(0, hotkeyID, 0, vkInsert)
		if r == 0 {
			log.Printf("[hotkey] RegisterHotKey gagal: %v", err)
			return
		}
		defer procUnregisterHotKey.Call(0, hotkeyID)
		log.Println("[hotkey] INS terdaftar")

		var m wMsg
		for {
			// GetMessage blocking — hanya kembalikan saat ada message.
			// Return 0 = WM_QUIT, -1 = error.
			r, _, _ := procGetMessage.Call(
				uintptr(unsafe.Pointer(&m)), 0, 0, 0,
			)
			if r == 0 || r == ^uintptr(0) {
				log.Println("[hotkey] GetMessage berhenti")
				return
			}
			procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
			procDispatchMessage.Call(uintptr(unsafe.Pointer(&m)))

			if m.Message == wmHotkey && m.WParam == hotkeyID {
				onIns()
			}
		}
	}()
}
