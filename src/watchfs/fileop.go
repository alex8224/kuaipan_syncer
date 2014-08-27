package watchfs

import (
    "os"
    "log"
    "os/exec"
    "strings"
    "path/filepath"
    "bytes"
)

type Fileop struct {
    srcfile string
    destfile string
    filetype int
    watchdir string
    syncdir string
}

type StringBuffer struct {
    strbuff *bytes.Buffer
}

func (self *StringBuffer) init() {
    if self.strbuff == nil {
        self.strbuff = new(bytes.Buffer)
    }
}

func (self *StringBuffer) Append(str string) *StringBuffer {
    if _, err := self.strbuff.WriteString(str); err != nil {
        return nil
    }
    return self
}

func (self *StringBuffer) String() string {
    return self.strbuff.String()
}

func (self *StringBuffer) Replace(oldstr, newstr string) string {
    orgistr := self.String()
    replacedstr := strings.Replace(orgistr, oldstr, newstr, -1)
    self.strbuff.Reset()
    self.strbuff.WriteString(replacedstr)
    return replacedstr
}

func NewStringBuffer(initstr string) *StringBuffer {
    sb := &StringBuffer{}
    sb.init()
    sb.Append(initstr)
    return sb
}

func Popen(cmdstr string,  arg ...string) {
    cmd := exec.Command(cmdstr, arg...)
    log.Printf("%s", arg)
    output, err := cmd.CombinedOutput()
    if err != nil {
        log.Printf("Popen failed, cmd: %s, error message:%s \n", cmdstr + " " + strings.Join(arg, " "), string(output))
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

   Popen("cp", "-f", self.srcfile, self.destfile)
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
