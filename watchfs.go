package main

import (
    "os"
    "log"
    "fmt"
    "os/exec"
    "path/filepath"
    "strings"
    inotify "code.google.com/p/go.exp/inotify"
)

const (
    QUIT = 100
    IN_CREATE = 200
    IN_DELETE = 300
    IN_CLOSE_WRITE = 400
    IN_MOVE_TO = 500
)

var (
    watchdir = ""
    syncdir = ""
)

type Event struct {
    typecode uint32
    src string
}

type Fileop struct {
    srcfile string
    destfile string
}

func (self *Fileop) tobytearray(src string) []byte {
    return []byte(src)
}

func (self *Fileop) createfile() {
   self.destfile = filepath.Join(syncdir, string(self.tobytearray(self.srcfile)[len(watchdir):]))
   if self.exists(self.destfile) {
       if err := os.Remove(self.destfile); err != nil {
           log.Println("destfile already existed, but delete failed!", self.destfile)
           return
       }
   }

   dirname := filepath.Dir(self.destfile)
   log.Println("dirname", dirname)
   if existed := self.exists(dirname); existed != true {
        log.Printf("create dir: %s\n", dirname)
        Popen("mkdir", "-p", dirname)
   }

   Popen("cp", self.srcfile, self.destfile)
   log.Printf("copy %s to %s ok\n", self.srcfile, self.destfile)
}

func (self *Fileop) removefile() {
    self.destfile = filepath.Join(syncdir, string(self.tobytearray(self.srcfile)[len(watchdir):]))
    Popen("rm", "-f", self.destfile)
    log.Printf("delete old file:%s ok\n", self.destfile)
}

func (op *Fileop) exists(src string) bool {
    _, err := os.Stat(src)
    return err == nil || os.IsExist(err)
}

func Popen(cmd string,  arg ...string) {
    commander := exec.Command(cmd, arg...)
    err := commander.Run()
    if err != nil {
        log.Printf("Popen failed, cmd: %s \n", cmd + " " + strings.Join(arg, ""))
    }
}

func wrapper(src string, opfunc func()) {
    basename := strings.ToLower(filepath.Base(src))
    if !strings.HasPrefix(basename, ".") && !strings.HasSuffix(basename, ".tmp") && strings.LastIndex(basename, ".") > -1 {
        opfunc()
    }else{
        log.Printf("%s dont fit requirements\n", src)
    }
}

func dosynctask(queue chan *Event, exitchannel chan int) {
    log.Println("fs syncer already started!")
    for {
        select {
            case evt := <- queue:
                switch evt.typecode {
                    case inotify.IN_CLOSE_WRITE:
                        fallthrough
                    case inotify.IN_MOVED_TO:
                        fileop := new(Fileop)
                        fileop.srcfile = evt.src
                        wrapper(evt.src, fileop.createfile)
                    case inotify.IN_DELETE:
                        fileop := new(Fileop)
                        fileop.srcfile = evt.src
                        wrapper(evt.src, fileop.removefile)
                }
            case exitcode := <- exitchannel:
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

    mask := inotify.IN_CLOSE_WRITE | inotify.IN_MOVED_TO | inotify.IN_DELETE
    filepath.Walk(watchdir, func(path string, file os.FileInfo, err error) error {
        if file == nil {
            return err
        }

        if err != nil {
            log.Printf("watch file %s failed, reason:%v\n", path, err)
            return err
        }

        if file.IsDir() {
            if err := watcher.AddWatch(path, mask); err != nil {
                log.Printf("watch %s failed!\n", path)
            }else {
                log.Printf("watch dir %s \n", path)
            }
        }
        return nil
    })

    if err != nil {
        log.Fatal(err)
    }

    queue := make(chan *Event, 1000)
    errchannel := make(chan int)

    go dosynctask(queue, errchannel)
    exitflag := false

    for {
        if exitflag == true {
            break
        }
        select {
            case ev := <- watcher.Event:
                send(queue, ev.Mask, ev.Name)
            case err := <- watcher.Error:
                log.Println("error:", err)
                exitflag = true
        }
    }
}
