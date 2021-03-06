# V4L2 Loopback Manager

This is a program that provides a wrapper around ffmpeg to conveniently pipe images and video to [v4l2loopback](https://github.com/umlaeute/v4l2loopback). You can choose a folder to import images and videos from, and switch between images and videos on the fly. Images and videos piped to v4l2loopback will loop.

## Use cases

Make it seem like you are paying attention by piping an image/video of you looking into the camera to v4l2loopback, letting you do whatever you want without fear of getting caught.

Make it seem like you are exercising in a virtual gym class, when in reality it's just a video of you doing the excercise once, looping over and over again.

Put up cards that say "I'm going to the bathroom" when you need to excuse yourself from class or a meeting.

## Installation

Right now this only supports Linux. Make sure you have v4l2loopback, go, and the gtk3 libraries installed. Run `go get github.com/vityavv/v4l2lm`.

## Usage

Run `v4l2lm` *in the command line* to start up the program. It will ask you for your password because it is enabling v4l2loopback. By default, it enables video device 63, naming it "V4L2LM Virtual Camera." As of now, there is no way to change this. Once you have run it in the command line once, you can run it any way you'd like until the next time you restart your computer.

Once you put in your password and v4l2loopback has been enabled, choose a folder using the button below the label that says "Choose a folder." Once you choose a folder, the list below the label "Choose an image or a video to display" will fill up with all of the files in that folder. If the list is long enough, you will be able to scroll through it. Press on the name of an image or a video to stream it to the v4l2loopback camera.

After you've chosen an image or video to stream, you may open up your consumer (i.e. zoom, google meets, etc.), choose the V4L2LM Virtual Camera, and it should show there. If you change the image or video, it will change what's streaming to the virtual webcam.

Under the `Settings` label, you may choose a different fit style (initially, images and videos of a different aspect ratio are stretched to fit the resolution, but you may letterbox them or crop (see quirks) them to size). Changing the fit style while the program is running is totally okay. You can also change the resolution by entering in different numbers and then pressing the `Set Resolution` button, but know that doing this after selecting an image or video may cause your consumer program (zoom, google meets, etc.) to crash or it may produce a very odd image. You should probably use this feature **before** choosing an image or video to display.

## Quirks and TODOs

Here are some things that happen that I don't want happening in the near future.

- The "Crop" method of fitting is massively fucked up and I don't know why
- Make releases/prebuilt binaries so users don't have to install all of the dependencies
- Windows/Mac support (very far future unless someone else wants to do it haha)
- Perhaps filter out certain file extensions

## Need help?

I understand that this documentation is very sparse and it might be difficult to set up. If you ever need help, don't hesitate to [contact me](https://victor.computer/about)---I can be reached easily on discord.

