// mplayer -vo null url
// vo=null
// ao=pulse
// really-quiet=1
// nolirc=1
// cvlc --play-and-exit --no-video url

package main

import (
	"os"
	"fmt"
	"log"
	"time"
	"path"
	"bytes"
	"os/exec"
	"strings"
	"net/http"
	"io/ioutil"
	"os/signal"
	"math/rand"
	"encoding/json"
	"container/list"
)

type Music struct {
	Title string
	Url string
	Hash string
}

const HTML_INDEX = "index.html"
const WMUS_JS = "wmus.js"
const HISTORY_MAX = 250;

// mplayer
// const MEDIA_PLAYER = "/usr/bin/mplayer"
// const MEDIA_PARAMETERS = "-vo null -really-quiet -ao pulse -nolirc"

// vlc
const MEDIA_PLAYER = "/usr/bin/cvlc"
const MEDIA_PARAMETERS = "--play-and-exit --no-video"

var musicQueue *list.List
var musicMap map[string]int
var historyMap map[string]bool

var musicHistory *list.List
var messageBuffer *list.List
var nowPlaying string
var neverStop bool
var player_errors chan error
var player_quit chan bool
var player_done chan bool
var player_resume chan bool
var player_stopped bool
var player *exec.Cmd
var rng *rand.Rand

func drainchan(commch chan bool) {
	for {
		select {
		case <-commch:
		default:
			return
		}
	}
}

func check_youtube(hash string) (string, string, error) {
	out, err := exec.Command("/usr/bin/env", "python", "pafyurl.py", hash).CombinedOutput()
	if err != nil {
		log.Printf("youtube error: %s", err)
		return "", "", err
	}
	lines := strings.Split(string(out[:]), "\n")
	title := lines[0]
	url := lines[1]
	return title, url, nil
}

func fileoryoutube(filename string) (string, string, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Printf("no such file or directory: %s", filename)
		title, url, err := check_youtube(filename)
		log.Printf("returning title=%s url=%s err=%s", title, url, err)
		return title, url, err
	}
	return path.Base(filename), filename, nil
}

func play(hash string) {
	title, url, err := fileoryoutube(hash)
	if err != nil {
		player_done <- true
		return
	}
	player_args := []string{}
	player_args = append(player_args, strings.Split(MEDIA_PARAMETERS, " ")...)
	player_args = append(player_args, url)
	log.Printf("%s %s", MEDIA_PLAYER, player_args)
	player = exec.Command(MEDIA_PLAYER, player_args...)
	err = player.Start()
	nowPlaying = title
	if err != nil {
		player_errors <- err
	}
	player.Wait()
	nowPlaying = ""
	player_done <- true
}

func queuePlayer() {
	log.Print("player started...")
	for {
		if player_stopped {
			<- player_resume
		}
		e := musicQueue.Front()
		if e == nil && neverStop && musicHistory.Len() > 0 {
			i := rng.Intn(musicHistory.Len())
			for e = musicHistory.Front(); e != nil; e = e.Next() {
				if i == 0 {
					musicQueue.PushBack(e.Value)
					break
				}
				i--
			}
			continue
		}
		if e != nil {
			log.Print("playing ", e.Value)
			musicQueue.Remove(e)
			playstart := time.Now()
			music := e.Value.(Music)
			go play(music.Hash)
			player_control:
			for {
				select {
					case <- player_done:
						log.Print("player exited")
						if time.Since(playstart) > 5 * time.Second {
							log.Printf("%s played more than 5s, adding to history", e.Value.(Music).Title)
							musicMap[e.Value.(Music).Hash]++
							if !historyMap[e.Value.(Music).Hash] {
								historyMap[e.Value.(Music).Hash] = true
							} else {
								// remove all previous titles
								for f := musicHistory.Front(); f != nil; f = f.Next() {
									if f.Value.(Music).Hash == e.Value.(Music).Hash {
										musicHistory.Remove(f);
									}
								}
							}
							// now that they're all gone, put it back on top
							musicHistory.PushBack(e.Value)
							for musicHistory.Len() > HISTORY_MAX {
								x := musicHistory.Front()
								musicHistory.Remove(x)
							}
						}
						break player_control
					case err := <- player_errors:
						log.Print("error reported: ", err)
						messageBuffer.PushBack(fmt.Sprintf("ERROR: %v", err))
					case quit := <- player_quit:
						if quit {
							log.Print("aborting player: ", quit)
							if player != nil && player.Process != nil {
								player.Process.Kill()
							}
							log.Print("killed player")
						}
				}
			}
			log.Print("trying next song")
		}
		time.Sleep(100*time.Millisecond)
	}
}

