package rfusb

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"go.bug.st/serial"
)

type Command struct {
	Set string `json:"set"`
	To  string `json:"to"`
}

type Report struct {
	Report string `json:"report"`
	Is     string `json:"is"`
}

type RFUSB struct {
	mu      *sync.Mutex
	sp      serial.Port
	port    string
	timeout time.Duration
}

type Mock struct {
	mu   *sync.Mutex
	port string
}

type Switch interface {
	Close() error
	Get() string
	Open(port string, baud int, timeout time.Duration) error
	SetPort(port string) error
	SetShort() error
	SetOpen() error
	SetLoad() error
	SetThru() error
	SetDUT1() error
	SetDUT2() error
	SetDUT3() error
	SetDUT4() error
}

func NewMock() *Mock {
	return &Mock{
		mu:   &sync.Mutex{},
		port: "unknown",
	}
}

func (m *Mock) Close() error {
	return nil
}

func (m *Mock) Get() string {
	return m.port
}

func (m *Mock) Open(port string, baud int, timeout time.Duration) error {
	return nil
}

func (m *Mock) SetPort(port string) error {
	m.port = port
	return nil
}

func (m *Mock) SetShort() error {
	return m.SetPort("short")
}

func (m *Mock) SetOpen() error {
	return m.SetPort("open")
}

func (m *Mock) SetLoad() error {
	return m.SetPort("load")
}

func (m *Mock) SetThru() error {
	return m.SetPort("thru")
}
func (m *Mock) SetDUT1() error {
	return m.SetPort("dut1")
}
func (m *Mock) SetDUT2() error {
	return m.SetPort("dut2")
}
func (m *Mock) SetDUT3() error {
	return m.SetPort("dut3")
}
func (m *Mock) SetDUT4() error {
	return m.SetPort("dut4")
}

func NewRFUSB() *RFUSB {
	return &RFUSB{
		mu:   &sync.Mutex{},
		port: "unknown",
		//don't initialise sp - use Open() for that
	}
}

func (r *RFUSB) Get() string {
	return r.port
}

func (r *RFUSB) Open(port string, baud int, timeout time.Duration) error {

	r.timeout = timeout

	mode := &serial.Mode{
		BaudRate: baud,
	}

	p, err := serial.Open(port, mode)

	if err != nil {
		log.WithFields(log.Fields{"port": port, "baud": baud, "timeout": timeout.String()}).Errorf("failed to open usb port")
		return err
	}

	r.sp = p

	err = r.sp.SetReadTimeout(timeout)

	if err != nil {
		log.WithFields(log.Fields{"port": port, "baud": baud, "timeout": timeout.String()}).Errorf("failed to set timeout when opening usb port")
		return err
	}

	log.WithFields(log.Fields{"port": port, "baud": baud, "timeout": timeout.String()}).Infof("opened usb port")

	return nil

}

func (r *RFUSB) Close() error {
	// don't take lock because there is read, close concurrency
	// https://github.com/bugst/go-serial/blob/e381f2c1332081ea593d73e97c71342026876857/serial_linux_test.go#L35
	return r.sp.Close()
}

func (r *RFUSB) SetShort() error {
	return r.SetPort("short")
}

func (r *RFUSB) SetOpen() error {
	return r.SetPort("open")
}

func (r *RFUSB) SetLoad() error {
	return r.SetPort("load")
}

func (r *RFUSB) SetThru() error {
	return r.SetPort("thru")
}
func (r *RFUSB) SetDUT1() error {
	return r.SetPort("dut1")
}
func (r *RFUSB) SetDUT2() error {
	return r.SetPort("dut2")
}
func (r *RFUSB) SetDUT3() error {
	return r.SetPort("dut3")
}
func (r *RFUSB) SetDUT4() error {
	return r.SetPort("dut4")
}

func (r *RFUSB) SetPort(port string) error {

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.sp == nil {
		return errors.New("port is nil")
	}

	resp := make([]byte, 128)

	// read any stale messages before we send our command
	// make a short timeout temporarily to avoid wasting time
	err := r.sp.SetReadTimeout(10 * time.Millisecond)
	if err != nil {
		return fmt.Errorf("setting short timeout before drain failed because %s", err.Error())
	}
DRAINED:
	for {

		n, err := r.sp.Read(resp)
		if err != nil {
			return err //port probably closed
		}
		//https://github.com/bugst/go-serial/blob/e381f2c1332081ea593d73e97c71342026876857/serial_unix.go#L94
		// timeout is n==0, err==nil
		if n == 0 {
			break DRAINED
		}
		continue
	}

	// restore normal timeout
	err = r.sp.SetReadTimeout(r.timeout)

	if err != nil {
		return fmt.Errorf("restoring timeout after drain failed because %s", err.Error())
	}

	request := Command{
		Set: "port",
		To:  port,
	}

	req, err := json.Marshal(request)

	if err != nil {
		return fmt.Errorf("marshal request failed because %s", err.Error())
	}

	n, err := r.sp.Write(req)

	log.WithFields(log.Fields{"count_expected": len(req), "count_actual": n, "data_expected": string(req), "data_actual": string(req[:n])}).Trace("wrote message to usb")

	if err != nil {
		return err
	}

	if n < len(req) {
		// TODO consider a follow up write?
		return errors.New("did not finish writing message")
	}

	// Get the response
	// note we do a drain afterwards to avoid this error:
	// unmarshalling reply failed because because unexpected end of JSON input. Reply was {"report":"port","is":"sho

	reply := make([]byte, 128)

	n, err = r.sp.Read(resp)

	if err != nil {
		return fmt.Errorf("reading reply failed because because %s", err.Error())
	}

	if n == 0 {
		return fmt.Errorf("empty reply")
	}

	idx := n - 1
	copy(reply[:], resp[:])

	//check we drained the whole message
	// make a short timeout temporarily to avoid wasting time if we got the whole message already
	err = r.sp.SetReadTimeout(100 * time.Millisecond) //don't make it too short or else get partial messages (that happens at 10ms)

	if err != nil {
		return fmt.Errorf("setting short timeout before drain failed because %s", err.Error())
	}
COMPLETED:
	for {

		n, err := r.sp.Read(resp)
		if err != nil {
			return err //port probably closed
		}
		//https://github.com/bugst/go-serial/blob/e381f2c1332081ea593d73e97c71342026876857/serial_unix.go#L94
		// timeout is n==0, err==nil
		if n == 0 {
			break COMPLETED
		}
		if (idx + n) < len(reply) {
			copy(reply[idx+1:idx+n], resp[:]) //TODO check if copies null?
			idx = idx + n

		} else {
			log.Fatal("pkg/rfusb: serial read buffer full")
		}
		continue
	}

	var report Report
	log.Debugf("(%d)%s", idx, string(reply[:idx]))
	err = json.Unmarshal(reply[:idx], &report) //truncate to bytes read to avoid \x00 char which breaks unmarshal

	if err != nil {
		return fmt.Errorf("unmarshalling reply failed because because %s. Reply was %s", err.Error(), string(resp))
	}
	log.WithFields(log.Fields{"count_actual": n, "data_actual": string(resp[:n])}).Trace("read message from usb")
	if strings.ToLower(report.Report) != "port" {
		return errors.New("response was not a port report")
	}
	if strings.ToLower(report.Is) != strings.ToLower(port) {
		return err
	}
	r.port = port
	return nil

}
