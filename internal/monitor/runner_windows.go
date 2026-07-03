//go:build windows

package monitor

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/hoza-labs/clip-history/internal/logging"
)

const (
	wmClose           = 0x0010
	wmDestroy         = 0x0002
	wmClipboardUpdate = 0x031D
	hwndMessage       = ^uintptr(2)
)

type point struct {
	x int32
	y int32
}

type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

var (
	user32                            = syscall.NewLazyDLL("user32.dll")
	kernel32                          = syscall.NewLazyDLL("kernel32.dll")
	procAddClipboardFormatListener    = user32.NewProc("AddClipboardFormatListener")
	procCreateWindowExW               = user32.NewProc("CreateWindowExW")
	procDefWindowProcW                = user32.NewProc("DefWindowProcW")
	procDestroyWindow                 = user32.NewProc("DestroyWindow")
	procDispatchMessageW              = user32.NewProc("DispatchMessageW")
	procGetMessageW                   = user32.NewProc("GetMessageW")
	procPostMessageW                  = user32.NewProc("PostMessageW")
	procPostQuitMessage               = user32.NewProc("PostQuitMessage")
	procRegisterClassExW              = user32.NewProc("RegisterClassExW")
	procRemoveClipboardFormatListener = user32.NewProc("RemoveClipboardFormatListener")
	procTranslateMessage              = user32.NewProc("TranslateMessage")
	procGetModuleHandleW              = kernel32.NewProc("GetModuleHandleW")
	monitorWindowProc                 = syscall.NewCallback(wndProc)
	monitorInstances                  sync.Map
)

type runner struct {
	hwnd      uintptr
	logger    logging.Logger
	statePath string
	stopPath  string
	startedAt time.Time
}

func Run(logger logging.Logger, statePath, stopPath string) error {
	if err := RemoveFile(stopPath); err != nil {
		return fmt.Errorf("clear stop signal: %w", err)
	}

	instance, _, err := procGetModuleHandleW.Call(0)
	if instance == 0 {
		return fmt.Errorf("get module handle: %w", err)
	}

	className, err := syscall.UTF16PtrFromString("ClipHistoryMonitorWindow")
	if err != nil {
		return fmt.Errorf("prepare class name: %w", err)
	}

	class := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		lpfnWndProc:   monitorWindowProc,
		hInstance:     instance,
		lpszClassName: className,
	}

	if atom, _, registerErr := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&class))); atom == 0 && registerErr != syscall.Errno(1410) {
		return fmt.Errorf("register window class: %w", registerErr)
	}

	hwnd, _, createErr := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(className)),
		0,
		0,
		0,
		0,
		0,
		hwndMessage,
		0,
		instance,
		0,
	)
	if hwnd == 0 {
		return fmt.Errorf("create message window: %w", createErr)
	}

	now := time.Now().UTC()
	r := &runner{
		hwnd:      hwnd,
		logger:    logger,
		statePath: statePath,
		stopPath:  stopPath,
		startedAt: now,
	}
	monitorInstances.Store(hwnd, r)
	defer monitorInstances.Delete(hwnd)
	defer func() {
		_ = RemoveFile(statePath)
		_ = RemoveFile(stopPath)
	}()

	if ok, _, addErr := procAddClipboardFormatListener.Call(hwnd); ok == 0 {
		_, _, _ = procDestroyWindow.Call(hwnd)
		return fmt.Errorf("register clipboard listener: %w", addErr)
	}
	defer procRemoveClipboardFormatListener.Call(hwnd)

	if err := SaveState(statePath, State{
		PID:         os.Getpid(),
		StartedAt:   now,
		HeartbeatAt: now,
	}); err != nil {
		_, _, _ = procDestroyWindow.Call(hwnd)
		return err
	}

	done := make(chan struct{})
	defer close(done)
	go r.heartbeat(done)

	logger.Info("clipboard monitor started")

	var message msg
	for {
		ret, _, getErr := procGetMessageW.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		switch int32(ret) {
		case -1:
			return fmt.Errorf("message loop failed: %w", getErr)
		case 0:
			logger.Info("clipboard monitor stopped")
			return nil
		default:
			procTranslateMessage.Call(uintptr(unsafe.Pointer(&message)))
			procDispatchMessageW.Call(uintptr(unsafe.Pointer(&message)))
		}
	}
}

func (r *runner) heartbeat(done <-chan struct{}) {
	ticker := time.NewTicker(DefaultHeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if _, err := os.Stat(r.stopPath); err == nil {
				procPostMessageW.Call(r.hwnd, wmClose, 0, 0)
				return
			}

			now := time.Now().UTC()
			_ = SaveState(r.statePath, State{
				PID:         os.Getpid(),
				StartedAt:   r.startedAt,
				HeartbeatAt: now,
			})
		}
	}
}

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	value, _ := monitorInstances.Load(hwnd)
	r, _ := value.(*runner)

	switch msg {
	case wmClipboardUpdate:
		if r != nil {
			now := time.Now().UTC()
			_ = SaveState(r.statePath, State{
				PID:         os.Getpid(),
				StartedAt:   r.startedAt,
				HeartbeatAt: now,
			})
			r.logger.Info("clipboard update detected")
		}
		return 0
	case wmClose:
		procDestroyWindow.Call(hwnd)
		return 0
	case wmDestroy:
		procPostQuitMessage.Call(0)
		return 0
	default:
		ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
		return ret
	}
}