func jsonList(l *list.List, reverse ...bool) (data []byte, err error) {
	items := make([]Music, l.Len())
	i := 0
	if len(reverse) > 0 {
		for e := l.Back(); e != nil; e = e.Prev() {
			items[i] = e.Value.(Music)
			i++
		}
	} else {
		for e := l.Front(); e != nil; e = e.Next() {
			items[i] = e.Value.(Music)
			i++
		}
	}
	return json.Marshal(items)
}

func listJson(l *list.List, data []byte) (err error) {
	var music_array []Music
	err = json.Unmarshal(data, &music_array);
	if err != nil {
		return err
	}
	l.Init()
	for _, value := range music_array {
		l.PushBack(value)
	}
	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch (r.URL.Path[1:]) {
		case "":
			body, err := ioutil.ReadFile(HTML_INDEX)
			if err != nil {
				fmt.Fprintf(w, "NO %v", err)
			} else {
				fmt.Fprintf(w, string(body))
				log.Print("index")
			}
		case "wmus.js":
			body, err := ioutil.ReadFile(WMUS_JS)
			if err != nil {
				fmt.Fprintf(w, "NO %v", err)
			} else {
				fmt.Fprintf(w, string(body))
				log.Print("wmus.js")
			}
		case "addq":
			// allow youtube userscript to interact
			origin := r.Header.Get("Origin")
			if strings.Index(origin, "http://www.youtube.com") == 0 || strings.Index(origin, "https://www.youtube.com") == 0 {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			v := r.URL.Query()
			hash := v.Get("hash")
			if hash != "" {
				title, url, err := fileoryoutube(hash)
				if err != nil {
					fmt.Fprintf(w, "NO %s", err)
					log.Printf("add failed: %s", err)
					break
				}
				musicQueue.PushFront(Music{title, url, hash})
				fmt.Fprintf(w, "OK %v", title)
				log.Print("add ", title)
			} else {
				fmt.Fprintf(w, "NO")
				log.Print("add failed")
			}
		case "delq":
			v := r.URL.Query()
			hash := v.Get("hash")
			if hash != "" {
				for e := musicQueue.Back(); e != nil; e = e.Prev() {
					if e.Value.(Music).Hash == hash {
						musicQueue.Remove(e)
						log.Printf("delq removed %v", e.Value.(Music))
					}
				}
				fmt.Fprintf(w, "OK removed")
			}
			log.Printf("delq %s", hash)
		case "delh":
			v := r.URL.Query()
			hash := v.Get("hash")
			if hash != "" {
				for e := musicHistory.Back(); e != nil; e = e.Prev() {
					if e.Value.(Music).Hash == hash {
						musicHistory.Remove(e)
						log.Printf("delh removed %v", e.Value.(Music))
					}
				}
				fmt.Fprintf(w, "OK removed")
			}
			log.Printf("delh %s", hash)
		case "nowp":
			if player_stopped {
				fmt.Fprintf(w, "STOPPED")
				break
			} else {
				fmt.Fprintf(w, "OK %s", nowPlaying)
			}
		case "list":
			data, err := jsonList(musicQueue, true)
			if err != nil { fmt.Fprintf(w, "NO %v", err) }
			w.Write(data)
			log.Printf("list %s", data)
		case "msgs":
			data, err:= jsonList(messageBuffer, true)
			if err != nil { fmt.Fprintf(w, "NO %v", err) }
			w.Write(data)
			log.Printf("msgs %s", data)
		case "abrt":
			log.Printf("aborting current player")
			drainchan(player_quit)
			player_quit <- true
			fmt.Fprintf(w, "OK next")
		case "hist":
			data, err := jsonList(musicHistory)
			if err != nil { fmt.Fprintf(w, "NO %v", err) }
			w.Write(data)
			log.Printf("hist %s", data)
		case "stop":
			log.Printf("stopping player")
			player_stopped = true
			drainchan(player_quit)
			player_quit <- true
			neverStop = false
			fmt.Fprintf(w, "OK next")
		case "loop":
			if musicHistory.Len() > 0 {
				log.Printf("looping shuffled history forever")
				player_stopped = false
				player_resume <- true
				neverStop = true
				fmt.Fprintf(w, "OK next")
			} else {
				log.Printf("can't loop without any history")
				fmt.Fprintf(w, "NO no history to play from")
			}
		case "resu":
			log.Printf("resuming player")
			player_stopped = false
			player_resume <- true
			fmt.Fprintf(w, "OK next")
		default:
			fmt.Fprintf(w, "NO %s", r.URL.Path[1:])
	}
}

func saveQueues() {
	var b bytes.Buffer
	data, err := jsonList(musicHistory)
	if err != nil {
		log.Printf("unable to encode history: %s", err)
	}
	b.Write(data)
	err = ioutil.WriteFile("history.json", b.Bytes(), 0644)
	if err != nil {
		log.Printf("error saving history: %s", err)
	}
	b.Reset()
	data, err = jsonList(musicQueue)
	b.Write(data)
	err = ioutil.WriteFile("queue.json", b.Bytes(), 0644)
	if err != nil {
		log.Printf("error saving queue: %s", err)
	}
}

func loadQueues() {
	// music queue
	log.Print("loading queue...")
	b, err := ioutil.ReadFile("queue.json")
	if err != nil {
		log.Printf("error loading queue: %s", err)
	}
	listJson(musicQueue, b)
	// history
	log.Print("loading history...")
	b, err = ioutil.ReadFile("history.json")
	if err != nil {
		log.Printf("error loading history: %s", err)
	}
	listJson(musicHistory, b)
}

func autoSave() {
	for {
		time.Sleep(time.Minute)
		log.Print("autosaving queues")
		saveQueues()
	}
}


func main() {

	player_errors = make(chan error, 10)
	player_quit = make(chan bool, 1)
	player_done = make(chan bool, 1)
	player_resume = make(chan bool, 1)
	player_stopped = false
	musicQueue = list.New()
	musicMap = make(map[string]int)
	historyMap = make(map[string]bool)
	musicHistory = list.New()
	messageBuffer = list.New()
	nowPlaying = ""
	neverStop = true
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	quit_signal := make(chan os.Signal, 1)
	signal.Notify(quit_signal, os.Interrupt)
	go func(){
		for _ = range quit_signal {
			saveQueues()
			os.Exit(0)
		}
	}()

	loadQueues()

	go autoSave()

	log.Print("starting queue player...")
	go queuePlayer()
	http.HandleFunc("/", handler)
	log.Print("starting web server...")
	if len(os.Args) < 2 {
		log.Printf("Usage: %s <host:port>", os.Args[0])
		log.Printf("Ex: %s :8080", os.Args[0])
		log.Printf("Ex: %s :8443 cert.pem key.pem", os.Args[0])
		log.Fatal("No listening socket specified, exiting!")
	}
	hostPort := os.Args[1]
	if len(os.Args) == 4 {
		certFile := os.Args[2]
		keyFile := os.Args[3]
		http.ListenAndServeTLS(hostPort, certFile, keyFile, nil)
	}
	err := http.ListenAndServe(hostPort, nil)
	if err != nil {
		log.Fatal(err)
	}
}

