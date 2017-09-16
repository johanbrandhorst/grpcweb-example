package book

import (
	"context"
	"io"
	"time"

	"honnef.co/go/js/dom"
	r "myitcv.io/react"

	"github.com/johanbrandhorst/grpcweb-example/client/proto/library"
)

//go:generate reactGen
//go:generate immutableGen

var document = dom.GetWindow().Document()

const chatBoxID = "chat-box"

// BookChatDef defines the BookChat component
type BookChatDef struct {
	r.ComponentDef
}

// BookChatProps defines the properties of this component
type BookChatProps struct {
	Client library.BookServiceClient
}

// _Imm_Messages is generated to an immutable
// type *books which we use in the state
type _Imm_Messages []string

// BookChatState holds the state for the BookChat component
type BookChatState struct {
	messageInput string
	nameInput    string
	messages     *Messages
	client       library.BookService_BookChatClient
	err          string
	connTimeout  time.Duration
}

// BookChat returns a new BookChatElem
func BookChat(p BookChatProps) *BookChatElem {
	return buildBookChatElem(p)
}

// Render renders the BookChat component
func (g BookChatDef) Render() r.Element {
	st := g.State()
	content := []r.Element{
		r.P(nil, r.S("Discuss books with others")),
	}

	if st.client == nil {
		content = append(content,
			r.Form(&r.FormProps{ClassName: "form-inline"},
				r.Div(
					&r.DivProps{ClassName: "form-group"},
					r.Label(&r.LabelProps{
						ClassName: "sr-only",
						For:       "nameText",
					}, r.S("Name")),
					r.Input(&r.InputProps{
						Type:        "text",
						ClassName:   "form-control",
						ID:          "nameText",
						Value:       st.nameInput,
						OnChange:    nameInputChange{g},
						Placeholder: "Your Name",
					}),
					r.Button(&r.ButtonProps{
						Type:      "submit",
						ClassName: "btn btn-default",
						OnClick:   toggleconnect{g},
					}, r.S("Connect to chat")),
				),
			))
	}

	if st.client != nil {
		var msgs []r.Element
		for _, msg := range st.messages.Range() {
			msgs = append(msgs,
				r.Code(
					nil,
					r.S(msg),
				),
				r.Br(nil),
			)
		}
		content = append(content,
			r.Div(&r.DivProps{
				ClassName: "panel panel-default panel-body",
				Style: &r.CSS{
					MaxHeight: "150px",
					OverflowY: "auto",
					MinHeight: "150px",
				},
				ID: chatBoxID,
			}, msgs...),
			r.Hr(nil),
		)

		content = append(content,
			r.Form(&r.FormProps{ClassName: "form-inline"},
				r.Div(
					&r.DivProps{ClassName: "form-group"},
					r.Label(&r.LabelProps{
						ClassName: "sr-only",
						For:       "noteText",
					}, r.S("Message")),
					r.Input(&r.InputProps{
						Type:      "text",
						ClassName: "form-control",
						ID:        "noteText",
						Value:     st.messageInput,
						OnChange:  messageInputChange{g},
					}),
					r.Button(&r.ButtonProps{
						Type:      "submit",
						ClassName: "btn btn-default",
						OnClick:   send{g},
					}, r.S("Send")),
					r.Button(&r.ButtonProps{
						Type:      "submit",
						ClassName: "btn btn-default",
						OnClick:   toggleconnect{g},
					}, r.S("Leave Chat")),
				),
				r.Div(
					&r.DivProps{ClassName: "form-group"},
					r.Code(nil,
						r.S("Auto logout: "+st.connTimeout.String()),
					),
				),
			),
		)
	}

	if st.err != "" {
		content = append(content,
			r.Div(nil,
				r.Hr(nil),
				r.S("Error: "+st.err),
			),
		)
	}

	return r.Div(nil, content...)
}

type toggleconnect struct{ g BookChatDef }
type messageInputChange struct{ g BookChatDef }
type nameInputChange struct{ g BookChatDef }
type send struct{ g BookChatDef }

