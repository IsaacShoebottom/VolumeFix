package main

import (
	"fmt"
	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
	"github.com/go-ole/go-ole"
	"github.com/moutend/go-hook/pkg/keyboard"
	"github.com/moutend/go-hook/pkg/types"
	"github.com/moutend/go-wca/pkg/wca"
	"math"
	"os"
	"os/signal"
)

var (
	// DeviceEnumerator Persistent reference to the device enumerator object
	DeviceEnumerator *wca.IMMDeviceEnumerator
	// IMMDevice Persistent reference to the audio endpoint device
	IMMDevice *wca.IMMDevice
	// AudioEndpointVolume Persistent reference to the audio endpoint volume object
	AudioEndpointVolume *wca.IAudioEndpointVolume
	// Volume tracks volume level
	Volume float32
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
	}
}

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("Volume Fix")
	systray.SetTooltip("Volume Fix")
	quit := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		<-quit.ClickedCh
		fmt.Println("Quitting")
		systray.Quit()
	}()
}

func onExit() {
	// clean up here
	os.Exit(0)
}

func run() error {
	// Set up audio endpoint volume
	err := setupEndpointVolume()
	if err != nil {
		return err
	}
	// Defer teardown of COM objects
	defer teardown()

	keyboardChan := make(chan types.KeyboardEvent, 3)
	if err := keyboard.Install(nil, keyboardChan); err != nil {
		return err
	}
	defer func() {
		err := keyboard.Uninstall()
		if err != nil {
			fmt.Println(err)
		}
	}()

	// Systray, run in a separate goroutine to avoid blocking
	go func() {
		systray.Run(onReady, onExit)
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	fmt.Println("start capturing keyboard input")
	for {
		select {
		case <-signalChan:
			fmt.Println("Received shutdown signal")
			return nil
		case key := <-keyboardChan:
			switch key.VKCode {
			case types.VK_VOLUME_UP:
				// Stop key from being passed to the system
				switch key.Message {
				case types.WM_KEYDOWN:
					fmt.Println("Volume up key pressed")
					Volume = getVolume()
				case types.WM_KEYUP:
					fmt.Println("Volume up key released")
					setVolume(Volume + 0.01)
				}
			case types.VK_VOLUME_DOWN:
				switch key.Message {
				case types.WM_KEYDOWN:
					fmt.Println("Volume down key pressed")
					Volume = getVolume()
				case types.WM_KEYUP:
					fmt.Println("Volume down key released")
					setVolume(Volume - 0.01)
				}
			}
		}
	}
}

func getVolume() float32 {
	var volume float32
	err := AudioEndpointVolume.GetMasterVolumeLevelScalar(&volume)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Current volume is ", volume)
	return volume
}

func setVolume(volume float32) {
	precision := float64(volume)
	// Clamp volume between 0 and 1
	precision = math.Max(0, math.Min(precision, 1))
	// Round to 2 decimal places
	precision = math.Round(precision*100) / 100
	volume = float32(precision)

	fmt.Println("Setting volume to ", volume)
	err := AudioEndpointVolume.SetMasterVolumeLevelScalar(volume, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func setupEndpointVolume() error {
	// Initialize COM library
	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return err
	}
	// Get Device Enumerator
	if err := wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &DeviceEnumerator); err != nil {
		return err
	}
	// Get default audio endpoint from the device enumerator
	if err := DeviceEnumerator.GetDefaultAudioEndpoint(wca.ERender, wca.EConsole, &IMMDevice); err != nil {
		return err
	}
	// Activate the audio endpoint volume object
	if err := IMMDevice.Activate(wca.IID_IAudioEndpointVolume, wca.CLSCTX_ALL, nil, &AudioEndpointVolume); err != nil {
		return err
	}
	return nil
}

func teardown() {
	// Release the audio endpoint volume object
	if AudioEndpointVolume != nil {
		AudioEndpointVolume.Release()
	}
	// Release the audio endpoint device
	if IMMDevice != nil {
		IMMDevice.Release()
	}
	// Release the device enumerator
	if DeviceEnumerator != nil {
		DeviceEnumerator.Release()
	}
	// Un-initialize COM library
	ole.CoUninitialize()
}
