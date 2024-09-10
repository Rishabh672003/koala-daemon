package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	player "koalad/server/music"

	dbus "github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	beep "github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
)

const intro = `
<node>
	<interface name="com.github.rishabh.Koalad">
		<method name="PlayMusic">
			<arg direction="in" type="s"/>
		</method>
		<method name="SetVolume">
			<arg direction="in" type="d"/>
		</method>
		<method name="PauseorUnpause">
			<arg direction="out" type="s"/>
		</method>
	</interface>` + introspect.IntrospectDataString + `</node> `

type foo struct {
	musicMD      *player.MusicMetaData
	musicState   *player.MusicState
	speakerReady bool
	mutex        sync.Mutex
}

func initilizeFoo(f *foo, filepath string) {
	f.musicMD = player.GetMetadata(filepath)
	streamer, format := player.GetStreamer(*f.musicMD)
	f.musicState = player.NewMState(filepath, format, streamer)

	f.mutex.Lock()
	defer f.mutex.Unlock()

	if !f.speakerReady {
		err := speaker.Init(beep.SampleRate(f.musicState.SampleRate), f.musicState.SampleRate.N(time.Second/10)) // 1/10 second buffer
		if err != nil {
			log.Fatal("Error initializing speaker: ", err)
		}
		f.speakerReady = true
		fmt.Println("Speaker initialized.")
	}
}

func (f *foo) PlayMusic(filepath string) (string, *dbus.Error) {
	// Stop the current music state if it exists
	f.mutex.Lock()
	if f.musicState != nil {
		f.musicState.Stop()
		f.musicMD = nil
	}
	f.mutex.Unlock()
	initilizeFoo(f, filepath)

	go func() {
		f.musicState.Play(*f.musicMD)
	}()

	return "music played successfully", nil
}

func (f *foo) SetVolume(delta float64) (string, *dbus.Error) {
	if f.musicState == nil {
		return "", dbus.NewError("com.github.rishabh.Koalad.Error", []interface{}{"No music is currently playing"})
	}
	f.musicState.SetVolume(delta)
	return "volume setted", nil
}

func (f *foo) PauseorUnpause() (string, *dbus.Error) {
	if f.musicState == nil {
		return "", dbus.NewError("com.github.rishabh.Koalad.Error", []interface{}{"No music is currently playing"})
	}
	f.musicState.PauseorUnpause()
	return "paused or unpause", nil
}

func main() {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	f := &foo{}

	conn.Export(f, "/com/github/rishabh/Koalad", "com.github.rishabh.Koalad")
	conn.Export(introspect.Introspectable(intro), "/com/github/rishabh/Koalad",
		"org.freedesktop.DBus.Introspectable")
	reply, err := conn.RequestName("com.github.rishabh.Koalad",
		dbus.NameFlagDoNotQueue)
	if err != nil {
		log.Fatal(err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		fmt.Fprintln(os.Stderr, "name already taken")
		os.Exit(1)
	}
	fmt.Println("Listening on com.github.rishabh.Koalad / /com/github/rishabh/Koalad ...")
	select {}
}
