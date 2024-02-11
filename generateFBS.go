package main

import (
	"net/url"
	"os"
	"time"

	"github.com/amitbet/vncproxy/common"
	"github.com/amitbet/vncproxy/encodings"
	"github.com/amitbet/vncproxy/logger"
	"github.com/blueBye/vnc_recorder/client"
	"github.com/blueBye/vnc_recorder/recorder"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/remoteconsoles"
	"github.com/gorilla/websocket"
)

func generateFBS(recordingData RecordingData, name string, remoteConsole *remoteconsoles.RemoteConsole) {
	u := url.URL{
		Scheme:   "ws",
		Host:     os.Getenv("OS_WSOCKHOST"),
		Path:     "/",
		RawQuery: "token=" + remoteConsole.URL[len(remoteConsole.URL)-36:]}

	// Create a new websocket connection to the noVNC endpoint
	logger.Info("creating a websocker connection to the noVNC endpoint")
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logger.Warn("dial error:", err)
		return
	}
	defer conn.Close()

	// define a new Read/Write consumer (implements net.Dial interface with websocket)
	logger.Info("creating a read/write consumer")
	state := "recording"

	var noauth client.ClientAuthNone
	authArr := []client.ClientAuth{&client.PasswordAuth{Password: ""}, &noauth}
	wsc := client.RWC{WSC: conn,
		Buffer:      make(chan []byte, 1),
		InputStream: make(chan []byte, 100)}
	defer wsc.Close()

	// websocket message handler goroutine (incomming messages): writes messages to wsc buffer
	go func() {
		logger.Info("listener goroutine started")
		for {
			if state != "recording" {
				return
			}

			_, message, err := conn.ReadMessage()
			if err != nil {
				logger.Warn("wsc read error:", err)
				return
			}
			wsc.InputStream <- []byte(message)
		}
	}()

	// create a new fbs/rfb client (noVNC client)
	logger.Info("connecting to fbs/rfb client")
	clientConn, _ := client.NewClientConn(wsc,
		&client.ClientConfig{
			Auth:      authArr,
			Exclusive: true,
		})
	defer clientConn.Close()

	// create a new recorder to save the rfb traffic as a fbs file
	logger.Info("creating the recorder")

	rec, err := recorder.NewRecorder(name + ".fbs")
	if err != nil {
		logger.Warn("error creating recorder:", err)
		return
	}
	defer rec.Close()

	// add listeners to the clientConn for recorder (see vncproxy/recorder)
	clientConn.Listeners.AddListener(&recorder.RfbRequester{Conn: clientConn, Name: "Rfb Requester"})
	clientConn.Listeners.AddListener(rec)

	// attempt connecting to the noVNC server
	logger.Info("connecting to noVNC server")
	clientConn.Connect()
	if err != nil {
		logger.Warn("error while connecting to noVNC server:", err)
		return
	}

	// add encodings to the client after successful connection (see vncproxy/recorder)
	logger.Info("setting encodings")
	encs := []common.IEncoding{
		// &encodings.TightEncoding{},
		&encodings.ZLibEncoding{},
	}
	clientConn.SetEncodings(encs)

	// record for 10 seconds
	time.Sleep(time.Second * time.Duration(recordingData.Duration))

	logger.Info("closing connection")
}
