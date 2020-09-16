package main

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	//To check content type
	"net/http"
)

type Playing struct {
	ContentType string
	Filename    string
}

func main() {
	//modprobe this bitch
	//TODO check if video driver exists
	v4l2Enable := exec.Command("sudo", "-S", "modprobe", "v4l2loopback", "video_nr=63", "card_label=\"V4L2LM Virtual Camera\"")
	v4l2Enable.Stderr = os.Stderr
	v4l2Enable.Stdin = os.Stdin
	err := v4l2Enable.Run()
	if err != nil {
		log.Fatal("Error starting v4l2loopback: ", err.Error())
	}

	started := false
	running := make(chan bool)
	var stdin io.WriteCloser
	var ffmpegCommand *exec.Cmd
	var nowPlaying Playing

	gtk.Init(nil)

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create gtk window: ", err.Error())
	}
	win.SetTitle("V4L2LM")
	win.Connect("destroy", func() {
		if started {
			stdin.Write([]byte("q\n"))
			<-running
		}
		gtk.MainQuit()
	})

	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	if err != nil {
		log.Fatal("Unable to create box: ", err.Error())
	}
	win.Add(box)

	tl, err := gtk.LabelNew("")
	tl.SetMarkup("<big><b>V4L2 Loopback Manager</b></big>")
	cfl, err := gtk.LabelNew("Choose a folder")
	civl, err := gtk.LabelNew("Choose an image or a video to display")
	sl, err := gtk.LabelNew("Settings")
	fsl, err := gtk.LabelNew("Fit style: ")

	fileChooserButton, err := gtk.FileChooserButtonNew("Choose a folder", gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER)

	fitStyleComboBox, err := gtk.ComboBoxTextNew()
	fitStyleComboBox.Append("Stretch", "Stretch")
	fitStyleComboBox.Append("Crop", "Crop")
	fitStyleComboBox.Append("Letterbox", "Letterbox")

	fitStyleBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	fitStyleBox.PackStart(fsl, false, false, 0)
	fitStyleBox.PackStart(fitStyleComboBox, false, false, 0)

	// no idea what much of this means, copied it from the examples in gotk
	treeView, err := gtk.TreeViewNew()

	renderer, err := gtk.CellRendererTextNew()
	column, err := gtk.TreeViewColumnNewWithAttribute("Filename", renderer, "text", 0)
	treeView.AppendColumn(column)

	listStore, err := gtk.ListStoreNew(glib.TYPE_STRING)
	treeView.SetModel(listStore)

	scrolledWindow, err := gtk.ScrolledWindowNew(nil, nil)
	scrolledWindow.Add(treeView)

	selection, err := treeView.GetSelection()
	selection.Connect("changed", func(sel *gtk.TreeSelection) {
		//So much of this stuff seems unnecessary. Next time I should use something that isn't as complicated as gtk
		_, iter, ok := sel.GetSelected()
		if !ok {
			//Sometimes it gets to an iter that doesn't exist
			//because the list has already changed
			//So we don't crash, we just report it, and hope that nothing bad has happened
			log.Println("Get selected returned not ok. This is probably because you switched folders, and can be ignored.")
			return
		}
		value, err := listStore.GetValue(iter, 0)
		videoFilename, err := value.GetString()
		if err != nil {
			log.Fatal("Couldn't get value out of list store with that iter: ", err.Error())
		}
		_, err = os.Stat(fileChooserButton.GetFilename() + "/" + videoFilename)
		if os.IsNotExist(err) {
			// we ignore files that do not exist but are chosen because the change event fires
			// when the folder changes, several times (for some reason, "it may occasionally be emitted when nothing has happened" - the documentation)
			return
		}
		//TODO be more gentle with errors
		fileContents, err := ioutil.ReadFile(fileChooserButton.GetFilename() + "/" + videoFilename)
		if err != nil {
			log.Fatal("Couldn't read file contents of file (to determine content type)")
		}
		contentType := http.DetectContentType(fileContents)
		//Stop current command
		if contentType[:len("video")] != "video" && contentType[:len("image")] != "image" {
			dialog := gtk.MessageDialogNew(win, gtk.DIALOG_DESTROY_WITH_PARENT, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "You chose something that isn't an image or a video!")
			dialog.Run()
			dialog.Destroy()
			return
		}
		if started {
			stdin.Write([]byte("q\n"))
			<-running
		}
		//Construct the two scaling arguments based on chosen fit style

		fitStyle := fitStyleComboBox.GetActiveText()
		nowPlaying = Playing{
			ContentType: contentType,
			Filename:    fileChooserButton.GetFilename() + "/" + videoFilename,
		}
		ffmpegCommand = createCommand(fitStyle, nowPlaying)
		stdin, err = ffmpegCommand.StdinPipe()
		if err != nil {
			log.Fatal("error getting stdin: ", err)
		}
		go func() {
			ffmpegCommand.Run()
			running <- false
		}()
		started = true
	})

	fileChooserButton.Connect("file-set", func(fcb *gtk.FileChooserButton) {
		listStore.Clear()
		fileInfos, err := ioutil.ReadDir(fcb.GetFilename())
		if err != nil {
			log.Fatal("Could not read that directory")
		}
		for _, fileInfo := range fileInfos {
			if !fileInfo.IsDir() {
				appendIter := listStore.Append()
				listStore.SetValue(appendIter, 0, fileInfo.Name())
			}
		}
	})

	fitStyleComboBox.Connect("changed", func(fscb *gtk.ComboBoxText) {
		if !started {
			return
		}
		stdin.Write([]byte("q\n"))
		<-running
		fitStyle := fitStyleComboBox.GetActiveText()
		ffmpegCommand = createCommand(fitStyle, nowPlaying)
		stdin, err = ffmpegCommand.StdinPipe()
		if err != nil {
			log.Fatal("Error getting stdin: ", err)
		}
		go func() {
			ffmpegCommand.Run()
			running <- false
		}()
	})

	if err != nil {
		log.Fatal("Unable to create object: ", err.Error())
	}

	box.PackStart(tl, false, false, 0)
	box.PackStart(cfl, false, false, 0)
	box.PackStart(fileChooserButton, false, false, 0)
	box.PackStart(civl, false, false, 0)
	box.PackStart(scrolledWindow, true, true, 0)
	box.PackStart(sl, false, false, 0)
	box.PackStart(fitStyleBox, false, false, 10)

	win.SetDefaultSize(800, 600)
	win.ShowAll()
	gtk.Main()
}

