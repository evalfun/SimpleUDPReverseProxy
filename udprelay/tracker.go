package udprelay

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	TK_STAT_INIT       = 0
	TK_STAT_CONNECTING = 1
	TK_STAT_CONNECTED  = 2
	TK_STAT_ERROR      = 3
	TK_STAT_CLOSED     = 9
)

const (
	TK_MSG_TYPE_UPDATEIP        = "UPDATEIP"
	TK_MSG_TYPE_OFFLINE         = "OFFLINE"
	TK_MSG_TYPE_CONNECT         = "CONNECT"
	TK_MSG_TYPE_CONNECT_SUCCESS = "CONNECT_SUCCESS"
)

type TrackerConfig struct {
	ServerID  string
	UserID    string
	ServerURL string
}

func (this *TrackerConfig) GetServerAddrList() ([]string, error) {
	original_url, err := url.Parse(this.ServerURL)
	if err != nil {
		return nil, err
	}
	query := original_url.Query()
	query.Add("server_id", this.ServerID)
	query.Add("user_id", this.UserID)
	new_url := &url.URL{
		Scheme:      original_url.Scheme,
		Opaque:      original_url.Opaque,
		User:        original_url.User,
		Host:        original_url.Host,
		Path:        original_url.Path + "/client",
		RawPath:     original_url.RawPath,
		RawQuery:    query.Encode(),
		Fragment:    original_url.Fragment,
		RawFragment: original_url.RawFragment,
	}
	res, err := http.Get(new_url.String())
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, errors.New(string(data))
	}
	var addrList []string
	err = json.Unmarshal(data, &addrList)
	if err != nil {
		return nil, err
	}
	return addrList, nil
}

func (this *TrackerConfig) ReqServerConnect(addrList []string) error {
	type Request struct {
		ServerID string
		AddrList []string
		UserID   string
	}
	req := Request{
		ServerID: this.ServerID,
		AddrList: addrList,
		UserID:   this.UserID,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	body := strings.NewReader(string(data))
	original_url, err := url.Parse(this.ServerURL)
	if err != nil {
		return err
	}
	new_url := &url.URL{
		Scheme:      original_url.Scheme,
		Opaque:      original_url.Opaque,
		User:        original_url.User,
		Host:        original_url.Host,
		Path:        original_url.Path + "/client",
		RawPath:     original_url.RawPath,
		RawQuery:    original_url.RawQuery,
		Fragment:    original_url.Fragment,
		RawFragment: original_url.RawFragment,
	}
	resp, err := http.Post(new_url.String(), "application/json", body)
	if err != nil {
		return err
	}
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New(string(data))
	}
	return nil
}

type TrackerMessage struct {
	MessageType string
	AddrList    []string
	Message     string
}

type ServerTracker struct {
	serverID    string
	userID      string
	url         string
	message     string
	stat        int
	messgaeChan chan *TrackerMessage
	connection  *websocket.Conn
}

func NewServerTracker(config *TrackerConfig, messageChan chan *TrackerMessage) *ServerTracker {
	this := &ServerTracker{
		serverID:    config.ServerID,
		userID:      config.UserID,
		url:         config.ServerURL,
		messgaeChan: messageChan,
	}
	go this.connectProc()
	return this
}

func (this *ServerTracker) GetStat() (string, int) {
	return this.message, this.stat
}

func (this *ServerTracker) SendAddrList(addrList []string) {
	if this.stat != TK_STAT_CONNECTED {
		return
	}
	data, err := json.Marshal(TrackerMessage{
		MessageType: TK_MSG_TYPE_UPDATEIP,
		AddrList:    addrList,
	})
	if err != nil {
		log.Println("Tracker Marshal Server addr list fail: ", err.Error())
	}
	err = this.connection.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Println("Tracker send Server addr list fail: ", err.Error())
	}
}

func (this *ServerTracker) Close() {
	if this.stat == TK_STAT_CLOSED {
		return
	}
	this.stat = TK_STAT_CLOSED
	if this.connection != nil {
		this.connection.Close()
	}
}

func (this *ServerTracker) connectProc() {
	var err error
	this.stat = TK_STAT_CONNECTING
	original_url, err := url.Parse(this.url)
	if err != nil {
		this.stat = TK_STAT_ERROR
		this.message = err.Error()
		return
	}
	query := original_url.Query()
	query.Add("server_id", this.serverID)
	query.Add("user_id", this.userID)
	new_url := &url.URL{
		Scheme:      original_url.Scheme,
		Opaque:      original_url.Opaque,
		User:        original_url.User,
		Host:        original_url.Host,
		Path:        original_url.Path + "/server",
		RawPath:     original_url.RawPath,
		RawQuery:    query.Encode(),
		Fragment:    original_url.Fragment,
		RawFragment: original_url.RawFragment,
	}
	tracker_url := new_url.String()
	for this.stat != TK_STAT_CLOSED {
		log.Printf("Tracker connect to %s", tracker_url)
		this.connection, _, err = websocket.DefaultDialer.Dial(tracker_url, nil)
		if this.stat == TK_STAT_CLOSED {
			return
		}
		if err != nil {
			log.Printf("Tracker connect to %s fail: %s", this.url, err.Error())
			this.message = err.Error()
		} else {
			this.stat = TK_STAT_CONNECTED
			this.message = "Connect success."
			log.Printf("Tracker connect to %s success", this.url)
			this.messgaeChan <- &TrackerMessage{
				MessageType: TK_MSG_TYPE_CONNECT_SUCCESS,
			}
		}
		for this.stat == TK_STAT_CONNECTED {
			_, message, err := this.connection.ReadMessage()
			if this.stat == TK_STAT_CLOSED {
				return
			}
			if err != nil {
				log.Printf("Tracker read from %s fail: %s", this.url, err.Error())
				this.stat = TK_STAT_CONNECTING
				this.message = err.Error()
				break
			}
			var tkMessage TrackerMessage
			err = json.Unmarshal(message, &tkMessage)
			if err == nil {
				this.messgaeChan <- &tkMessage
			}
		}
		time.Sleep(2 * time.Second)
	}
}
