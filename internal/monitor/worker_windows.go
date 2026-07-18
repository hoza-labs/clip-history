//go:build windows

package monitor

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	// wmClose asks the hidden monitor window to close.
	wmClose = 0x0010
	// wmDestroy indicates the hidden monitor window is being destroyed.
	wmDestroy = 0x0002
	// wmClipboardUpdate is delivered when the Windows clipboard contents change.
	wmClipboardUpdate = 0x031D
	// hwndMessage creates a message-only window for clipboard notifications.
	hwndMessage = ^uintptr(2)
)

// point mirrors the Win32 POINT structure used inside msg.
type point struct {
	x int32
	y int32
}

// msg mirrors the Win32 MSG structure read by the worker message loop.
type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}

// wndClassEx mirrors the Win32 WNDCLASSEX structure used to register the worker window.
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
	// user32 loads clipboard and windowing procedures from user32.dll.
	user32 = syscall.NewLazyDLL("user32.dll")
	// kernel32 loads module-handle procedures from kernel32.dll.
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	// procAddClipboardFormatListener registers a window for clipboard update notifications.
	procAddClipboardFormatListener = user32.NewProc("AddClipboardFormatListener")
	// procCreateWindowExW creates the hidden message-only worker window.
	procCreateWindowExW = user32.NewProc("CreateWindowExW")
	// procDefWindowProcW handles messages not processed by wndProc.
	procDefWindowProcW = user32.NewProc("DefWindowProcW")
	// procDestroyWindow destroys the hidden worker window.
	procDestroyWindow = user32.NewProc("DestroyWindow")
	// procDispatchMessageW dispatches messages to wndProc.
	procDispatchMessageW = user32.NewProc("DispatchMessageW")
	// procGetMessageW blocks until the worker receives the next window message.
	procGetMessageW = user32.NewProc("GetMessageW")
	// procPostMessageW posts a close message to the worker window.
	procPostMessageW = user32.NewProc("PostMessageW")
	// procPostQuitMessage exits the worker message loop.
	procPostQuitMessage = user32.NewProc("PostQuitMessage")
	// procRegisterClassExW registers the worker window class.
	procRegisterClassExW = user32.NewProc("RegisterClassExW")
	// procRemoveClipboardFormatListener unregisters clipboard notifications for the worker window.
	procRemoveClipboardFormatListener = user32.NewProc("RemoveClipboardFormatListener")
	// procTranslateMessage translates virtual-key messages before dispatch.
	procTranslateMessage = user32.NewProc("TranslateMessage")
	// procGetModuleHandleW returns the module handle for this process.
	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	// monitorWindowProc is the syscall callback used by the hidden worker window.
	monitorWindowProc = syscall.NewCallback(wndProc)
	// monitorWorkers maps hidden window handles to worker instances for wndProc callbacks.
	monitorWorkers sync.Map
)

// worker represents one clipboard monitor instance with its own hidden message window.
type worker struct {
	hwnd      uintptr
	logger    workerLogger
	statePath string
	stopPath  string
	startedAt time.Time
}

// workerLogger is the logging behavior required by RunWorker.
type workerLogger interface {
	// Info records worker lifecycle and clipboard update messages.
	Info(string)
}

// RunWorker creates the hidden clipboard listener and blocks until a stop signal or error occurs.
func RunWorker(logger workerLogger, statePath, stopPath string) error {
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
	w := &worker{
		hwnd:      hwnd,
		logger:    logger,
		statePath: statePath,
		stopPath:  stopPath,
		startedAt: now,
	}
	monitorWorkers.Store(hwnd, w)
	defer monitorWorkers.Delete(hwnd)
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
	go w.heartbeat(done)

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

// heartbeat refreshes the worker state file until done is closed or the stop file appears.
func (w *worker) heartbeat(done <-chan struct{}) {
	ticker := time.NewTicker(DefaultHeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if _, err := os.Stat(w.stopPath); err == nil {
				w.logger.Info("stop signal detected; stopping clipboard monitor")
				procPostMessageW.Call(w.hwnd, wmClose, 0, 0)
				return
			}

			now := time.Now().UTC()
			_ = SaveState(w.statePath, State{
				PID:         os.Getpid(),
				StartedAt:   w.startedAt,
				HeartbeatAt: now,
			})
		}
	}
}

// wndProc handles clipboard, close, and destroy messages for the hidden worker window.
func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	value, _ := monitorWorkers.Load(hwnd)
	w, _ := value.(*worker)

	switch msg {
	case wmClipboardUpdate:
		if w != nil {
			now := time.Now().UTC()
			_ = SaveState(w.statePath, State{
				PID:         os.Getpid(),
				StartedAt:   w.startedAt,
				HeartbeatAt: now,
			})
			w.logger.Info("clipboard update detected")
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