func createCommand(fitStyle string, nowPlaying Playing) (command *exec.Cmd) {
	arg1, arg2 := "-s", "1024x768"
	if fitStyle == "Letterbox" {
		arg1, arg2 = "-vf", "scale=1024:768:force_original_aspect_ratio=decrease,pad=1024:768:(ow-iw)/2:(oh-ih)/2"
	} else if fitStyle == "Crop" {
		arg1, arg2 = "-vf", "scale=1280:720:force_original_aspect_ratio=increase,crop=1280:720"
	}
	if nowPlaying.ContentType[:len("video")] == "video" {
		command = exec.Command("ffmpeg", "-stream_loop", "-1", "-re", "-i", nowPlaying.Filename, "-f", "v4l2", arg1, arg2, "-pix_fmt", "yuv420p", "/dev/video63")
	} else if nowPlaying.ContentType[:len("image")] == "image" {
		//TODO: make it so users can change the size
		//set the size manually because it crashes reading from it if the size of the webcam is constantly changing
		command = exec.Command("ffmpeg", "-loop", "true", "-re", "-i", nowPlaying.Filename, "-f", "v4l2", arg1, arg2, "-pix_fmt", "yuv420p", "/dev/video63")
	}
	command.Stderr = os.Stderr
	command.Stdout = os.Stdout
	return
}
