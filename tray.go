package main

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/getlantern/systray"
)

// Sediakan dua file icon:
//
//	icon.ico       → ikon normal (biru / warna bebas)
//	icon_alert.ico → ikon alert untuk flash saat data kosong (merah)
//
//go:embed .\icon.ico
var iconNormal []byte

//go:embed .\icon_alert.ico
var iconAlert []byte

var mStatus *systray.MenuItem

func updateStatus(s string) {
	if mStatus != nil {
		mStatus.SetTitle("Status: " + s)
	}
	systray.SetTooltip("ClipRelay — " + s)
	log.Printf("[status] %s", s)
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}

func onReady() {
	systray.SetIcon(iconNormal)
	systray.SetTooltip("ClipRelay — Idle")

	// ── Menu items ────────────────────────────────────────────────
	mStatus = systray.AddMenuItem("Status: Idle", "")
	mStatus.Disable() // non-clickable, hanya info

	systray.AddSeparator()
	mClear := systray.AddMenuItem("Clear", "Kosongkan buffer lokal")

	systray.AddSeparator()
	mDataFetch := systray.AddMenuItem("Data Fetch", "")
	mManual := mDataFetch.AddSubMenuItem("Manual Fetch", "Ambil data dari GAS sekarang (tanpa typing)")
	mPreview := mDataFetch.AddSubMenuItem("Preview", "Lihat isi buffer saat ini")

	systray.AddSeparator()
	mDelay := systray.AddMenuItem("Delay Setting", "")
	mSlow := mDelay.AddSubMenuItem("Slow  (80–160ms)", "")
	mNormal := mDelay.AddSubMenuItem("Normal (60–130ms)", "")
	mFast := mDelay.AddSubMenuItem("Fast  (30–80ms)", "")
	mCustom := mDelay.AddSubMenuItem("Custom...", "Set min dan max delay secara manual")

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Exit", "Keluar dari ClipRelay")

	// Tandai preset aktif saat startup
	markPreset(mSlow, mNormal, mFast, getCfg())

	// ── Hotkey INS ────────────────────────────────────────────────
	startHotkeyListener(func() {
		if isTyping.Load() || isFetching.Load() {
			log.Println("[hotkey] busy, diabaikan")
			return
		}
		go handleInsert()
	})

	// ── Tray event loop ───────────────────────────────────────────
	go func() {
		for {
			select {

			case <-mClear.ClickedCh:
				setBuffer("")
				updateStatus("Idle")

			case <-mManual.ClickedCh:
				go func() {
					if isFetching.Load() {
						return
					}
					isFetching.Store(true)
					defer isFetching.Store(false)

					updateStatus("Fetching...")
					text, err := fetchFromGAS(getCfg().GASURL)
					if err != nil {
						updateStatus("Fetch error")
						log.Printf("[manual] %v", err)
						return
					}
					if text == "" {
						updateStatus("Empty")
						flashTrayIcon()
						return
					}
					setBuffer(text)
					updateStatus(fmt.Sprintf("Ready — %d karakter", len([]rune(text))))
				}()

			case <-mPreview.ClickedCh:
				b := getBuffer()
				if b == "" {
					showMsgBox("ClipRelay — Preview", "(Buffer kosong)")
				} else {
					showMsgBox("ClipRelay — Preview", b)
				}

			case <-mSlow.ClickedCh:
				applyPreset(PresetSlow, mSlow, mNormal, mFast)

			case <-mNormal.ClickedCh:
				applyPreset(PresetNormal, mSlow, mNormal, mFast)

			case <-mFast.ClickedCh:
				applyPreset(PresetFast, mSlow, mNormal, mFast)

			case <-mCustom.ClickedCh:
				handleCustomDelay(mSlow, mNormal, mFast)

			case <-mQuit.ClickedCh:
				log.Println("[app] exit")
				systray.Quit()
				return
			}
		}
	}()
}

// handleInsert dipanggil saat INS ditekan: fetch → typing.
func handleInsert() {
	isFetching.Store(true)
	updateStatus("Fetching...")

	text, err := fetchFromGAS(getCfg().GASURL)
	isFetching.Store(false)

	if err != nil {
		updateStatus("Fetch error")
		log.Printf("[ins] fetch error: %v", err)
		return
	}

	if text == "" {
		updateStatus("Empty")
		log.Println("[ins] data kosong → flash")
		flashTrayIcon()
		return
	}

	setBuffer(text)
	isTyping.Store(true)
	updateStatus(fmt.Sprintf("Typing... %d karakter", len([]rune(text))))
	log.Printf("[ins] typing: %q", truncate(text, 50))

	typeText(text)

	isTyping.Store(false)
	updateStatus("Done")
}

// applyPreset menyimpan preset baru ke config dan memperbarui tampilan menu.
func applyPreset(p DelayPreset, mSlow, mNormal, mFast *systray.MenuItem) {
	cfg := getCfg()
	cfg.DelayPreset = p
	setCfg(cfg)
	saveConfig(cfg)
	markPreset(mSlow, mNormal, mFast, cfg)
	log.Printf("[delay] preset: %s", p)
}

// markPreset menampilkan checkmark (✓) pada preset yang aktif.
func markPreset(mSlow, mNormal, mFast *systray.MenuItem, cfg Config) {
	mSlow.SetTitle("Slow  (80–160ms)")
	mNormal.SetTitle("Normal (60–130ms)")
	mFast.SetTitle("Fast  (30–80ms)")
	switch cfg.DelayPreset {
	case PresetSlow:
		mSlow.SetTitle("✓ Slow  (80–160ms)")
	case PresetFast:
		mFast.SetTitle("✓ Fast  (30–80ms)")
	case PresetCustom:
		// tidak ada checkmark di preset — user tahu sudah custom
	default:
		mNormal.SetTitle("✓ Normal (60–130ms)")
	}
}

// handleCustomDelay membuka dua dialog input untuk min dan max delay.
func handleCustomDelay(mSlow, mNormal, mFast *systray.MenuItem) {
	cfg := getCfg()

	minStr, ok := showInputDialog(
		"Delay Setting — Custom",
		fmt.Sprintf("Min delay (ms)\nSekarang: %d", cfg.DelayCustomMin),
		fmt.Sprintf("%d", cfg.DelayCustomMin),
	)
	if !ok {
		return
	}

	maxStr, ok := showInputDialog(
		"Delay Setting — Custom",
		fmt.Sprintf("Max delay (ms)\nSekarang: %d", cfg.DelayCustomMax),
		fmt.Sprintf("%d", cfg.DelayCustomMax),
	)
	if !ok {
		return
	}

	minVal := parseIntInput(minStr, cfg.DelayCustomMin)
	maxVal := parseIntInput(maxStr, cfg.DelayCustomMax)

	if minVal <= 0 || maxVal <= minVal {
		showMsgBox("ClipRelay", "Input tidak valid.\nMin harus > 0 dan Max harus lebih besar dari Min.")
		return
	}

	cfg.DelayPreset = PresetCustom
	cfg.DelayCustomMin = minVal
	cfg.DelayCustomMax = maxVal
	setCfg(cfg)
	saveConfig(cfg)
	markPreset(mSlow, mNormal, mFast, cfg)
	log.Printf("[delay] custom: %d–%dms", minVal, maxVal)
}
