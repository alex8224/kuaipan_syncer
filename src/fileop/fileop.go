package fileop

import (
    "os"
    "log"
    "os/exec"
    "strings"
    "consts"
    "path/filepath"
)

type Fileop struct {
    Srcfile string
    Destfile string
    Filetype int
    Watchdir string
    Syncdir string
}

func Popen(cmd string,  arg ...string) {
    commander := exec.Command(cmd, arg...)
    err := commander.Run()
    if err != nil {
        log.Printf("Popen failed, cmd: %s \n", cmd + " " + strings.Join(arg, ""))
    }
}


func (self *Fileop) tobytearray(src string) []byte {
    return []byte(src)
}

func (self *Fileop) Createfile() {
   self.Destfile = filepath.Join(self.Syncdir, string(self.tobytearray(self.Srcfile)[len(self.Watchdir):]))
   if self.exists(self.Destfile) {
       if err := os.Remove(self.Destfile); err != nil {
           log.Println("destfile already existed, but delete failed!", self.Destfile)
           return
       }
   }

   dirname := filepath.Dir(self.Destfile)
   log.Println("dirname", dirname)
   if existed := self.exists(dirname); existed != true {
        log.Printf("create dir: %s\n", dirname)
        Popen("mkdir", "-p", dirname)
   }

   Popen("cp", self.Srcfile, self.Destfile)
   log.Printf("copy %s to %s ok\n", self.Srcfile, self.Destfile)
}

func (self *Fileop) Removefile() {
    self.Destfile = filepath.Join(self.Syncdir, string(self.tobytearray(self.Srcfile)[len(self.Watchdir):]))
    if self.Filetype == consts.TYPE_FILE {
        Popen("rm", "-f", self.Destfile)
        log.Printf("delete old file:%s ok\n", self.Destfile)
    }else {
        Popen("rmdir", self.Destfile)
        log.Printf("delete old dir:%s ok\n", self.Destfile)
    }
}

func (op *Fileop) exists(src string) bool {
    _, err := os.Stat(src)
    return err == nil || os.IsExist(err)
}

func New(src, watchdir, syncdir string, filetype int) (fileop *Fileop) {
    fileop = new(Fileop)
    fileop.Srcfile = src
    fileop.Watchdir = watchdir
    fileop.Syncdir = syncdir
    fileop.Filetype = filetype
    return fileop
}
