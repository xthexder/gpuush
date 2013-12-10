gpuush
======

This is a very basic puush.me client for linux written in golang.  
**gpuush** can either use imagemagick's import command to select an area/window to upload or upload a file from command line.

The protocol has been reverse engineering in order to upload files without the official client, which does not support linux.

Usage
-----

Once compiled (via `go build` or `go install`), the executable should be placed in /usr/local/bin/

    gpuush [filename]     Upload any file to puush.me
    gpuush -screenshot    Take a screenshot (for use with hotkeys)
    gpuush -background    Stay open as a tray icon

The icon file is read from /usr/local/share/gpuush/icon.png

To configure the user gpuush will upload as, create the file `~/.gpuush` with the contents:

    {
      "email": "example@example.com",
      "pass": "hunter2"
    }

Some external dependencies that may need to be installed are:

 - imagemagick
 - libnotify
 - xclip
