package watchfs

import (
    "os"
    "log"
    "os/exec"
    "strings"
//    "consts"
    "path/filepath"
)

type Fileop struct {
    srcfile string
    destfile string
    filetype int
    watchdir string
    syncdir string
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
   self.destfile = filepath.Join(self.syncdir, string(self.tobytearray(self.srcfile)[len(self.watchdir):]))
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

func (self *Fileop) Removefile() {
    self.destfile = filepath.Join(self.syncdir, string(self.tobytearray(self.srcfile)[len(self.watchdir):]))
    if self.filetype == TYPE_FILE {
        Popen("rm", "-f", self.destfile)
        log.Printf("delete old file:%s ok\n", self.destfile)
    }else {
        Popen("rmdir", self.destfile)
        log.Printf("delete old dir:%s ok\n", self.destfile)
    }
}

func (op *Fileop) exists(src string) bool {
    _, err := os.Stat(src)
    return err == nil || os.IsExist(err)
}

func NewFileop(src, watchdir, syncdir string, filetype int) (fileop *Fileop) {
    fileop = new(Fileop)
    fileop.srcfile = src
    fileop.watchdir = watchdir
    fileop.syncdir = syncdir
    fileop.filetype = filetype
    return fileop
}
