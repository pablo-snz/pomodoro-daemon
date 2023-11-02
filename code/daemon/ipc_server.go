package daemon

import (
	"code/pomodoro"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type PomodoroIPCServer struct {
	pomodoro *pomodoro.Pomodoro
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewPomodoroIPCServer(pomodoroStates []pomodoro.PomodoroStates) *PomodoroIPCServer {
	ctx, cancel := context.WithCancel(context.Background())
	pom := pomodoro.NewPomodoro(pomodoroStates)
	return &PomodoroIPCServer{
		pomodoro: pom,
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (s *PomodoroIPCServer) Start() {

	listenPath := "/tmp/pomodoro.sock"
	os.Remove(listenPath)

	l, err := net.Listen("unix", listenPath)
	if err != nil {
		fmt.Printf("Error al escuchar en el socket Unix: %v\n", err)
		return
	}
	defer l.Close()

	go s.pomodoro.Start(s.ctx)

	fmt.Println("Servidor IPC Pomodoro iniciado")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		s.cancel()
		l.Close()
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("Error al aceptar la conexión: %v\n", err)
			return
		}
		go s.handleConnection(conn)
	}
}

func (s *PomodoroIPCServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("Error al leer desde la conexión: %v\n", err)
		return
	}

	request := string(buf)
	switch request {
	case "status":
		status := s.pomodoro.GetStatus()
		conn.Write([]byte(status))
	case "stop":
		conn.Write([]byte("Deteniendo el servidor"))
		conn.Close()
		os.Exit(0)
	default:
		conn.Write([]byte("Comando no válido"))
	}
}