func (n messageInputChange) OnChange(se *r.SyntheticEvent) {
	target := se.Target().(*dom.HTMLInputElement)

	newSt := n.g.State()
	newSt.messageInput = target.Value
	n.g.SetState(newSt)
}

func (n nameInputChange) OnChange(se *r.SyntheticEvent) {
	target := se.Target().(*dom.HTMLInputElement)

	newSt := n.g.State()
	newSt.nameInput = target.Value
	n.g.SetState(newSt)
}

func (t toggleconnect) OnClick(se *r.SyntheticMouseEvent) {
	// Wrapped in goroutine because BookChat is blocking
	go func() {
		newSt := t.g.State()
		defer func() {
			t.g.SetState(newSt)
		}()
		newSt.err = ""
		newSt.messages = nil

		if newSt.client != nil {
			newSt.connTimeout = 0
			err := newSt.client.CloseSend()
			newSt.client = nil
			if err != nil {
				newSt.err = err.Error()
			}
			return
		}

		if newSt.nameInput == "" {
			newSt.err = "Name must not be empty"
			return
		}

		var err error
		timeout := 5 * time.Minute
		ctx, _ := context.WithTimeout(context.Background(), timeout)
		newSt.client, err = t.g.Props().Client.BookChat(ctx)
		if err != nil {
			newSt.err = err.Error()
			return
		}

		newSt.messages = NewMessages("Welcome to the BookChat, " + newSt.nameInput + "!")
		newSt.connTimeout = timeout
		// Start automatic disconnect countdown
		go func() {
			st := t.g.State()
			for st.connTimeout > 0 {
				time.Sleep(time.Second)
				st = t.g.State()
				st.connTimeout -= time.Second
				t.g.SetState(st)
			}
		}()

		// Start listener
		go func() {
			for {
				msg, err := newSt.client.Recv()
				if err == io.EOF {
					return
				}
				newSt := t.g.State()
				if err != nil {
					newSt.err = err.Error()
					newSt.client = nil
					t.g.SetState(newSt)
					return
				}

				// Must be done before updating state
				shouldScroll := scrollIsAtBottom()

				newSt.messages = newSt.messages.Append(msg.GetMessage())
				t.g.SetState(newSt)

				// Scroll to bottom of chatbox on new messages
				if shouldScroll {
					scrollToBottom()
				}
			}
		}()

		err = newSt.client.Send(&library.BookMessage{Content: &library.BookMessage_Name{Name: newSt.nameInput}})
		if err != nil {
			newSt.err = err.Error()
			newSt.client = nil
		}

		return
	}()

	se.PreventDefault()
}

func (s send) OnClick(se *r.SyntheticMouseEvent) {
	// Send is blocking
	go func() {
		newSt := s.g.State()
		defer func() {
			s.g.SetState(newSt)
		}()
		if newSt.messageInput == "" {
			return
		}

		err := newSt.client.Send(&library.BookMessage{Content: &library.BookMessage_Message{Message: newSt.messageInput}})
		if err != nil {
			newSt.err = err.Error()
		}
		newSt.messageInput = ""
	}()

	se.PreventDefault()
}

func scrollIsAtBottom() bool {
	node := document.GetElementByID(chatBoxID)
	if node != nil {
		div := node.(*dom.HTMLDivElement)
		boxHeight := div.Get("clientHeight").Int()
		scrollHeight := div.Get("scrollHeight").Int()
		scrollTop := div.Get("scrollTop").Int()
		// Scrolls down only if scroll is already close
		// to bottom.
		if scrollHeight-boxHeight < scrollTop+1 {
			return true
		}
	}

	return false
}

func scrollToBottom() {
	node := document.GetElementByID(chatBoxID)
	if node != nil {
		div := node.(*dom.HTMLDivElement)
		div.Set("scrollTop", div.Get("scrollHeight"))
	}
}
