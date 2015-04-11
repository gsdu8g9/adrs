package wiseman

import (
	"github.com/chzyer/adrs/customer"
	"github.com/chzyer/adrs/dns"
	"github.com/chzyer/adrs/mailman"
	"github.com/chzyer/adrs/utils"
	"gopkg.in/logex.v1"
)

type WiseMan struct {
	frontDoor   customer.Corridor
	backDoor    customer.Corridor
	mailMan     *mailman.MailMan
	incomingBox chan *mailman.Envelope
	outgoingBox chan *mailman.Envelope
}

func NewWiseMan(frontDoor, backDoor customer.Corridor, incomingBox, outgoingBox chan *mailman.Envelope) (*WiseMan, error) {
	w := &WiseMan{
		frontDoor:   frontDoor,
		backDoor:    backDoor,
		incomingBox: incomingBox,
		outgoingBox: outgoingBox,
	}
	return w, nil
}

func (w *WiseMan) ServeAll() {
	var customer *customer.Customer
	var envelope *mailman.Envelope
	var err error
	for {
		select {
		case envelope = <-w.incomingBox:
			// new mail is receive
			err = w.receiveMail(envelope)
			if err != nil {
				logex.Error(err)
				envelope.Customer.LetItGo()
				continue
			}
			// say goodbye
			w.backDoor <- envelope.Customer
		case customer = <-w.frontDoor:
			// new customer is comming
			err = w.serve(customer)
			if err != nil {
				// oops!, the wise man is passed out!
				logex.Error(err)
				customer.LetItGo()
				continue
			}
		}
	}
}

func (w *WiseMan) serve(c *customer.Customer) error {
	r := utils.NewRecordReader(c.Raw)
	msg, err := dns.NewDNSMessage(r)
	if err != nil {
		return logex.Trace(err)
	}

	logex.Info("here comes a customer")
	// looking up the wikis.
	// ask others

	// we don't know where to send yet
	mail := w.writeMail(c, msg)

	logex.Info("sending a mail to others")

	// but don't worry, my mail man know
	w.sendMail(mail, c)
	return nil
}

func (w *WiseMan) sendMail(mail *mailman.Mail, c *customer.Customer) {
	w.outgoingBox <- &mailman.Envelope{mail, c}
}

func (w *WiseMan) receiveMail(e *mailman.Envelope) error {
	if e.Mail.Reply == nil {
		return logex.NewError("oops!")
	}

	logex.Info("we got a answer from remote")

	// write to wiki in case someone ask the same question
	e.Customer.SetRaw(e.Mail.Reply)
	msg, err := w.readMailAndDestory(e.Mail)
	if err != nil {
		return logex.Trace(err)
	}
	e.Customer.Msg = msg
	return nil
}

func (w *WiseMan) writeMail(c *customer.Customer, msg *dns.DNSMessage) *mailman.Mail {
	return &mailman.Mail{
		From:    c.Session,
		Content: msg,
	}
}

func (w *WiseMan) readMailAndDestory(m *mailman.Mail) (*dns.DNSMessage, error) {
	msg, err := dns.NewDNSMessage(utils.NewRecordReader(m.Reply))
	if err != nil {
		return nil, logex.Trace(err)
	}
	return msg, nil
}
