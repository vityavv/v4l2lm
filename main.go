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

func main() {
	//modprobe this bitch
	cmd := exec.Command("sudo", "-S", "modprobe", "v4l2loopback", "video_nr=63", "card_label=\"V4L2LM Virtual Camera\"")
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		log.Fatal("Error starting v4l2loopback: ", err.Error())
	}

	started := false
	running := make(chan bool)
	var stdin io.WriteCloser
	var ffmpegCommand *exec.Cmd

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

	fileChooserButton, err := gtk.FileChooserButtonNew("Choose a folder", gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER)

	// no idea what much of this means, copied it from the examples in gotk
	treeView, err := gtk.TreeViewNew()

	renderer, err := gtk.CellRendererTextNew()
	column, err := gtk.TreeViewColumnNewWithAttribute("Filename", renderer, "text", 0)
	treeView.AppendColumn(column)

	listStore, err := gtk.ListStoreNew(glib.TYPE_STRING)
	treeView.SetModel(listStore)
	selection, err := treeView.GetSelection()
	selection.Connect("changed", func(sel *gtk.TreeSelection) {
		//So much of this stuff seems unnecessary. Next time I should use something that isn't as complicated as gtk
		model, iter, ok := sel.GetSelected()
		if !ok {
			log.Fatal("Get selected returned not ok, I guess? Here's the model and iter: ", model, iter)
		}
		value, err := listStore.GetValue(iter, 0)
		videoFilename, err := value.GetString()
		if err != nil {
			log.Fatal("Couldn't get value out of list store with that iter: ", err.Error())
		}
		//Stop current command
		if started {
			stdin.Write([]byte("q\n"))
			<-running
		}
		//TODO be more gentle with errors
		fileContents, err := ioutil.ReadFile(fileChooserButton.GetFilename() + "/" + videoFilename)
		if err != nil {
			log.Fatal("Couldn't read file contents of file (to determine content type)")
		}
		contentType := http.DetectContentType(fileContents)
		if contentType[:len("video")] == "video" {
			ffmpegCommand = exec.Command("ffmpeg", "-stream_loop", "-1", "-re", "-i", fileChooserButton.GetFilename()+"/"+videoFilename, "-f", "v4l2", "-s", "1024x768", "/dev/video63")
		} else if contentType[:len("image")] == "image" {
			//TODO: make it so users can change the size
			//set the size manually because it crashes reading from it if the size of the webcam is constantly changing
			ffmpegCommand = exec.Command("ffmpeg", "-stream_loop", "-1", "-re", "-i", fileChooserButton.GetFilename()+"/"+videoFilename, "-f", "v4l2", "-s", "1024x768", "-vcodec", "rawvideo", "-pix_fmt", "yuv420p", "/dev/video63")
		} else {
			log.Fatal("You chose something that isn't an image or a video")
		}
		ffmpegCommand.Stderr = os.Stderr
		ffmpegCommand.Stdout = os.Stdout
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

	scrolledWindow, err := gtk.ScrolledWindowNew(nil, nil)
	scrolledWindow.Add(treeView)

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

	if err != nil {
		log.Fatal("Unable to create object: ", err.Error())
	}

	box.PackStart(tl, false, false, 0)
	box.PackStart(cfl, false, false, 0)
	box.PackStart(fileChooserButton, false, false, 0)
	box.PackStart(civl, false, false, 0)
	box.PackStart(scrolledWindow, true, true, 0)

	win.SetDefaultSize(800, 600)
	win.ShowAll()
	gtk.Main()
}
