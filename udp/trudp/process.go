package trudp

import "time"

// This module process all trudp internal events:
// - read (receive from udp),
// - write (receive from user level, need write to udp)
// - keep alive timer
// - resend packet from send queue timer

type process struct {
	chanWrite   chan []byte      // channel to write (used to send data from user level)
	chanRead    chan []byte      // channel to read (used to process packets received from udp)
	timerKeep   *time.Ticker     // keep alive timer
	timerResend <-chan time.Time // resend packet from send queue timer
}

// init
func (proc *process) init() *process {

	// Set time variables
	resendTime := defaultRTT * time.Millisecond
	pingTime := pingInterval * time.Millisecond
	//disconnectTime := disconnectAfter * time.Millisecond

	// Init channels and timers
	proc.chanRead = make(chan []byte)
	proc.chanWrite = make(chan []byte)
	//
	proc.timerResend = time.After(resendTime)
	proc.timerKeep = time.NewTicker(pingTime)

	// Do it on return
	defer func() { proc.timerKeep.Stop() }()

	// Module worker
	go func() {
		for {
			select {
			case <-proc.chanRead:
			case <-proc.chanWrite:
			case <-proc.timerKeep.C:
			case <-proc.timerResend:
				proc.timerResend = time.After(resendTime) // Set new timer value
			}
		}
	}()

	return proc
}

// destroy
func (proc *process) destroy() {

}
