package toolbox

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	"io"
	"strings"
	"time"
)

type SSHShellClient struct {
	SSHClient
	shellSession     *ssh.Session
	startReadChannel bool
	shellOK          bool
}

type ptyRequestMsg struct {
	Term     string
	Columns  uint32
	Rows     uint32
	Width    uint32
	Height   uint32
	Modelist string
}

type TerminalSize struct {
	Cols   int `json:"cols"`
	Rows   int `json:"rows"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

func (this_ *SSHShellClient) changeSize(terminalSize TerminalSize) (err error) {

	if this_.shellSession == nil {
		return
	}
	if terminalSize.Cols > 0 && terminalSize.Rows > 0 {
		err = this_.shellSession.WindowChange(terminalSize.Rows, terminalSize.Cols)
		if err != nil {
			this_.Logger.Error("SSH Shell Session Window Change error", zap.Error(err))
			return
		}
	}
	if terminalSize.Width > 0 && terminalSize.Height > 0 {
		err = this_.shellSession.WindowChange(terminalSize.Height, terminalSize.Width)
		if err != nil {
			this_.Logger.Error("SSH Shell Session Window Change error", zap.Error(err))
			return
		}
	}
	return
}

func (this_ *SSHShellClient) closeSession(session *ssh.Session) {
	if session == nil {
		return
	}
	err := session.Close()
	if err != nil {
		fmt.Println("SSH Shell Session close error", err)
		return
	}
}
func (this_ *SSHShellClient) startShell(terminalSize TerminalSize) (err error) {
	this_.shellOK = false
	this_.startReadChannel = false
	defer func() {
		if x := recover(); x != nil {
			this_.Logger.Error("SSH Shell Start Error", zap.Any("err", x))
			return
		}
		this_.shellSession = nil
	}()
	if this_.shellSession != nil {
		err = this_.shellSession.Close()
		if err != nil {
			this_.Logger.Error("SSH Shell Shell Session Close Error", zap.Error(err))
		}
		this_.shellSession = nil
	}
	err = this_.initClient()
	if err != nil {
		this_.Logger.Error("createShell initClient error", zap.Error(err))
		this_.WSWriteError("SSH客户端创建失败:" + err.Error())
		return
	}

	this_.shellSession, err = this_.sshClient.NewSession()
	if err != nil {
		this_.Logger.Error("createShell OpenChannel error", zap.Error(err))
		this_.WSWriteError("SSH会话创建失败:" + err.Error())
		return
	}
	defer this_.closeSession(this_.shellSession)

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	var modeList []byte
	for k, v := range modes {
		kv := struct {
			Key byte
			Val uint32
		}{k, v}
		modeList = append(modeList, ssh.Marshal(&kv)...)
	}
	modeList = append(modeList, 0)
	req := ptyRequestMsg{
		Term:     "xterm",
		Modelist: string(modeList),
	}
	if terminalSize.Cols > 0 && terminalSize.Rows > 0 {
		req.Columns = uint32(terminalSize.Cols)
		req.Rows = uint32(terminalSize.Rows)
	}
	if terminalSize.Width > 0 && terminalSize.Height > 0 {
		req.Width = uint32(terminalSize.Width)
		req.Height = uint32(terminalSize.Height)
	}
	_, err = this_.shellSession.SendRequest("pty-req", true, ssh.Marshal(&req))
	if err != nil {
		this_.Logger.Error("createShell SendRequest pty-req error", zap.Error(err))
		return
	}

	this_.shellOK, err = this_.shellSession.SendRequest("shell", true, nil)
	if !this_.shellOK || err != nil {
		if err != nil {
			err = errors.New("ssh shell send request fail")
		}
		this_.Logger.Error("createShell SendRequest shell error", zap.Error(err))
		this_.WSWriteError("SSH Shell创建失败:" + err.Error())
		return
	}

	for {
		if !this_.startReadChannel {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		var bs = make([]byte, 1024)
		var n int
		var reader io.Reader
		reader, err = this_.shellSession.StdoutPipe()
		if err != nil {
			this_.Logger.Error("SSH Shell Stderr Pipe Error", zap.Error(err))
		}
		n, err = reader.Read(bs)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			this_.Logger.Error("SSH Shell 消息读取异常", zap.Error(err))
			//this_.WSWriteError("SSH Shell 消息读取失败:" + err.Error())
			continue
		}
		bs = bs[0:n]
		this_.WSWrite(bs)
	}
}

func (this_ *SSHShellClient) start() {
	SSHShellCache[this_.Token] = this_
	go this_.ListenWS(this_.onEvent, this_.onMessage, this_.CloseClient)
	this_.WSWriteEvent("shell ready")
}

func (this_ *SSHShellClient) onEvent(event string) {
	var err error
	this_.Logger.Info("SSH Shell On Event:", zap.Any("event", event))

	if strings.HasPrefix(event, "shell start") {
		jsonStr := event[len("shell start"):]
		var terminalSize *TerminalSize
		if jsonStr != "" {
			_ = json.Unmarshal([]byte(jsonStr), &terminalSize)
		}
		go func() {
			err = this_.startShell(*terminalSize)
			if err != nil {
				this_.Logger.Error("SSH Shell startShell error", zap.Error(err))
			}
		}()
		for {
			time.Sleep(100 * time.Millisecond)
			if err == nil || this_.shellOK {
				break
			}
		}
		if err != nil {
			return
		}
		this_.WSWriteEvent("shell created")
		time.Sleep(1000 * time.Millisecond)
		this_.startReadChannel = true
		return
	} else if strings.HasPrefix(event, "change size") {
		jsonStr := event[len("change size"):]
		var terminalSize *TerminalSize
		err = json.Unmarshal([]byte(jsonStr), &terminalSize)
		if err != nil {
			return
		}
		err = this_.changeSize(*terminalSize)
	}
	switch strings.ToLower(event) {
	}
}

func (this_ *SSHShellClient) onMessage(bs []byte) {
	defer func() {
		if x := recover(); x != nil {
			this_.Logger.Error("SSH Shell Write Error", zap.Any("err", x))
			return
		}
	}()
	if this_.shellSession == nil {
		return
	}
	var err error
	var writer io.Writer
	writer, err = this_.shellSession.StdinPipe()
	if err != nil {
		this_.Logger.Error("SSH Shell Stderr Pipe Error", zap.Error(err))
	}

	_, err = writer.Write(bs)
	if err != nil {
		this_.WSWriteError("SSH Shell Write失败:" + err.Error())
	}
}