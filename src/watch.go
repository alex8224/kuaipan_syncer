package main

import (
	inotify "code.google.com/p/go.exp/inotify"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"watchfs"
)

var (
	watchdir = ""
	syncdir  = ""
    SUFFIXS = []string{".tmp", ".crdownload"}
)


type Event struct {
	typecode uint32
	src      string
}

func insuffixs(basename string) bool {
    for _, suffix := range(SUFFIXS) {
        if strings.HasSuffix(basename, suffix) {
            return true
        }
    }

    return false
}

func wrapper(src string, opfunc func()) {
	basename := strings.ToLower(filepath.Base(src))
	if !strings.HasPrefix(basename, ".") && !insuffixs(basename) && strings.LastIndex(basename, ".") > -1 {
		opfunc()
	} else {
		log.Printf("%s dont fit requirements\n", src)
	}
}

func dosynctask(queue chan *Event, exitchannel chan int) {
	log.Println("fs syncer already started!")
	for {
		select {
		case evt := <-queue:
			switch evt.typecode {
			case inotify.IN_CLOSE_WRITE:
				fallthrough
			case inotify.IN_MOVED_TO:
				fileop := watchfs.NewFileop(evt.src, watchdir, syncdir, watchfs.TYPE_FILE)
				wrapper(evt.src, fileop.Createfile)
			case watchfs.IN_MOVED_DIR_FROM:
				fallthrough
			case watchfs.IN_DELETE_DIR:
				fileop := watchfs.NewFileop(evt.src, watchdir, syncdir, watchfs.TYPE_DIR)
				fileop.Removefile()
			case inotify.IN_MOVED_FROM:
				fallthrough
			case inotify.IN_DELETE:
				fileop := watchfs.NewFileop(evt.src, watchdir, syncdir, watchfs.TYPE_FILE)
				wrapper(evt.src, fileop.Removefile)
			case watchfs.IN_MOVED_DIR_TO:
				log.Println("move to dir", evt.src)
				filepath.Walk(evt.src, func(pathname string, file os.FileInfo, err error) error {
					if err != nil {
						return err
					}

					if !file.IsDir() {
						fileop := watchfs.NewFileop(pathname, watchdir, syncdir, watchfs.TYPE_FILE)
						wrapper(pathname, fileop.Createfile)
					}
					return nil
				})
			}
		case exitcode := <-exitchannel:
			log.Println("exit do sync task", exitcode)
			break
		}
	}
}

func send(queue chan *Event, typecode uint32, pathname string) {
	event := new(Event)
	event.typecode = typecode
	event.src = pathname
	queue <- event
}

func addwatch(pathname string, watcher *inotify.Watcher) {
	filepath.Walk(pathname, func(path string, file os.FileInfo, err error) error {
		if file == nil {
			return err
		}

		if err != nil {
			log.Printf("watch file %s failed, reason:%v\n", path, err)
			return err
		}

		if file.IsDir() {
			if err := watcher.AddWatch(path, watchfs.MASK); err != nil {
				log.Printf("watch %s failed!\n", path)
			} else {
				log.Printf("watch dir %s \n", path)
			}
		}
		return nil
	})
}

func removewatch(pathname string, watcher *inotify.Watcher) {
	if err := watcher.RemoveWatch(pathname); err != nil {
		log.Printf("remove watch %s failed! the reason is:%v\n", pathname, err)
	} else {
		log.Printf("remove watch %s ok!\n", pathname)
	}
}

func init() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s watchdir syncdir\n", os.Args[0])
		os.Exit(1)
	}

	watchdir, syncdir = os.Args[1], os.Args[2]
}

func main() {
	watcher, err := inotify.NewWatcher()

	if err != nil {
		log.Fatal(err)
	}

	addwatch(watchdir, watcher)

	if err != nil {
		log.Fatal(err)
	}

	queue := make(chan *Event, 1000)
	errchannel := make(chan int)

	go dosynctask(queue, errchannel)
	exitflag := false

	for {
		if exitflag == true {
			watcher.Close()
			break
		}
		select {
		case ev := <-watcher.Event:
			switch ev.Mask {
			case watchfs.IN_CREATE_DIR:
				addwatch(ev.Name, watcher)
			}
			send(queue, ev.Mask, ev.Name)
		case err := <-watcher.Error:
			log.Println("error:", err)
			exitflag = true
		}
	}
}
