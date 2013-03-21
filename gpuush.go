/** @license
 * gpuush <https://github.com/xthexder/gpuush/>
 * License: MIT
 * Author: Jacob Wirth
 */

package main

import (
	"bytes"
	"crypto/md5"
	"flag"
	"fmt"
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var session string

func notify(msg string) {
	fmt.Println(msg)
	sh := exec.Command("notify-send", "-i", "/usr/local/share/gpuush/icon.png", "gpuush", msg)
	err := sh.Start()
	if err != nil {
		fmt.Println(err)
	} else {
		err = sh.Wait()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func copy(msg string) {
	sh := exec.Command("xclip", "-selection", "clipboard")
	stdin, err := sh.StdinPipe()
	if err != nil {
		fmt.Println(err)
	} else {
		err := sh.Start()
		if err != nil {
			fmt.Println(err)
		} else {
			io.WriteString(stdin, msg)
			stdin.Close()
			err = sh.Wait()
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func login(email, pass string) bool {
	r, err := http.PostForm("http://puush.me/api/auth", url.Values{"e": {email}, "p": {pass}, "z": {"poop"}}) // Don't ask me why this is in the protocol...
	if err != nil {
		fmt.Println(err)
		return false
	}
	body, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	info := strings.Split(string(body), ",")
	if info[0] == "1" {
		session = info[1]
		return true
	} else {
		notify("Login failed:" + string(body))
	}
	return false
}

func uploadFile(filename string) string {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	kwriter, err := w.CreateFormField("k")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	io.WriteString(kwriter, session)

	h := md5.New()
	h.Write(file)

	cwriter, err := w.CreateFormField("c")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	io.WriteString(cwriter, fmt.Sprintf("%x", h.Sum(nil)))

	zwriter, err := w.CreateFormField("z")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	io.WriteString(zwriter, "poop") // They must think their protocol is shit

	fwriter, err := w.CreateFormFile("f", filename)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	fwriter.Write(file)

	w.Close()

	req, err := http.NewRequest("POST", "http://puush.me/api/up", buf)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	body, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	info := strings.Split(string(body), ",")
	if info[0] == "0" {
		return info[1]
	} else {
		notify("Upload failed:" + string(body))
	}
	return ""
}

func takeScreenshot() {
	filename := "/tmp/gpuush" + strconv.FormatInt(time.Now().Unix(), 10) + ".png"
	fmt.Println("Taking screenshot:", filename)
	sh := exec.Command("import", filename)
	err := sh.Start()
	if err != nil {
		fmt.Println(err)
	} else {
		err = sh.Wait()
		if err != nil {
			fmt.Println(err)
		} else {
			result := uploadFile(filename)
			if len(result) > 0 {
				notify(result)
				copy(result)
			} else {
				notify("Upload failed")
			}
		}
	}
}

var conf Config

type Config struct {
	Email string
	Pass  string
}

var background bool
var screenshot bool

func init() {
	flag.BoolVar(&background, "background", false, "Open as a background process")
	flag.BoolVar(&screenshot, "screenshot", false, "Take a screenshot")
}

func main() {
	flag.Parse()

func main() {
	file, err := ioutil.ReadFile("~/.gpuush")
	if err != nil {
		fmt.Println("Config file error:", err)
		return
	}
	err = json.Unmarshal(file, &conf)
	if err != nil {
		fmt.Println("Config file error:", err)
		return
	}

	success := login(conf.Email, conf.Pass)
	if !success {
		return
	}

	if background {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)

		gtk.Init(&os.Args)
		glib.SetApplicationName("gpuush")
		defer gtk.MainQuit()

		si := gtk.NewStatusIconFromFile("/usr/local/share/gpuush/icon.png")
		si.SetTitle("gpuush")
		si.SetTooltipMarkup("gpuush")

		nm := gtk.NewMenu()

		mi := gtk.NewMenuItemWithLabel("Take Screenshot")
		mi.Connect("activate", func() {
			go takeScreenshot()
		})
		nm.Append(mi)
		nm.ShowAll()

		mi = gtk.NewMenuItemWithLabel("Quit")
		mi.Connect("activate", func() {
			quit <- syscall.SIGINT
		})
		nm.Append(mi)
		nm.ShowAll()

		si.Connect("popup-menu", func(cbx *glib.CallbackContext) {
			nm.Popup(nil, nil, gtk.StatusIconPositionMenu, si, uint(cbx.Args(0)), uint32(cbx.Args(1)))
		})

		go gtk.Main()

		for {
			select {
			case <-quit:
				return
			}
		}
	} else if screenshot {
		takeScreenshot()
	} else {
		result := uploadFile(flag.Arg(0))
		if len(result) > 0 {
			notify(result)
			copy(result)
		} else {
			notify("Upload failed")
		}
	}
}
