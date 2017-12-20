package watch

import (
	"github.com/astaxie/beego/logs"
	"os"
	"strings"
	"os/signal"
	"github.com/codeskyblue/kexec"
	"syscall"
	"path/filepath"
	"io/ioutil"
	"strconv"
	"fmt"
)

func Defersig() {
	log.Debug("默认sigfn")
}

type watch struct {
	programCmd string
	watchArg string
	programArg string
	erverprocess *kexec.KCommand
	openprogram int
	pidfile string
	sigfn func()
	sigdaemonfn func()
}

func (this *watch) Start() {
	cmd := kexec.CommandString(this.programCmd + " " + this.programArg)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	err :=cmd.Wait()
	if err != nil {
		log.Debug(err.Error())
	}
}


func (this *watch) ForErver() {
	cmd := kexec.CommandString(this.programCmd + " " + "erver")
	//cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true,}
	log.Debug(this.programCmd + " " + "erver")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	log.Debug("父进程关闭")
	os.Exit(0)
}

func (this *watch) Erver() {
	this.RecordPid()
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGTERM)
		for {
			s := <-c
			this.openprogram = 1
			this.erverprocess.Terminate(syscall.SIGKILL)
			log.Debug(strconv.Itoa(os.Getpid()), strconv.Itoa(os.Getppid()), strconv.Itoa(os.Getgid()))
			log.Debug("forerver get signal: ", s.String())
			this.sigfn()
			os.Exit(0)
		}
	}()
		for {
			if this.openprogram == 0 {
				cmd := kexec.CommandString(this.programCmd + " " + this.programArg)
				//cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true,}
				this.erverprocess = cmd
				log.Debug(this.programCmd + " " + this.programArg)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err := cmd.Run()
				if err != nil {
					log.Debug(err.Error())
				}
				log.Debug("意外死亡， 重启")
			}else {
				os.Exit(0)
			}
		}
}

func (this *watch) RecordPid() {
	file, err := os.Create(this.pidfile)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	pidInt := os.Getpid()
	log.Debug("pid", pidInt)
	_, err = file.WriteString(strconv.Itoa(pidInt))
	if err != nil {
		panic(err)
	}
}

func (this *watch) Stop() {
	pidByte, err := ioutil.ReadFile(this.pidfile)
	if err != nil {
		panic(err)
	}
	pidInt, err := strconv.Atoi(string(pidByte))
	if err != nil {
		panic(err)
	}
	process, err := os.FindProcess(pidInt)
	if err != nil {
		panic(err)
	}
	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		panic(err)
	}
	log.Debug("程序已关闭")
}

var log *logs.BeeLogger = logs.NewLogger()


func CreateWatch(sigfn func(), sigdaemonfn func()) {
	log.EnableFuncCallDepth(true)
	watchData := new(watch)
	watchData.programCmd = os.Args[0]
	if len(os.Args) >=2 {
		watchData.watchArg = os.Args[1]
		watchData.programArg = strings.Join(os.Args[2:], " ")
	}
	watchData.sigfn = sigfn
	watchData.sigdaemonfn = sigdaemonfn
	watchData.pidfile = "/var/run/" + filepath.Base(watchData.programCmd) + ".pid"

	switch watchData.watchArg {
	case "stop":
		watchData.Stop()
	case "start":
		watchData.Start()
	case "forerver":
		watchData.ForErver()
	case "erver":
		watchData.Erver()
	default:
		go func() {
			c := make(chan os.Signal)
			signal.Notify(c, syscall.SIGTERM)
			for {
				s := <-c
				log.Debug(strconv.Itoa(os.Getpid()), strconv.Itoa(os.Getppid()), strconv.Itoa(os.Getgid()))
				log.Debug("default get signal: ", s.String())
				watchData.sigdaemonfn()
				os.Exit(0)
			}
		}()
		log.Debug("daemon 启动")
		fmt.Println()
	}
}

